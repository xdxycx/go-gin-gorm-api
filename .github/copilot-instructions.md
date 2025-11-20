# Copilot / AI Agent 指导说明（项目特定）

以下为在本仓库中使 AI 编码代理立即高效工作的关键信息与约定。请严格遵循可执行示例与文件引用，避免做出与代码结构不一致的修改。

概览
- **架构位置**: 主要代码位于 `app/` 目录，包含 `app/main.go`, `app/router/router.go`, `app/config/database.go`, `app/handlers/*`, `app/models/*`, `app/utils/response.go`。
- **核心功能**: 项目提供标准的用户 CRUD（见 `app/handlers/user.go`）和一个“动态 SQL API 服务”功能（见 `app/handlers/dynamic_api.go` 与 `app/models/api_service.go`）。

运行与调试
- 本地快速运行: 在项目根目录，可运行 `go run ./app` 来启动服务（需要 `.env` 或相应环境变量）。
- Docker: 推荐使用 `docker-compose up --build -d`（`docker-compose.yml` 已定义 `app` 与 `mysql` 服务）。
- 注意: 仔细检查 `Dockerfile` 中的 `go build -o main main.go` 行。代码的 `main` 实现位于 `app/main.go`，构建命令可能需要调整为 `go build -o main ./app` 或将 `main.go` 移至根目录；在做此修改前请确认 CI/维护者意图。

重要约定与模式（务必遵守）
- 统一响应格式: 所有响应使用 `app/utils/response.go` 中的 `APIResponse`（字段 `Code`：0 表示成功，非0 表示业务错误）。示例：`c.JSON(http.StatusOK, utils.APIResponse{Code:0, Message:"查询成功", Data:...})`。
- 路由和入口: 路由在 `app/router/router.go` 中定义。API 基础路径为 `/api/v1`，用户管理在 `/api/v1/users`，动态服务注册在 `/api/v1/dynamic/register`，执行路径为 `/api/v1/dynamic/*path`（由 `dynamic.Any("/*path", handlers.ExecuteService)` 捕获）。
- 数据库初始化: 数据库由 `app/config/database.go` 的 `InitDatabase(cfg DBConfig)` 初始化，接收 `DBConfig` 接口（用于解耦与测试桩）。`InitDatabase` 使用 GORM 并在启动时 `AutoMigrate(&models.User{}, &models.APIService{})`。

动态 API 服务关键细节（必须精确实现）
- 模型: `app/models/api_service.go`。`SQL` 字段必须使用 `?` 占位符。`ParamKeys` 与 `ParamTypes` 是 JSON 数组字符串，顺序须与 SQL 中 `?` 一一对应。
- 参数与类型: 在 `app/handlers/dynamic_api.go` 中，服务在执行前会把请求参数按 `ParamKeys` 顺序读取并用 `ParamTypes` 做严格转换（支持 `int|int64`, `float|float64`, `bool`, `string`）。不要绕过或更改该转换逻辑，除非同时更新注册与文档。
- 安全限制: 动态服务只允许只读查询。允许前缀在 `app/handlers/dynamic_api.go` 的 `allowedQueryPrefixes` 中定义（默认 `SELECT, WITH, EXPLAIN, DESCRIBE, DESC`）。非允许语句会被拦截——返回 HTTP 200 但业务 `Code` 非 0，且不会执行写操作。任何更改都需要非常谨慎并注明安全理由。

常见任务示例
- 注册动态服务（示例请求体）:
  {
    "name": "GetUserByIdDynamic",
    "method": "GET",
    "path": "/user_by_id",
    "sql": "SELECT id, username, email FROM users WHERE id = ? AND deleted_at IS NULL",
    "param_keys": "[\"user_id\"]",
    "param_types": "[\"int\"]"
  }
- 执行动态服务（GET）: `GET /api/v1/dynamic/user_by_id?user_id=1`。代理需要确保对 `user_id` 做 `int` 转换，按 `ParamTypes` 验证。

开发者/AI 变更准则（只做可验证的最小改动）
- 若要修改响应格式或路由，请同时更新 `app/utils/response.go` 与所有 handlers，保持统一。
- 若要修改动态服务的允许 SQL 列表，请在 `app/handlers/dynamic_api.go` 更新 `allowedQueryPrefixes` 并在注释中注明风险与测试要点。
- 修改数据库迁移模型（`AutoMigrate`）时要考虑向后兼容性，优先创建新的迁移脚本并通知维护者。
- 对于构建/CI 变更（例如修正 `Dockerfile` 的 `go build` 路径），请先在本地通过 `go build ./app` 验证可执行文件生成，再提交变更。

文件索引（查询时优先打开）
- 入口与配置: `app/main.go`, `app/config/database.go`, `Dockerfile`, `docker-compose.yml`
- 路由与处理: `app/router/router.go`, `app/handlers/user.go`, `app/handlers/dynamic_api.go`
- 模型与迁移: `app/models/user.go`, `app/models/api_service.go`
- 响应与工具: `app/utils/response.go`

最后的提示
- 不要假设测试存在：仓库中未发现自动化测试文件（请手动验证关键路径）。
- 在对安全相关逻辑（动态 SQL、类型转换、允许语句）做出任何修改前，先在本地以 MySQL 实例和 representative SQL 测试用例验证行为。

如果本文件中有任何不清楚或遗漏的地方，请告知我需要补充的特定区域（例如：更多运行命令、CI 脚本、或特定 handler 的行为样例），我会立即迭代。
