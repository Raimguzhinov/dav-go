package main

import (
	"log"

	"github.com/Raimguhinov/dav-go/config"
	"github.com/Raimguhinov/dav-go/internal/app"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.Run(cfg)
}
