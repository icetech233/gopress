# ============================================================
# 阶段一：构建阶段
# 使用官方 Go Alpine 镜像作为构建环境，Alpine 体积小且包含必要工具
# ============================================================
FROM golang:1.23-alpine AS builder

# 安装构建所需的系统依赖（如 git 用于 go mod 下载私有仓库）
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 优先复制 go.mod 和 go.sum，利用 Docker 层缓存机制
# 只要依赖不变，这一层就会被缓存，避免每次都重新下载依赖
COPY go.mod go.sum ./

# 下载所有依赖模块
RUN go mod download

# 复制项目源代码
COPY . .

# 编译应用程序
# - CGO_ENABLED=0：禁用 CGO，生成纯静态链接二进制文件，适合在 scratch/alpine 中运行
# - ldflags="-s -w"：去除调试信息和符号表，减小二进制文件体积
# - 输出文件名为 gopress
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gopress ./cmd/gopress

# ============================================================
# 阶段二：运行阶段
# 使用更小的基础镜像，仅包含运行时所需的最小环境
# ============================================================
FROM alpine:3.19

# 安装运行时必要的 CA 证书（用于 HTTPS 请求）和时区数据
RUN apk add --no-cache ca-certificates tzdata

# 创建非 root 用户和组，遵循最小权限原则
# - -D：不分配密码
# - -g：指定 GID
# - -u：指定 UID
# - -h：创建家目录
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup -h /home/appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/gopress .

# 创建站点内容目录并设置权限
# 用户需要将 Markdown 文档挂载到此目录
RUN mkdir -p /app/docs && chown -R appuser:appgroup /app

# 切换到非 root 用户运行应用
USER appuser

# 声明容器暴露的端口（开发服务器默认端口）
EXPOSE 5173

# 声明数据卷挂载点，用于持久化站点内容
VOLUME ["/app/docs"]

# 容器启动命令：启动开发服务器
# Go 的 http.ListenAndServe(":port", ...) 默认监听所有网络接口，无需额外配置
ENTRYPOINT ["./gopress"]
CMD ["dev", "--port", "5173", "/app/docs"]
