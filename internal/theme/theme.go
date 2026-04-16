package theme

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"log/slog"
	"strings"
	"text/template"

	"github.com/icetech233/gopress/internal/config"
	"github.com/icetech233/gopress/internal/markdown"
)

//go:embed theme.css
var themeCSS string

//go:embed theme.js
var themeJS string

//go:embed layout.tmpl
var layoutHTML string

// HeroAction represents a call-to-action button in the hero section.
type HeroAction struct {
	Theme string `json:"theme" yaml:"theme"`
	Text  string `json:"text" yaml:"text"`
	Link  string `json:"link" yaml:"link"`
}

// Hero represents the hero section on the home page.
type Hero struct {
	Name    string       `json:"name" yaml:"name"`
	Text    string       `json:"text" yaml:"text"`
	Tagline string       `json:"tagline" yaml:"tagline"`
	Actions []HeroAction `json:"actions" yaml:"actions"`
}

// Feature represents a feature block on the home page.
type Feature struct {
	Title   string `json:"title" yaml:"title"`
	Details string `json:"details" yaml:"details"`
}

// PageData represents data passed to the template.
type PageData struct {
	SiteConfig  *config.SiteConfig
	PageTitle   string
	Content     string
	Meta        map[string]interface{}
	IsHome      bool
	Hero        Hero
	Features    []Feature
	SidebarData []config.SidebarItem
	HasSidebar  bool
}

func str(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// convertMap recursively converts map[interface{}]interface{} to map[string]interface{}
func convertMap(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convertMap(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convertMap(v)
		}
	}
	return i
}

func parseHero(meta map[string]interface{}) Hero {
	var hero Hero
	if meta == nil || meta["hero"] == nil {
		return hero
	}

	// Convert map[interface{}]interface{} to map[string]interface{} for json.Marshal
	cleanData := convertMap(meta["hero"])

	b, err := json.Marshal(cleanData)
	if err == nil {
		json.Unmarshal(b, &hero)
	}
	return hero
}

func parseFeatures(meta map[string]interface{}) []Feature {
	var features []Feature
	if meta == nil || meta["features"] == nil {
		return features
	}

	cleanData := convertMap(meta["features"])

	b, err := json.Marshal(cleanData)
	if err == nil {
		json.Unmarshal(b, &features)
	}
	return features
}

// GetThemeCSS returns the base CSS for the theme.
func GetThemeCSS() string {
	return themeCSS
}

// GetThemeJS returns the client-side SPA routing script, and optionally WebSocket HMR.
func GetThemeJS(devMode bool) string {
	js := themeJS

	if devMode {
		js += `
		// Live Reload WebSocket
		const ws = new WebSocket('ws://' + location.host + '/ws');
		ws.onmessage = (e) => {
			if (e.data === 'reload') {
				console.log('Live reload triggered...');
				location.reload();
			}
		};
		ws.onclose = () => console.log('Live reload disconnected.');
		`
	}

	return js
}

// GenerateHTML renders the full HTML page using a base template.
func GenerateHTML(siteConfig *config.SiteConfig, result *markdown.RenderResult, currentPath string) (string, error) {
	tmpl, err := template.New("page").Parse(layoutHTML)
	if err != nil {
		return "", err
	}
	// TODO 调试meta的数据
	slog.Info("Meta", "meta", result.Meta)

	isHome := str(result.Meta["layout"]) == "home"
	hero := parseHero(result.Meta)
	features := parseFeatures(result.Meta)

	sidebarKey := GetSidebarKey(currentPath)
	// Determine matching sidebar based on sidebarKey
	var matchedSidebar []config.SidebarItem
	// Try to match specific path prefix first, fall back to "/"
	if sidebar, ok := siteConfig.ThemeConfig.Sidebar[sidebarKey]; ok {
		matchedSidebar = sidebar
	} else if sidebar, ok := siteConfig.ThemeConfig.Sidebar["/"]; ok {
		matchedSidebar = sidebar
	}

	data := PageData{
		SiteConfig:  siteConfig,
		PageTitle:   result.Title,
		Content:     result.HTML,
		Meta:        result.Meta,
		IsHome:      isHome,
		Hero:        hero,
		Features:    features,
		SidebarData: matchedSidebar,
		HasSidebar:  !isHome && len(matchedSidebar) > 0,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetSidebarKey 严谨修改这个函数
func GetSidebarKey(s string) string {
	lastIdx := strings.LastIndex(s, "/")
	if lastIdx == -1 {
		return "/"
	}
	return s[:lastIdx+1]
}
