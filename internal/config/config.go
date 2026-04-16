package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// NavItem represents a navigation link.
type NavItem struct {
	Text        string `json:"text" yaml:"text"`
	Link        string `json:"link" yaml:"link"`
	ActiveMatch string `json:"activeMatch,omitempty" yaml:"activeMatch,omitempty"`
}

// SidebarItem represents a sidebar group or link.
type SidebarItem struct {
	Text string `json:"text" yaml:"text"`
	Link string `json:"link,omitempty" yaml:"link,omitempty"`
	// 规则：
	//   1. 不指定该字段 或者 子项Items为空 → 分组**不可折叠**
	//   2. 设置为 true  → 分组**可折叠**，且**默认折叠**
	//   3. 设置为 false → 分组**可折叠**，且**默认展开**
	Collapsed *bool         `json:"collapsed,omitempty" yaml:"collapsed,omitempty"`
	Items     []SidebarItem `json:"items,omitempty" yaml:"items,omitempty"`
}

// IsCollapsible returns true if the sidebar item is collapsible.
func (s SidebarItem) IsCollapsible() bool {
	return s.Collapsed != nil && len(s.Items) > 0
}

// IsCollapsed returns true if the sidebar item should be collapsed by default.
func (s SidebarItem) IsCollapsed() bool {
	return s.IsCollapsible() && *s.Collapsed
}

// SocialLink represents a social icon link (e.g. GitHub).
type SocialLink struct {
	Icon string `json:"icon" yaml:"icon"`
	Link string `json:"link" yaml:"link"`
}

type ThemeableImage struct {
	Light string `json:"light,omitempty" yaml:"light,omitempty"`
	Dark  string `json:"dark,omitempty" yaml:"dark,omitempty"`
	Alt   string `json:"alt,omitempty" yaml:"alt,omitempty"`
}

// ThemeConfig represents the theme configuration options.
type ThemeConfig struct {
	Logo        ThemeableImage           `json:"logo,omitempty" yaml:"logo,omitempty"`
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
			Logo: ThemeableImage{
				Light: "https://lf-flow-web-cdn.doubao.com/obj/flow-doubao/doubao/web/doubao_avatar.png",
				Dark:  "https://lf-flow-web-cdn.doubao.com/obj/flow-doubao/doubao/web/doubao_avatar.png",
				Alt:   "Logo",
			},
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

		jsonData, _ := json.Marshal(config)
		slog.Warn("Loaded config.json", "path", jsonPath, "data", string(jsonData))

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
