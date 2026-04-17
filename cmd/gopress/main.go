package main

import (
	"fmt"
	"os"

	"github.com/icetech233/cobra"
	"github.com/icetech233/gopress/internal/build"
	"github.com/icetech233/gopress/internal/config"
	"github.com/icetech233/gopress/internal/server"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "gopress",
	Short: "gopress - a fast static site generator in Go",
	Long:  "gopress, delivering fast static site generation for Markdown documents.",
}

var devCmd = &cobra.Command{
	Use:   "dev [root]",
	Short: "Start the development server",
	Run: func(cmd *cobra.Command, args []string) {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}

		port, _ := cmd.Flags().GetInt("port")

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
	},
}

var buildCmd = &cobra.Command{
	Use:   "build [root]",
	Short: "Build the static site for production",
	Run: func(cmd *cobra.Command, args []string) {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}

		outDir, _ := cmd.Flags().GetString("outDir")

		siteConfig, err := config.LoadConfig(root)
		if err != nil {
			fmt.Printf("Warning: Could not load config: %v\n", err)
			siteConfig = config.DefaultSiteConfig()
		}

		builder := &build.Builder{
			Root:       root,
			OutDir:     outDir,
			SiteConfig: siteConfig,
		}

		if err := builder.Build(); err != nil {
			fmt.Printf("Error during build: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	devCmd.Flags().IntP("port", "p", 5173, "Port to listen on")
	buildCmd.Flags().StringP("outDir", "o", "", "Output directory (default .gopress/dist)")

	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(buildCmd)
}
