package main

import (
	"fmt"
	"os"
	"path/filepath"

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

var previewCmd = &cobra.Command{
	Use:   "preview [root]",
	Short: "Preview the built static site",
	Long:  "Starts a local web server to preview the built static site. Supports hot-reload.",
	Run: func(cmd *cobra.Command, args []string) {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}

		outDir, _ := cmd.Flags().GetString("outDir")
		if outDir == "" {
			outDir = filepath.Join(root, ".gopress", "dist")
		}

		port, _ := cmd.Flags().GetInt("port")

		// Check if outDir exists
		if stat, err := os.Stat(outDir); os.IsNotExist(err) || !stat.IsDir() {
			fmt.Printf("Error: Output directory '%s' does not exist. Please run 'gopress build' first.\n", outDir)
			os.Exit(1)
		}

		previewServer := &server.PreviewServer{
			OutDir: outDir,
		}

		if err := previewServer.Start(port); err != nil {
			fmt.Printf("Error starting preview server: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	devCmd.Flags().IntP("port", "p", 5173, "Port to listen on")
	buildCmd.Flags().StringP("outDir", "o", "", "Output directory (default .gopress/dist)")
	previewCmd.Flags().IntP("port", "p", 3000, "Port to listen on for preview")
	previewCmd.Flags().StringP("outDir", "o", "", "Output directory to preview (default .gopress/dist)")

	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(previewCmd)
}
