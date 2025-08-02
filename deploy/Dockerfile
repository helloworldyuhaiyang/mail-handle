# 使用多阶段构建
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o mail-handle-linux \
    cmd/main.go

# 运行阶段
FROM ccr.ccs.tencentyun.com/library/alpine:latest

# 安装必要的工具
RUN apk add --no-cache wget curl ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建必要的目录
RUN mkdir -p /config /app/data

# 从构建阶段复制二进制文件
COPY --from=builder /app/mail-handle-linux /mail-handle-linux

# 复制配置文件（如果存在）
COPY --from=builder /app/config/default.yaml /config/default.yaml

# 设置执行权限
RUN chmod +x /mail-handle-linux

# 设置工作目录
WORKDIR /

# 设置健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# 暴露端口
EXPOSE 8082

# 启动命令
ENTRYPOINT [ "./mail-handle-linux", "--config", "/config/default.yaml", "run"]
