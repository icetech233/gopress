package main

import (
	"fmt"
	"os"

	"github.com/icetech233/gopress/internal/config"
	"github.com/icetech233/gopress/internal/server"
)

func main() {
	root := "docs"
	port := 5173

	siteConfig, err := config.LoadConfig(root)
	if err != nil {
		fmt.Printf("Warning: Could not load config: %v\n", err)
		siteConfig = config.DefaultSiteConfig()
	}
	devServer := &server.DevServer{
		Root:       root,
		SiteConfig: siteConfig,
	}
	if err := devServer.Start(port); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
