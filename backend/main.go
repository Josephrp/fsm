package main

import (
	"flag"
	"log"
	"os"

	"github.com/snarf-dev/fsm/v2/internal/config"
	"github.com/snarf-dev/fsm/v2/internal/server"
)

func main() {
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	err, cfg := config.Load(configPath)
	if err != nil {
		log.Println("unable to use config, exiting")
		os.Exit(1)
	}

	server := server.CreateRestServer(cfg)
	server.Start()
}
