---
prev:
  text: '侧边栏状态重构'
  link: '/history/sidebar-state-refactor.html'
next:
  text: '折叠状态持久化'
  link: '/history/sidebar-localstorage-persistence.html'
---

# Logo功能实现

## 1. 背景

随着暗色模式（Dark Mode）的普及，部分站点的 Logo 在亮色和暗色背景下可能需要不同的图片资产，以确保最佳的视觉效果。为了满足这一需求，我们在 GoPress 中对 Logo 属性进行了重构，引入了 `ThemeableImage` 的概念。

## 2. 功能特性

重构后的 Logo 配置现支持两种格式，并且保证了向下兼容：

- **传统格式（字符串）**：仅配置一个图片，在任何主题下都显示该图片。
- **对象格式**：可以分别为亮色模式和暗色模式指定不同的图片，还可以配置 `alt` 文本。

### 示例配置 (`config.json`)

```json
{
  "themeConfig": {
    "logo": {
      "light": "/logo-light.png",
      "dark": "/logo-dark.png",
      "alt": "站点Logo"
    }
  }
}
```

## 3. 技术实现细节

该功能的实现贯穿了后端配置解析、前端模板渲染以及 CSS 样式控制三个层面：

### 3.1 后端配置解析 (Go)

我们在 `internal/config/config.go` 中引入了 `ThemeableImage` 结构体：

```go
type ThemeableImage struct {
	Light string `json:"light,omitempty" yaml:"light,omitempty"`
	Dark  string `json:"dark,omitempty" yaml:"dark,omitempty"`
	Alt   string `json:"alt,omitempty" yaml:"alt,omitempty"`
}
```

为了**保证向下兼容性**，我们为 `ThemeableImage` 实现了自定义的 `UnmarshalJSON` 和 `UnmarshalYAML` 方法。当解析器遇到原有的纯字符串格式时，它会自动将其映射到 `Light` 字段，并将 `Dark` 留空。

### 3.2 前端模板渲染 (HTML)

在 `internal/theme/layout.tmpl` 模板的导航栏部分，我们加入了更复杂的条件渲染逻辑：

- 当同时存在 `Light` 和 `Dark` 时，会**同时渲染两个 `<img>` 标签**，并分别打上 `light-only` 和 `dark-only` 类名。

```html
<div class="title">
	<a href="{{ .SiteConfig.Base }}">
		{{ if .SiteConfig.ThemeConfig.Logo.Light }}<img src="{{ .SiteConfig.ThemeConfig.Logo.Light }}" alt="{{ .SiteConfig.ThemeConfig.Logo.Alt }}" class="logo light-only">{{ end }}
		{{ if .SiteConfig.ThemeConfig.Logo.Dark }}<img src="{{ .SiteConfig.ThemeConfig.Logo.Dark }}" alt="{{ .SiteConfig.ThemeConfig.Logo.Alt }}" class="logo dark-only">{{ end }}
		{{ .SiteConfig.Title }}
	</a>
</div>
```

### 3.3 样式控制 (CSS)

为了避免页面在加载时判断主题时的闪烁（FOUC），并且无需编写复杂的 JavaScript 监听器，我们采用了纯 CSS 的方式控制主题图片的显隐。在 `internal/theme/theme.css` 中添加了如下规则：

```css
/* 当主题为暗色时，隐藏仅亮色可见的元素 */
html.dark .light-only { display: none !important; }

/* 当主题不为暗色时，隐藏仅暗色可见的元素 */
html:not(.dark) .dark-only { display: none !important; }
```

依托于 `html.dark` 这个类（这通常在 `<head>` 中极早的同步脚本中设置），我们可以做到图片渲染时的无缝切换，完全规避闪烁问题。

## 4. 总结

通过数据结构升级、实现自定义的反序列化方法以及运用纯 CSS 的显隐策略，我们完美实现了对多主题 Logo 的支持。