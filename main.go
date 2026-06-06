package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Kamausimon/gator/internal/config"
	"github.com/Kamausimon/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {

	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("there was an error reading the config file %v", err)
		return
	}
	db, err := sql.Open("postgres", os.Getenv("DBURL"))
	if err != nil {
		fmt.Printf("there was an error connecting to the database: %v\n", err)
		return
	}
	defer db.Close()
	dbQueries := database.New(db)
	appState := &config.State{Config: cfg, Db: dbQueries}
	cmds := &config.Commands{
		Handlers: make(map[string]func(*config.State, config.Command) error)}
	cmds.Register("login", config.HandlerLogin)
	cmds.Register("register", config.HandlerRegister)
	cmds.Register("reset", config.HandlerReset)
	cmds.Register("users", config.Handlerusers)
	if len(os.Args) < 2 {
		fmt.Println("Please provide a command")
		os.Exit(1)
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
