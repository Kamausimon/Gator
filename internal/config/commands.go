package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Kamausimon/gator/internal/database"
	"github.com/Kamausimon/gator/internal/rss"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type State struct {
	Config *Config
	Db     *database.Queries
}

type Command struct {
	Name      string
	Arguments []string
}

type Commands struct {
	Handlers map[string]func(*State, Command) error
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Arguments) < 1 {
		if err := fmt.Errorf("username argument is required for login command"); err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
	}
	username := cmd.Arguments[0]
	if err := s.Config.SetUser(username); err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}
	_, err := s.Db.GetUserByName(context.Background(), username)
	if err != nil {
		fmt.Printf("user with name %s does not exist, please register first\n", username)
		os.Exit(1)
	}
	fmt.Printf("username has been set successfully to %s\n", username)
	return nil
}
func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Arguments) < 1 {
		if err := fmt.Errorf("name argument is required for register command"); err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
	}
	//create a new user in the database
	name := cmd.Arguments[0]
	ctx := context.Background()
	_, err := s.Db.GetUserByName(ctx, name)
	if err == nil {
		fmt.Printf("user with name %s already exists", name)
		os.Exit(1)
	}
	_, err = s.Db.CreateUser(ctx,
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
			Name:      name,
		})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	//set the current user in the config to the given name
	if err := s.Config.SetUser(name); err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}
	fmt.Printf("user created successfully: %s\n", name)
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	ctx := context.Background()
	if err := s.Db.DeleteAllUsers(ctx); err != nil {
		return fmt.Errorf("failed to delete all users: %w", err)
	}
	fmt.Println("all users have been deleted")
	return nil
}

func Handlerusers(s *State, cmd Command) error {
	ctx := context.Background()
	users, err := s.Db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}
	for _, user := range users {
		if user.Name == s.Config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("  %s\n", user.Name)
		}
	}
	return nil
}

func HandlerAgg(s *State, cmd Command) error {
	if len(cmd.Arguments) < 1 {
		return fmt.Errorf("time_between_reqs argument is required")
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.Arguments[0])
	if err != nil {
		return fmt.Errorf("invalid time_between_reqs duration: %w", err)
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func scrapeFeeds(s *State) {
	ctx := context.Background()

	feed, err := s.Db.GetNextFeedToFetch(ctx)
	if err != nil {
		fmt.Printf("error getting next feed to fetch: %v\n", err)
		return
	}

	if _, err := s.Db.MarkFeedFetched(ctx, feed.ID); err != nil {
		fmt.Printf("error marking feed fetched: %v\n", err)
		return
	}

	rssFeed, err := rss.FetchFeed(ctx, feed.Url.String)
	if err != nil {
		fmt.Printf("error fetching feed %s: %v\n", feed.Url.String, err)
		return
	}

	for _, item := range rssFeed.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := parsePubDate(item.PubDate); err == nil {
			publishedAt = sql.NullTime{Time: t, Valid: true}
		}

		_, err := s.Db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt:   sql.NullTime{Time: time.Now(), Valid: true},
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				continue
			}
			fmt.Printf("error saving post %q: %v\n", item.Title, err)
			continue
		}
		fmt.Printf("Saved post: %s\n", item.Title)
	}
}

var pubDateLayouts = []string{
	time.RFC1123Z,
	time.RFC1123,
	time.RFC3339,
	"2006-01-02T15:04:05Z07:00",
	"Mon, 2 Jan 2006 15:04:05 -0700",
}

func parsePubDate(pubDate string) (time.Time, error) {
	pubDate = strings.TrimSpace(pubDate)
	var lastErr error
	for _, layout := range pubDateLayouts {
		t, err := time.Parse(layout, pubDate)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

func HandlerAddFeed(s *State, cmd Command) error {
	currentUser := s.Config.CurrentUserName
	if currentUser == "" {
		return fmt.Errorf("You must be logged in to access command please login first")
	}

	if len(cmd.Arguments) < 2 {
		if err := fmt.Errorf("name and feed arguments are required"); err != nil {
			fmt.Printf("error: %s", err)
			os.Exit(1)
		}
	}
	feedName := cmd.Arguments[0]
	feedUrl := cmd.Arguments[1]

	ctx := context.Background()
	user, err := s.Db.GetUserByName(ctx, currentUser)
	if err != nil {
		return fmt.Errorf("There was an error retrieveing user from the database")
	}
	newFeed, err := s.Db.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Name:      feedName,
		UserID:    user.ID,
		Url:       sql.NullString{String: feedUrl, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("there was an error adding the newsfeed to the db, %s", err)
	}

	_, err = s.Db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    newFeed.ID,
	})
	if err != nil {
		return fmt.Errorf("feed created successfully, but auto-follow registration failed: %w", err)
	}
	fmt.Println("Feed successfully registered!")
	fmt.Printf("ID:         %s\n", newFeed.ID)
	fmt.Printf("Name:       %s\n", newFeed.Name)
	fmt.Printf("URL:        %v\n", newFeed.Url)
	fmt.Printf("Owner ID:   %s\n", newFeed.UserID)

	return nil
}

