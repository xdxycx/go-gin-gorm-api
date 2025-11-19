# --- 构建阶段 ---
FROM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 并下载依赖
COPY go.mod .
COPY go.sum .
RUN go mod download

# 复制整个项目到工作目录
COPY . .

# CGO_ENABLED=0 交叉编译静态链接的二进制文件
# -ldflags "-s -w" 减少二进制文件大小
RUN go build -o main main.go

# --- 运行阶段 ---
FROM alpine:latest

# 安装证书 (如果需要 HTTPS)
RUN apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/.env . # 复制 .env 文件到运行环境

# 暴露应用端口
EXPOSE 8080

# 运行应用
CMD ["./main"]
