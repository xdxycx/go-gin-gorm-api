# --- Dockerfile (Builder Stage) ---
FROM golang:1.21 AS builder

WORKDIR /app

# 1. 复制 go.mod 和 go.sum (用于缓存优化)
COPY go.mod go.sum ./

# 2. 复制所有 Go 源代码文件
# 这一步是关键，它将 config/, router/ 等目录平铺到 /app
COPY ./app/ ./ 

# 3. 运行 go mod tidy
RUN go mod tidy 

# 4. 编译
# 目标文件名为 /app/main
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -o main ./


# =======================================================
# --- 阶段 2: 运行阶段 (Runner Stage) ---
# =======================================================
# 使用轻量级的 alpine 镜像作为最终运行环境
FROM alpine:latest

# 安装证书，确保HTTPS连接正常（如连接外部服务）
RUN apk --no-cache add ca-certificates

# 设置工作目录，/root/ 是惯例，但 /app/ 也可以，只要和 CMD 匹配
WORKDIR /root/ 

# 5. 🚨 关键步骤：从 builder 阶段复制编译好的可执行文件
COPY --from=builder /app/main .

# 暴露应用端口 (Gin 默认 8080)
EXPOSE 8080

# 6. 运行应用程序
CMD ["./main"]
