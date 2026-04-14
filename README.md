# GoPress

GoPress 是一个用 Go 语言开发的快速静态站点生成器，专注于 Markdown 文档的高效转换和展示。它提供了开发服务器和构建工具，让您可以快速创建、预览和部署文档网站。

## 技术栈

- **编程语言**: Go 1.23.0+
- **Markdown 渲染**: Goldmark
- **命令行框架**: Cobra
- **实时重载**: Gorilla WebSocket + fsnotify
- **语法高亮**: Chroma
- **配置管理**: YAML / JSON

## 安装指南

### 系统环境要求

- Go 1.24.3 或更高版本

### 分步骤安装说明

1. **克隆仓库**

   ```bash
   git clone https://github.com/icetech233/gopress.git
   cd gopress
   ```

2. **安装依赖**

   ```bash
   go mod download
   ```

3. **构建项目**

   ```bash
   go build .\cmd\gopress
   ```

   或者在 Linux/macOS 上：

   ```bash
   go build ./cmd/gopress
   ```

## 使用说明

### 启动开发服务器

```bash
# 在当前目录启动开发服务器，默认端口 5173
gopress dev

# 指定端口
gopress dev -p 8080

# 指定项目根目录
gopress dev ./docs
```

开发服务器支持实时重载功能，当您修改 Markdown 文件时，浏览器会自动刷新显示最新内容。

### 构建生产版本

```bash
# 构建到默认目录 .gopress/dist
gopress build

# 指定输出目录
gopress build -o ./output

# 指定项目根目录
gopress build ./docs -o ./public
```

### 基本使用流程

1. 创建项目目录并初始化
2. 编写 Markdown 文档文件
3. 使用 `gopress dev` 启动开发服务器进行预览
4. 使用 `gopress build` 构建静态网站
5. 部署构建产物到您的 Web 服务器

## 参数详解

### 全局配置

配置文件应放置在 `.gopress/config.json` 或 `.gopress/config.yaml` 中。

#### 站点基础配置

| 参数名称 | 数据类型 | 默认值 | 详细说明 | 使用示例 |
|---------|---------|--------|---------|---------|
| `title` | string | "gopress" | 网站标题 | `"title": "我的文档"` |
| `description` | string | "A gopress website" | 网站描述 | `"description": "技术文档网站"` |
| `base` | string | "/" | 站点基础路径 | `"base": "/docs/"` |
| `lang` | string | "en-US" | 网站语言 | `"lang": "zh-CN"` |

#### 主题配置 (themeConfig)

##### Logo

| 参数名称 | 数据类型 | 默认值 | 详细说明 | 使用示例 |
|---------|---------|--------|---------|---------|
| `logo` | string | - | 网站 Logo 路径 | `"logo": "/assets/logo.png"` |

##### 导航栏 (nav)

导航栏配置是一个数组，每个元素包含：

| 参数名称 | 数据类型 | 默认值 | 详细说明 | 使用示例 |
|---------|---------|--------|---------|---------|
| `text` | string | - | 导航文本 | `"text": "指南"` |
| `link` | string | - | 导航链接 | `"link": "/guide/"` |
| `activeMatch` | string | - | 激活匹配规则 | `"activeMatch": "^/guide/"` |

**示例**:

```json
{
  "themeConfig": {
    "nav": [
      {
        "text": "指南",
        "link": "/guide/",
        "activeMatch": "^/guide/"
      },
      {
        "text": "API",
        "link": "/api/"
      }
    ]
  }
}
```

##### 侧边栏 (sidebar)

侧边栏是一个对象，键为路径前缀，值为侧边栏项目数组。

| 参数名称 | 数据类型 | 默认值 | 详细说明 | 使用示例 |
|---------|---------|--------|---------|---------|
| `text` | string | - | 侧边栏文本 | `"text": "快速开始"` |
| `link` | string | - | 侧边栏链接（可选） | `"link": "/guide/quick-start.html"` |
| `collapsed` | boolean | false | 是否默认折叠（可选） | `"collapsed": true` |
| `items` | array | - | 子项数组（可选） | `"items": [...]` |

**示例**:

```json
{
  "themeConfig": {
    "sidebar": {
      "/guide/": [
        {
          "text": "介绍",
          "link": "/guide/index.html"
        },
        {
          "text": "快速开始",
          "collapsed": false,
          "items": [
            {
              "text": "安装",
              "link": "/guide/install.html"
            },
            {
              "text": "配置",
              "link": "/guide/config.html"
            }
          ]
        }
      ]
    }
  }
}
```

##### 社交链接 (socialLinks)

| 参数名称 | 数据类型 | 默认值 | 详细说明 | 使用示例 |
|---------|---------|--------|---------|---------|
| `icon` | string | - | 图标类型 | `"icon": "github"` |
| `link` | string | - | 链接地址 | `"link": "https://github.com/icetech233/gopress"` |

**示例**:

```json
{
  "themeConfig": {
    "socialLinks": [
      {
        "icon": "github",
        "link": "https://github.com/icetech233/gopress"
      }
    ]
  }
}
```

##### 页脚 (footer)

页脚配置是一个键值对对象。

**示例**:

```json
{
  "themeConfig": {
    "footer": {
      "copyright": "Copyright © 2024 IceTech",
      "message": "Powered by GoPress"
    }
  }
}
```

### 命令行参数

#### dev 命令

| 参数名称 | 简写 | 数据类型 | 默认值 | 详细说明 | 使用示例 |
|---------|------|---------|--------|---------|---------|
| `port` | `p` | int | 5173 | 开发服务器监听端口 | `gopress dev -p 3000` |

#### build 命令

| 参数名称 | 简写 | 数据类型 | 默认值 | 详细说明 | 使用示例 |
|---------|------|---------|--------|---------|---------|
| `outDir` | `o` | string | ".gopress/dist" | 输出目录 | `gopress build -o ./public` |

## 开发指南

### 如何贡献代码

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

### 项目结构

```
gopress/
├── cmd/
│   └── gopress/
│       └── main.go          # 命令行入口
├── internal/
│   ├── build/               # 构建相关逻辑
│   ├── config/              # 配置管理
│   ├── markdown/            # Markdown 渲染
│   ├── server/              # 开发服务器
│   └── theme/               # 主题模板
├── docs/                    # 文档示例
├── go.mod
└── go.sum
```

### 构建流程

```bash
# 开发模式
go run ./cmd/gopress dev ./docs

# 构建二进制文件
go build -o gopress.exe ./cmd/gopress  # Windows
go build -o gopress ./cmd/gopress       # Linux/macOS
```

## FAQ 常见问题

### Q: 如何自定义主题样式？

A: 目前 GoPress 使用内置主题。您可以通过修改 `internal/theme/theme.go` 中的 CSS 来自定义样式。未来版本将支持外部主题配置。

### Q: 支持哪些 Markdown 扩展语法？

A: GoPress 使用 Goldmark 渲染引擎，支持：
- 标准 Markdown
- GFM (GitHub Flavored Markdown)
- 语法高亮
- 元数据 (Front Matter)

### Q: 如何部署生成的静态网站？

A: 构建后的静态文件位于 `.gopress/dist` 目录，您可以将其部署到任何静态文件托管服务，如：
- GitHub Pages
- Vercel
- Netlify
- Nginx
- 阿里云 OSS / CDN

### Q: 实时重载不工作怎么办？

A: 请检查：
1. 确保没有防火墙阻止 WebSocket 连接
2. 确保文件系统权限正确
3. 尝试刷新浏览器页面

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。
