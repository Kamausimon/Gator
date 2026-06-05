package config

import (
	"fmt"
	"os"
)

type State struct {
	Config *Config
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
	fmt.Printf("username has been set successfully to %s\n", username)
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
