package main

import (
	"github.com/Raimguhinov/dav-go/internal/app"
	"github.com/Raimguhinov/dav-go/internal/config"
)

func main() {
	cfg := config.GetConfig()

	app.Run(cfg)
}
