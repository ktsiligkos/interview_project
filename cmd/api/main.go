package main

import (
	"log"

	"github.com/ktsiligkos/xm_project/internal/platform/app"
	"github.com/ktsiligkos/xm_project/pkg/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	defer func() {
		if err := application.Close(); err != nil {
			log.Printf("error while closing application: %v", err)
		}
	}()

	if err := application.Run(); err != nil {
		log.Fatalf("application stopped with error: %v", err)
	}
}
