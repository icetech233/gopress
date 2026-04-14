package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// NavItem represents a navigation link.
type NavItem struct {
	Text   string `json:"text" yaml:"text"`
	Link   string `json:"link" yaml:"link"`
	Active string `json:"activeMatch,omitempty" yaml:"activeMatch,omitempty"`
}

// SidebarItem represents a sidebar group or link.
type SidebarItem struct {
	Text      string        `json:"text" yaml:"text"`
	Link      string        `json:"link,omitempty" yaml:"link,omitempty"`
	Collapsed *bool         `json:"collapsed,omitempty" yaml:"collapsed,omitempty"`
	Items     []SidebarItem `json:"items,omitempty" yaml:"items,omitempty"`
}

// SocialLink represents a social icon link (e.g. GitHub).
type SocialLink struct {
	Icon string `json:"icon" yaml:"icon"`
	Link string `json:"link" yaml:"link"`
}

// ThemeConfig represents the theme configuration options.
type ThemeConfig struct {
	Logo        string                   `json:"logo,omitempty" yaml:"logo,omitempty"`
	Nav         []NavItem                `json:"nav,omitempty" yaml:"nav,omitempty"`
	Sidebar     map[string][]SidebarItem `json:"sidebar,omitempty" yaml:"sidebar,omitempty"`
	SocialLinks []SocialLink             `json:"socialLinks,omitempty" yaml:"socialLinks,omitempty"`
	Footer      map[string]string        `json:"footer,omitempty" yaml:"footer,omitempty"`
}

// SiteConfig represents the root configuration of GoPress.
type SiteConfig struct {
	Title       string      `json:"title" yaml:"title"`
	Description string      `json:"description" yaml:"description"`
	Base        string      `json:"base" yaml:"base"`
	Lang        string      `json:"lang" yaml:"lang"`
	ThemeConfig ThemeConfig `json:"themeConfig" yaml:"themeConfig"`
}

// DefaultSiteConfig returns the default configuration.
func DefaultSiteConfig() *SiteConfig {
	return &SiteConfig{
		Title:       "gopress",
		Description: "A gopress website",
		Base:        "/",
		Lang:        "en-US",
		ThemeConfig: ThemeConfig{
			Nav:         []NavItem{},
			Sidebar:     make(map[string][]SidebarItem),
			SocialLinks: []SocialLink{},
			Footer:      make(map[string]string),
		},
	}
}

// LoadConfig loads configuration from .gopress/config.json or config.yaml.
func LoadConfig(root string) (*SiteConfig, error) {
	config := DefaultSiteConfig()
	configDir := filepath.Join(root, ".gopress")

	// Try JSON
	jsonPath := filepath.Join(configDir, "config.json")
	if data, err := os.ReadFile(jsonPath); err == nil {
		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("error parsing config.json: %w", err)
		}
		return config, nil
	}

	// Try YAML
	yamlPath := filepath.Join(configDir, "config.yaml")
	if data, err := os.ReadFile(yamlPath); err == nil {
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("error parsing config.yaml: %w", err)
		}
		return config, nil
	}

	// If no config is found, just return defaults. We don't support .ts out of the box in Go cleanly without a JS runtime.
	return config, nil
}
