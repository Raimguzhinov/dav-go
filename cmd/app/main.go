package main

import (
	"log"

	"github.com/Raimguhinov/dav-go/configs"
	"github.com/Raimguhinov/dav-go/internal/app"
)

func main() {
	cfg, err := configs.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.Run(cfg)
}
