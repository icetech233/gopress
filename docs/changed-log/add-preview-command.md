# 新增 `gopress preview` 命令功能说明

## 1. 背景与目标

参考 VitePress 中 `vitepress preview` 的设计，我们在 `GoPress` 中引入了 `preview` 命令。该命令旨在提供对本地构建产物（生产环境静态文件）的预览服务，并在其基础上额外增强了**热重载（Live Reload）**功能，以便在多终端操作重新构建时能自动刷新浏览器体验。

## 2. 功能特性

*   **本地预览服务器**：启动一个轻量级的 HTTP 静态服务器，将构建目录（默认 `.gopress/dist`）作为静态资源根目录进行分发。
*   **热重载支持 (Live Reload)**：有别于普通的静态服务器，我们通过在读取 HTML 文件时自动注入 WebSocket 脚本，监听 `.gopress/dist` 目录的变化。如果运行 `preview` 期间再次执行了 `build` 更新了产物，页面会自动触发重载。
*   **智能路由回退**：自动处理无后缀 URL 请求。对于像 `/guide/index` 或 `/about` 这样的请求，如果对应的 `.html` 文件存在，会自动降级匹配，对齐 VitePress 的原生路由体验。
*   **安全与容错提示**：
    *   在启动服务器前，检查目标输出目录是否存在。如果尚未执行构建，会提示用户先执行 `gopress build`。
    *   包含针对目录穿越（Directory Traversal）漏洞的基础拦截逻辑。

## 3. 代码实现细节

### 3.1 核心命令注册 (`cmd/gopress/main.go`)

新增了 `previewCmd` 命令行指令，并在 `init()` 中进行注册，与原有的 `dev`、`build` 保持一致的参数规范：

*   默认监听 `3000` 端口（可通过 `-p` 或 `--port` 选项指定）。
*   默认代理目录为 `.gopress/dist`（可通过 `-o` 或 `--outDir` 选项指定）。

```go
var previewCmd = &cobra.Command{
	Use:   "preview [root]",
	Short: "Preview the built static site",
	Long:  "Starts a local web server to preview the built static site. Supports hot-reload.",
	Run: func(cmd *cobra.Command, args []string) {
		// 容错检查：确保用户在预览前已构建站点
		if stat, err := os.Stat(outDir); os.IsNotExist(err) || !stat.IsDir() {
			fmt.Printf("Error: Output directory '%s' does not exist. Please run 'gopress build' first.\n", outDir)
			os.Exit(1)
		}
        // 启动 PreviewServer
		previewServer := &server.PreviewServer{
			OutDir: outDir,
		}
		if err := previewServer.Start(port); err != nil {
			fmt.Printf("Error starting preview server: %v\n", err)
			os.Exit(1)
		}
	},
}
```

### 3.2 预览服务器实现 (`internal/server/preview.go`)

为了职责分离，我们在 `internal/server/preview.go` 中专门封装了 `PreviewServer` 结构体：

1.  **复用 LiveReload 模块**：通过 `s.lr.Watch(s.OutDir)` 监听构建产物目录的变化。
2.  **HTML 劫持与脚本注入**：在 `handleRequest` 阶段，如果判断返回内容是 HTML 页面，会在响应写入前拦截内容，在 `</body>` 前注入如下脚本以实现热刷新功能：
    
    ```html
    <script>
    (function() {
        var protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        var ws = new WebSocket(protocol + '//' + window.location.host + '/ws');
        ws.onmessage = function(e) {
            if (e.data === 'reload') {
                window.location.reload();
            }
        };
    })();
    </script>
    ```

## 4. 使用方式

1. 先执行构建操作：
   ```bash
   gopress build
   ```
2. 启动预览服务：
   ```bash
   gopress preview
   ```
3. 也可以通过标志位指定端口和目录：
   ```bash
   gopress preview -p 8080 -o ./my-custom-dist
   ```
