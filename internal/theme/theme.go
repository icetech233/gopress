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

// PageLink represents a link to another page (e.g. prev/next).
type PageLink struct {
	Text string `json:"text" yaml:"text"`
	Link string `json:"link" yaml:"link"`
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
	Prev        *PageLink
	Next        *PageLink
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

// parsePrevNextLinks 从侧边栏推断上一页/下一页链接，或者从页面前置元数据中读取
// 参数：
//
//	meta - 页面前置元数据
//	matchedSidebar - 当前匹配的侧边栏配置
//	currentPath - 当前页面路径
//
// 返回值：
//
//	上一页链接和下一页链接的指针，如果不存在则返回 nil
func parsePrevNextLinks(meta map[string]interface{}, matchedSidebar []config.SidebarItem, currentPath string) (*PageLink, *PageLink) {
	var prev, next *PageLink

	// 1. 将嵌套的侧边栏结构扁平化为一维数组，方便查找当前项及其相邻项
	flat := flattenSidebar(matchedSidebar)

	// 2. 在扁平化后的数组中查找当前页面的索引
	idx := -1
	for i, item := range flat {
		// 匹配当前页面路径，考虑.html 后缀的情况
		if item.Link == currentPath || item.Link == currentPath+".html" || item.Link+".html" == currentPath {
			idx = i
			break
		}
	}

	// 3. 从侧边栏推断上一页和下一页链接
	if idx > 0 {
		prev = &PageLink{Text: flat[idx-1].Text, Link: flat[idx-1].Link}
	}
	if idx >= 0 && idx < len(flat)-1 {
		next = &PageLink{Text: flat[idx+1].Text, Link: flat[idx+1].Link}
	}

	// 4. 使用前置元数据中的配置覆盖侧边栏推断的链接
	if meta != nil {
		prev = parsePageLinkFromMeta(meta, "prev", prev)
		next = parsePageLinkFromMeta(meta, "next", next)
	}

	return prev, next
}

// flattenSidebar 将嵌套的侧边栏结构扁平化为一维数组
// 只保留带有链接的侧边栏项，递归处理所有子菜单
func flattenSidebar(items []config.SidebarItem) []config.SidebarItem {
	var flat []config.SidebarItem
	var flatten func(items []config.SidebarItem)
	flatten = func(items []config.SidebarItem) {
		for _, item := range items {
			// 只添加带有链接的侧边栏项
			if item.Link != "" {
				flat = append(flat, item)
			}
			// 递归处理子菜单
			if len(item.Items) > 0 {
				flatten(item.Items)
			}
		}
	}
	flatten(items)
	return flat
}

// parsePageLinkFromMeta 从元数据中解析页面链接
// 参数：
//
//	meta - 页面前置元数据
//	key - 元数据键名（"prev"或"next"）
//	defaultLink - 默认链接，当元数据未配置时使用
//
// 返回值：
//
//	解析后的 PageLink 指针，如果禁用或解析失败则返回 nil 或默认值
func parsePageLinkFromMeta(meta map[string]interface{}, key string, defaultLink *PageLink) *PageLink {
	v, ok := meta[key]
	if !ok {
		return defaultLink
	}

	// 如果设置为 false，则禁用链接
	if b, isBool := v.(bool); isBool && !b {
		return nil
	}

	// 将前置元数据转换为 PageLink 结构体
	cleanData := convertMap(v)
	b, err := json.Marshal(cleanData)
	if err == nil {
		var pl PageLink
		if err := json.Unmarshal(b, &pl); err == nil && pl.Text != "" {
			return &pl
		}
	}

	return defaultLink
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

	prev, next := parsePrevNextLinks(result.Meta, matchedSidebar, currentPath)

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
		Prev:        prev,
		Next:        next,
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