func HandlerFeed(s *State, cmd Command) error {
	ctx := context.Background()

	feeds, err := s.Db.ListFeedsWithUser(ctx)
	if err != nil {
		return fmt.Errorf("there was an error retrieving the list")
	}

	if len(feeds) == 0 {
		fmt.Println("no feeds registred in the database")
		return nil
	}
	fmt.Println("=== Registered Feeds ===")
	for _, feed := range feeds {
		fmt.Printf("* Name:       %s\n", feed.FeedName)
		fmt.Printf("  URL:        %v\n", feed.Url)
		fmt.Printf("  Created By: %s\n", feed.UserName)
		fmt.Println("  --------------------")
	}
	return nil
}

func HandlerFollow(s *State, cmd Command) error {
	if len(cmd.Arguments) < 1 {
		if err := fmt.Errorf("url is needed as an argument"); err != nil {
			fmt.Printf("error, %s", err)
			os.Exit(1)
		}
	}
	feedUrl := cmd.Arguments[0]
	currentUserName := s.Config.CurrentUserName

	ctx := context.Background()
	//get the user
	user, err := s.Db.GetUserByName(ctx, currentUserName)
	if err != nil {
		return fmt.Errorf("THere was an error retrieveing the user from the db")
	}

	//get the feed from the url
	feed, err := s.Db.GetFeedByUrl(ctx, sql.NullString{String: feedUrl, Valid: true})
	if err != nil {
		return fmt.Errorf("There was an error getting the feed from the db %s", err)
	}

	//create the feedfollow for the user
	feedfollow, err := s.Db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("There was an error adding the feed to the db")
	}
	fmt.Println("successfully created the feeedflow")
	fmt.Printf("ID:       %s\n", feedfollow.ID)
	fmt.Printf("User:     %s\n", feedfollow.UserName)
	fmt.Printf("feed     %s\n", feedfollow.FeedName)

	return nil
}

func HandlerFollowing(s *State, cmd Command) error {
	currentUser := s.Config.CurrentUserName

	ctx := context.Background()

	userfeed, err := s.Db.GetFeedFollowsForUser(ctx, currentUser)
	if err != nil {
		fmt.Printf("There was an error getting the user feed follows: %s", err)
		return nil
	}

	if len(userfeed) == 0 {
		fmt.Println("no feed follows for this user")
		return nil
	}

	for _, follow := range userfeed {
		fmt.Printf("%s\n", follow.FeedName)
	}

	return nil

}

func HandlerUnfollow(s *State, cmd Command) error {
	if len(cmd.Arguments) < 1 {
		return fmt.Errorf("url argument is required")
	}
	currentUser := s.Config.CurrentUserName
	ctx := context.Background()

	user, err := s.Db.GetUserByName(ctx, currentUser)
	if err != nil {
		return fmt.Errorf("there was an error retrieving the user from the db")
	}

	feed, err := s.Db.GetFeedByUrl(ctx, sql.NullString{String: cmd.Arguments[0], Valid: true})
	if err != nil {
		return fmt.Errorf("there was an error retrieving the feed from the db")
	}

	err = s.Db.DeleteFeedFollowForUser(ctx, database.DeleteFeedFollowForUserParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("there was an error deleting the feed follow: %s", err)
	}
	fmt.Println("successfully unfollowed the feed")
	return nil
}

func HandlerBrowse(s *State, cmd Command) error {
	limit := int32(2)
	if len(cmd.Arguments) > 0 {
		parsed, err := strconv.Atoi(cmd.Arguments[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
		limit = int32(parsed)
	}

	ctx := context.Background()
	user, err := s.Db.GetUserByName(ctx, s.Config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("there was an error retrieving the user from the db")
	}

	posts, err := s.Db.GetPostsForUser(ctx, database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		return fmt.Errorf("there was an error retrieving posts: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("no posts found")
		return nil
	}

	for _, post := range posts {
		fmt.Printf("* %s\n", post.Title)
		fmt.Printf("  URL:  %s\n", post.Url)
		if post.PublishedAt.Valid {
			fmt.Printf("  Published: %s\n", post.PublishedAt.Time.Format(time.RFC1123))
		}
		if post.Description.Valid {
			fmt.Printf("  %s\n", post.Description.String)
		}
		fmt.Println("  --------------------")
	}
	return nil
}

// if it does not exist
func (c *Commands) Register(name string, handler func(*State, Command) error) {
	if c.Handlers == nil {
		c.Handlers = make(map[string]func(*State, Command) error)
	}
	c.Handlers[name] = handler
}

// if it exists
func (c *Commands) Run(s *State, cmd Command) error {
	handler, exists := c.Handlers[cmd.Name]
	if !exists {
		return fmt.Errorf("command not found: %s", cmd.Name)
	}
	return handler(s, cmd)
}
