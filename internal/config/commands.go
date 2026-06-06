package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/Kamausimon/gator/internal/database"
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
