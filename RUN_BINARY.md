# 运行已编译 Go 单程序（单二进制）指南

本文档说明如何将本仓库编译为单个可执行文件并在目标主机上运行。包括构建命令、运行示例、使用环境变量、systemd 单元示例与常见运维注意事项。

**前提**
- 已安装 Go（用于构建）。推荐 Go 1.20+。
- 目标主机需要相应的运行时（Linux 可直接运行编译好的 ELF）。若希望生成静态二进制，请参考构建选项。

1) 在开发机上构建单文件可执行程序

在仓库根目录下执行（构建 `./app` 包）：

```bash
# 构建到可执行文件 `app`（在当前平台/架构）
go build -o app ./app

# 或者显式指定包路径（等价）：
go build -o go-gin-gorm-api ./app
```

可选：生成适用于 linux/amd64 的静态二进制（方便在目标主机上直接运行）：

```bash
# 在支持 CGO 的代码若无 cgo 依赖，可禁用 CGO 以便生成静态二进制
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o app ./app
```

交叉编译示例（在 macOS 或其他平台为 linux/amd64 构建）：

```bash
env GOOS=linux GOARCH=amd64 go build -o app ./app
```

2) 将可执行文件部署到目标主机

- 把生成的 `app` 文件复制到目标主机（例如 `/opt/go-gin-gorm-api/app`）：

```bash
scp app user@target:/opt/go-gin-gorm-api/app
```

- 在目标主机上设置可执行权限：

```bash
chmod +x /opt/go-gin-gorm-api/app
```

3) 环境变量与配置

该程序读取环境变量（与 `docker-compose` / `.env` 中同名），常用环境变量示例：

- `PORT`：应用监听端口（例如 `8080`）
- `MYSQL_HOST`, `MYSQL_PORT`, `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DATABASE`, `MYSQL_ROOT_PASSWORD`
- `DYNAMIC_MAX_ROWS`：动态 SQL 返回最大行数（默认 1000）
- `DYNAMIC_QUERY_TIMEOUT_SECONDS`：动态 SQL 执行超时（秒，默认 5）

推荐将环境变量放到 `.env` 文件或 systemd 单元环境文件中。运行前可以使用 `env` 前缀或 `source`：

```bash
# 临时在当前 shell 中加载 .env（简单方法，基于 bash）
set -o allexport; source /opt/go-gin-gorm-api/.env; set +o allexport

# 然后运行
/opt/go-gin-gorm-api/app
```

或者直接在命令行传入：

```bash
PORT=8080 MYSQL_HOST=127.0.0.1 MYSQL_USER=user MYSQL_PASSWORD=pass /opt/go-gin-gorm-api/app
```

4) 启动与后台运行

- 前台运行（用于调试）：

```bash
/opt/go-gin-gorm-api/app
```

- 后台运行（示例使用 `nohup`）：

```bash
nohup /opt/go-gin-gorm-api/app > /var/log/go-gin-gorm-api.log 2>&1 &
```

5) systemd 单元示例（建议在 Linux 生产环境使用）

将下面的内容保存为 `/etc/systemd/system/go-gin-gorm-api.service`，根据实际路径和用户调整 `User`、`ExecStart` 与环境文件路径：

```ini
[Unit]
Description=Go Gin Gorm API
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/go-gin-gorm-api
EnvironmentFile=/opt/go-gin-gorm-api/.env
ExecStart=/opt/go-gin-gorm-api/app
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

启动并启用服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now go-gin-gorm-api.service
sudo journalctl -u go-gin-gorm-api --no-pager -f
```

6) 日志和诊断

- 若使用 systemd，使用 `journalctl -u go-gin-gorm-api` 查看日志。
- 若使用 `nohup` 或重定向，请查看指定的日志文件（如 `/var/log/go-gin-gorm-api.log`）。
- 启动失败常见原因：环境变量未设置、数据库不可达、端口被占用、权限问题。

7) 升级流程（零停机不是自动支持，示例为简单重启）

简单升级：

```bash
# 停止服务
sudo systemctl stop go-gin-gorm-api

# 替换可执行文件（先备份旧版本）
mv /opt/go-gin-gorm-api/app /opt/go-gin-gorm-api/app.bak
scp app user@target:/opt/go-gin-gorm-api/app
chmod +x /opt/go-gin-gorm-api/app

# 启动服务
sudo systemctl start go-gin-gorm-api
```

更安全的升级策略：
- 使用负载均衡器 + 多实例进行滚动升级
- 使用临时端口健康检查并完成就绪探针

8) 权限与安全建议

- 不要以 `root` 用户直接运行应用，创建专用运行用户（例如 `www-data` 或 `goapi`）。
- 把包含敏感信息的 `.env` 文件权限设置为仅运行用户可读：

```bash
chown goapi:goapi /opt/go-gin-gorm-api/.env
chmod 600 /opt/go-gin-gorm-api/.env
```

- 在生产环境中使用更安全的密钥管理（Kubernetes Secrets、Vault、云平台 Secret Manager）。

9) 常见问题排查

- 程序无法连接数据库：检查 `MYSQL_HOST` / `MYSQL_PORT` / `MYSQL_USER` / `MYSQL_PASSWORD` 是否正确，目标端口是否可达。
- 程序启动但 404/500：查看日志，确认路由是否正确注册，数据库迁移是否完成。
- 端口被占用：使用 `ss -ltnp | grep <PORT>` 或 `lsof -i:<PORT>` 查找占用进程。

10) 附加：生成带版本信息的二进制

可以在构建时通过 `-ldflags` 注入版本、构建时间与 git commit：

```bash
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD || echo dev)
go build -ldflags "-X 'main.BuildTime=${BUILD_TIME}' -X 'main.GitCommit=${GIT_COMMIT}'" -o app ./app
```

（注意：此处 `main.BuildTime` 与 `main.GitCommit` 需在代码中定义为可被 `-X` 注入的变量。）

---

如果你需要，我可以：
- 把此文档内容追加到 `README.md` 的部署部分，或
- 添加一个 `systemd` 模板到 `scripts/` 下并把 `--skip-checks` 参数加入脚本便于 CI 使用。
