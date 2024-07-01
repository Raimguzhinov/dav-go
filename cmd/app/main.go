package main

import (
	"github.com/Raimguzhinov/dav-go/internal/app"
	"github.com/Raimguzhinov/dav-go/internal/config"
)

func main() {
	cfg := config.GetConfig()

	app.Run(cfg)
}
