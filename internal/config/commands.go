package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/Kamausimon/gator/internal/database"
	"github.com/Kamausimon/gator/internal/rss"
	"github.com/google/uuid"
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
	targetUrl := "https://www.wagslane.dev/index.xml"
	feed, err := rss.FetchFeed(context.Background(), targetUrl)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}

	// Print out some individual items to confirm inner structural parsing and HTML unescaping works

	for _, item := range feed.Channel.Item {
		fmt.Printf("%s\n", item.Title)
		fmt.Printf("Description: %s\n", item.Description)

	}
	return nil
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
	fmt.Println("Feed successfully registered!")
	fmt.Printf("ID:         %s\n", newFeed.ID)
	fmt.Printf("Name:       %s\n", newFeed.Name)
	fmt.Printf("URL:        %v\n", newFeed.Url)
	fmt.Printf("Owner ID:   %s\n", newFeed.UserID)

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
