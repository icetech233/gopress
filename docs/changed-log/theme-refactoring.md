# theme.go 代码重构文档

## 变更概述

本次重构对 `vitepress-go/internal/theme/theme.go` 文件中的 `parsePrevNextLinks` 函数及其相关代码进行了全面优化，包括中文注释补充、重复代码消除和函数拆分。

## 变更文件

- `vitepress-go/internal/theme/theme.go`

## 详细变更内容

### 1. 函数注释中文化

**原注释：**
```go
// parsePrevNextLinks infers prev/next links from the sidebar or reads from frontmatter.
```

**更新为：**
```go
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
```

### 2. 新增辅助函数 flattenSidebar

**目的：** 将嵌套的侧边栏结构扁平化逻辑独立封装

**代码：**
```go
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
```

**使用方式：**
```go
// 原代码（15 行）
var flat []config.SidebarItem
var flatten func(items []config.SidebarItem)
flatten = func(items []config.SidebarItem) {
    for _, item := range items {
        if item.Link != "" {
            flat = append(flat, item)
        }
        if len(item.Items) > 0 {
            flatten(item.Items)
        }
    }
}
flatten(matchedSidebar)

// 优化后（1 行）
flat := flattenSidebar(matchedSidebar)
```

### 3. 新增辅助函数 parsePageLinkFromMeta

**目的：** 消除处理 prev 和 next 元数据时的重复代码

**代码：**
```go
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
```

**使用方式：**
```go
// 原代码（36 行重复代码）
if meta != nil {
    if p, ok := meta["prev"]; ok {
        if b, isBool := p.(bool); isBool && !b {
            prev = nil
        } else {
            cleanData := convertMap(p)
            b, err := json.Marshal(cleanData)
            if err == nil {
                var pl PageLink
                if err := json.Unmarshal(b, &pl); err == nil && pl.Text != "" {
                    prev = &pl
                }
            }
        }
    }
    if n, ok := meta["next"]; ok {
        if b, isBool := n.(bool); isBool && !b {
            next = nil
        } else {
            cleanData := convertMap(n)
            b, err := json.Marshal(cleanData)
            if err == nil {
                var pl PageLink
                if err := json.Unmarshal(b, &pl); err == nil && pl.Text != "" {
                    next = &pl
                }
            }
        }
    }
}

// 优化后（2 行）
if meta != nil {
    prev = parsePageLinkFromMeta(meta, "prev", prev)
    next = parsePageLinkFromMeta(meta, "next", next)
}
```

### 4. 步骤注释中文化

所有步骤注释均从英文更新为中文：

| 原英文注释 | 新中文注释 |
|-----------|-----------|
| `// 1. Flatten the sidebar to find the current item and its neighbors` | `// 1. 将嵌套的侧边栏结构扁平化为一维数组，方便查找当前项及其相邻项` |
| `// 2. Find current index` | `// 2. 在扁平化后的数组中查找当前页面的索引` |
| `// 3. Infer from sidebar` | `// 3. 从侧边栏推断上一页和下一页链接` |
| `// 4. Override with frontmatter` | `// 4. 使用前置元数据中的配置覆盖侧边栏推断的链接` |

### 5. 关键逻辑注释补充

为以下关键代码块添加了中文注释：

- `// 只添加带有链接的侧边栏项`
- `// 递归处理子菜单`
- `// 匹配当前页面路径，考虑.html 后缀的情况`
- `// 如果设置为 false，则禁用链接`
- `// 将前置元数据转换为 PageLink 结构体`

## 重构前后对比

### 代码行数

| 函数 | 重构前 | 重构后 | 变化 |
|-----|-------|-------|------|
| `parsePrevNextLinks` | ~70 行 | ~30 行 | -40 行 |
| 新增 `flattenSidebar` | - | ~20 行 | +20 行 |
| 新增 `parsePageLinkFromMeta` | - | ~25 行 | +25 行 |

### 代码质量提升

1. **可读性提升**
   - 中文注释降低理解门槛
   - 函数职责更加清晰
   - 主函数逻辑更简洁

2. **可维护性提升**
   - 消除重复代码（DRY 原则）
   - 函数可独立测试
   - 便于后续扩展

3. **代码复用**
   - `flattenSidebar` 可被其他函数复用
   - `parsePageLinkFromMeta` 统一处理元数据解析

## 变更影响

- ✅ 代码逻辑完全保持不变
- ✅ 仅对代码结构和注释进行优化
- ✅ 提高中文开发者的代码阅读体验
- ✅ 便于后续维护和扩展
- ✅ 无外部依赖变更
- ✅ 无 API 接口变更

## 测试建议

建议对以下场景进行测试：

1. 侧边栏正常导航（上一页/下一页）
2. 前置元数据中配置 `prev: false` 或 `next: false`
3. 前置元数据中自定义 `prev` 或 `next` 的文本和链接
4. 嵌套侧边栏菜单的扁平化
5. 页面路径带.html 后缀和不带的情况

## 变更时间

2026 年 4 月 16 日
