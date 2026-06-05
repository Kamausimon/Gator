package main

import (
	"fmt"

	"github.com/Kamausimon/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("there was an error reading the config file %v", err)
		return
	}
	fmt.Printf("Initial config read: %+v\n", cfg)
	targetName := "Kamausimon"
	if err := cfg.SetUser(targetName); err != nil {
		fmt.Printf("there was an error setting the user name in the config file: %v", err)
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
