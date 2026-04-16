---
title: 上下页链接功能实现
---

# 上下页链接 (Prev/Next Links) 功能实现

本文档是对 GoPress (类 VitePress 的 Go 实现版本) 中 **“上一页”与“下一页”链接功能** 的技术方案总结。

## 1. 需求背景

在浏览技术文档时，页面底部通常会提供“上一页”和“下一页”的链接导航，帮助用户顺畅地进行顺序阅读。为了提供完整的体验，我们需要实现以下核心功能：
1. **自动推断**：默认根据当前页面在“侧边栏 (Sidebar)”中的相对位置，自动推断并生成前一个和后一个页面的跳转链接。
2. **Frontmatter 覆盖**：允许用户在 Markdown 文件的 frontmatter 中通过 `prev` 和 `next` 自定义链接文本和地址。
3. **隐藏链接**：支持在 frontmatter 中通过 `prev: false` 或 `next: false` 来隐藏对应侧的链接。

---

## 2. 后端实现 (Go)

后端的职责是为模板准备好上下页的数据 (即 `PageLink` 结构体)。在 `internal/theme/theme.go` 中，我们进行了如下扩展：

### 2.1 数据结构扩充
增加 `PageLink` 结构体，并在 `PageData` 中补充 `Prev` 和 `Next` 字段。
```go
// PageLink represents a link to another page (e.g. prev/next).
type PageLink struct {
	Text string `json:"text" yaml:"text"`
	Link string `json:"link" yaml:"link"`
}

type PageData struct {
	// ... 其他原有字段
	Prev *PageLink
	Next *PageLink
}
```

### 2.2 核心解析逻辑 (`parsePrevNextLinks`)
新增了 `parsePrevNextLinks` 函数用于计算和处理链接逻辑：
1. **展平侧边栏**：递归遍历当前匹配的 `matchedSidebar` 树形结构，将其展平为一维数组，并仅保留具有 `Link` 属性的有效页面节点。
2. **定位当前索引**：在展平后的数组中查找与当前请求路径 `currentPath` 相匹配的索引位置 (`idx`)。
3. **自动推断**：
   - 若 `idx > 0`，则 `prev` 为数组中的 `idx - 1` 项。
   - 若 `idx < len - 1`，则 `next` 为数组中的 `idx + 1` 项。
4. **Frontmatter 覆盖**：
   - 读取 Markdown 解析后的 `result.Meta`。
   - 若包含 `prev` 或 `next` 配置：如果值为 `false`，则将对应链接指针置为 `nil`；如果值为包含 `text` 和 `link` 的对象结构，则反序列化为 `PageLink` 实例以覆盖推断的默认值。

在页面主渲染入口 `GenerateHTML` 中，我们将 `parsePrevNextLinks` 返回的指针直接赋值给 `PageData` 供模板使用。

---

## 3. 前端实现 (Template & CSS)

前端部分负责在页面文档底部按需渲染卡片样式的导航链接。

### 3.1 模板结构 (`layout.tmpl`)
在 `.VPDoc` 容器底部（主体 `<main>` 标签之后）新增 `<footer class="VPDocFooter">`，通过 Go 模板引擎的 `{{ if or .Prev .Next }}` 条件判断来渲染结构：

```go-html-template
<footer class="VPDocFooter">
    {{ if or .Prev .Next }}
    <div class="prev-next">
        {{ if .Prev }}
        <div class="pager">
            <a class="pager-link prev" href="{{ .Prev.Link }}">
                <span class="desc">上一页</span>
                <span class="title">{{ .Prev.Text }}</span>
            </a>
        </div>
        {{ end }}
        {{ if .Next }}
        <div class="pager" style="margin-left: auto;">
            <a class="pager-link next" href="{{ .Next.Link }}">
                <span class="desc">下一页</span>
                <span class="title">{{ .Next.Text }}</span>
            </a>
        </div>
        {{ end }}
    </div>
    {{ end }}
</footer>
```

### 3.2 页面样式 (`theme.css`)
为上下页链接实现现代化的 UI 样式，对齐 VitePress 默认风格：
- **布局**：使用 Flexbox 实现 `.prev-next` 两端对齐 (`justify-content: space-between`)，并通过 `margin-left: auto;` 确保“下一页”靠右侧排列。
- **交互与外观**：链接卡片 `.pager-link` 设置圆角、边框、以及平滑的过渡动画 (`transition`)，当鼠标悬停时边框和标题颜色呈现主题色 (`var(--vp-c-brand)`)。
- **层级感**：卡片内部的“上一页/下一页”提示字样（`.desc`）使用较小字号及次级文本颜色，强调页面的标题（`.title`）。

---

## 4. 总结
本方案采用“后端计算+前端渲染”的模式，通过展平侧边栏实现了自动推断能力，同时充分尊重用户的 Frontmatter 显式声明，以最小的性能开销达成了与 VitePress 完全一致的文档底部导读体验。