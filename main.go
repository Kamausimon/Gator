package main

import (
	"fmt"
	"os"

	"github.com/Kamausimon/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("there was an error reading the config file %v", err)
		return
	}
	appState := &config.State{Config: cfg}
	cmds := &config.Commands{
		Handlers: make(map[string]func(*config.State, config.Command) error)}
	cmds.Register("login", config.HandlerLogin)
	if len(os.Args) < 2 {
		fmt.Println("Please provide a command")
		return
	}
	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]
	cmd := config.Command{Name: cmdName, Arguments: cmdArgs}

	if err := cmds.Run(appState, cmd); err != nil {
		fmt.Printf("there was an error running the command: %v\n", err)
		return
	}

	fmt.Printf("Config updated: %+v\n", cfg)
	freshCfg, err := config.Read()
	if err != nil {
		fmt.Printf("there was an error reading the config file after update: %v", err)
		return
	}
	fmt.Printf("Config read after update: %+v\n", freshCfg)
}
