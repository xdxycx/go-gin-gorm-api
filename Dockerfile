# --- build stage ---
FROM golang:1.22-alpine AS builder
WORKDIR /src

# 复制模块文件并下载依赖
COPY go.mod go.sum ./
RUN apk add --no-cache git && go mod download

# 复制项目
COPY . .

# 使用静态构建（根据需要可删除 CGO_ENABLED）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-s -w" -o /build/app ./app

# --- runtime stage ---
FROM alpine:latest
# 安装证书并提供 wget 用于 HEALTHCHECK
RUN apk --no-cache add ca-certificates wget
# 创建非 root 用户
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app

# 复制二进制
COPY --from=builder /build/app /app/app
# 不复制 .env（由 docker-compose 或运行时注入环境变量）
# COPY --from=builder /src/.env .   <-- 不建议

RUN chown app:app /app/app
USER app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s CMD wget -qO- --timeout=2 http://localhost:8080/ || exit 1

CMD ["/app/app"]
