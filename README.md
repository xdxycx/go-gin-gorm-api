Go Gin Gorm 动态 API 项目手册

I. 项目简介 (Project Introduction)

go-gin-gorm-api 是一个基于 Go 语言，使用 Gin 框架和 Gorm ORM 构建的 RESTful API 服务模板。该项目设计旨在提供快速开发的基础结构，并包含一个独特的核心功能：动态 SQL API 服务。

核心特性允许用户通过注册接口配置 SQL 语句，使数据库操作通过标准 HTTP 接口动态暴露，无需修改后端代码即可实现新的数据查询或操作逻辑。

II. 核心功能 (Core Features)

A. 用户管理模块 (User Management)

提供了用户资源的标准的 CRUD（创建、读取、更新、删除）操作接口，是基础业务功能实现的一个示例。

模型: models.User (包含 ID, Username, Email, CreatedAt 等字段，并使用 Gorm 软删除)

操作: 提供了完整的增删改查 API。

B. 动态 API 服务模块 (Dynamic API Service)

这是项目的核心创新功能。它允许开发者通过 HTTP POST 请求注册一个新的 API 服务，该服务与一个自定义的 SQL 语句绑定。

关键特性:

灵活注册: 可注册 GET, POST, PUT, DELETE 等任意 HTTP 方法的服务。

SQL 绑定: 每个服务都绑定一个带有 ? 占位符的原始 SQL 语句。

参数顺序锁定 (ParamKeys): 使用 JSON 数组定义请求参数（如 URL Query 或 Request Body）的名称，顺序必须与 SQL 语句中的 ? 占位符顺序严格一致。

强制类型转换 (ParamTypes): 使用 JSON 数组定义每个参数的预期 Go 类型（如 "int", "string", "float", "bool"）。系统在执行 SQL 前会对参数进行严格的类型转换，确保 SQL 执行的健壮性。

严格查询限制（安全策略）:

为了保护数据库安全，动态服务仅允许执行只读的查询和诊断操作。任何修改数据库状态的语句（如 INSERT, UPDATE, DELETE, DROP 等）都会被系统拦截并记录日志，而不会执行。

允许的语句前缀 (不区分大小写):

SELECT (标准查询)

WITH (CTE，复杂查询)

EXPLAIN (查询执行计划诊断)

DESCRIBE / DESC (查看表结构)

III. 环境搭建与启动 (Setup and Startup)

项目推荐使用 Docker Compose 进行快速、完整的环境搭建。

A. 依赖 (Prerequisites)

Docker

Docker Compose

B. Docker Compose 启动

配置 .env 文件:

确保项目根目录下的 .env 文件配置正确，特别是数据库的连接信息。

MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_USER=user
MYSQL_PASSWORD=password
MYSQL_DATABASE=godb
MYSQL_ROOT_PASSWORD=rootpassword
PORT=8080


启动服务:

在项目根目录下执行 Docker Compose 命令，它将自动构建 Go 应用镜像、启动 MySQL 容器，并运行应用。

docker-compose up --build -d


验证:

服务启动后，可以通过访问健康检查接口验证应用是否正常运行：

curl http://localhost:8080/
# 预期输出: {"message":"Welcome to Go Gin Gorm API"}

C. 配置项与环境变量（运行时）
D. 使用 `.env` 与 Docker Compose 部署（示例）

1. 复制示例环境文件并编辑实际值：

```bash
cp .env.example .env
# 编辑 .env，填写数据库密码等敏感信息
```

2. 使用 `docker-compose` 启动（推荐，包含构建步骤）：

```bash
# 使用默认 .env 文件
docker-compose --env-file .env up --build -d

# 查看日志，确认服务与数据库启动完成
docker-compose logs -f app

# 停止并移除容器
docker-compose down
```

3. 覆盖环境变量示例：

```bash
# 以临时变量覆盖并启动
DYNAMIC_MAX_ROWS=500 DYNAMIC_QUERY_TIMEOUT_SECONDS=10 docker-compose --env-file .env up --build -d
```

说明:
- `.env.example` 包含运行所需的环境变量示例（包括 `DYNAMIC_MAX_ROWS` 和 `DYNAMIC_QUERY_TIMEOUT_SECONDS` 的默认值）。
- 在生产环境中，尽量通过 CI/CD secret 或 Docker secrets 管理敏感配置，不要把生产密码提交到仓库。


项目的一些运行时行为可通过环境变量配置，推荐在 `docker-compose.yml` 或 `.env` 中设置：

- `DYNAMIC_MAX_ROWS`：执行返回的最大行数，默认值 `1000`。当查询结果超过该值时，API 会截断返回并在响应中标注 `truncated`。示例：`DYNAMIC_MAX_ROWS=500`
- `DYNAMIC_QUERY_TIMEOUT_SECONDS`：动态 SQL 执行的超时时间（秒），默认值 `5`。超过该时间查询将被取消并返回业务码 `2`（超时）。示例：`DYNAMIC_QUERY_TIMEOUT_SECONDS=10`

审计日志说明：动态 SQL 的每次执行都会生成一条持久化审计记录（表名 `audits`），记录 `path`, `method`, `client_ip`, `sql`, `args`, `duration_ms`, `rows`, `truncated` 等字段。审计表由应用在启动时通过 GORM 的 `AutoMigrate` 自动创建。


IV. API 接口参考 (API Reference)

API 基础路径为 /api/v1。

A. 用户接口 (User Endpoints)

方法

路径

描述

请求体示例

响应示例

POST

/api/v1/users

创建新用户

{"username": "test_user", "email": "test@example.com"}

{"code": 0, ...}

GET

/api/v1/users

获取所有用户

(无)

{"code": 0, "data": [...]}

GET

/api/v1/users/:id

根据 ID 获取用户

(无)

{"code": 0, "data": {...}}

PUT

/api/v1/users/:id

更新指定用户

{"username": "new_name"}

{"code": 0, "data": {...}}

DELETE

/api/v1/users/:id

软删除指定用户

(无)

{"code": 0, "message": "删除成功"}

B. 动态服务接口 (Dynamic Service Endpoints)

1. 服务注册 (Registration)

方法

路径

描述

请求体示例

POST

/api/v1/dynamic/register

注册一个新的动态 API 服务

见下文示例

注册请求体示例：查询指定 ID 用户

该服务将注册为 GET /api/v1/dynamic/user_by_id，需要一个名为 id 的整数参数。

{
  "name": "GetUserByIdDynamic",
  "method": "GET",
  "path": "/user_by_id",
  "sql": "SELECT id, username, email FROM users WHERE id = ? AND deleted_at IS NULL",
  "param_keys": "[\"user_id\"]",
  "param_types": "[\"int\"]" 
}


2. 服务执行 (Execution)

注册成功后，通过 GET /api/v1/dynamic/*path 或其他方法路径访问。

示例 1: 执行 GET 查询 (使用 URL Query 参数)

假设已注册上述 GetUserByIdDynamic 服务。

请求: GET http://localhost:8080/api/v1/dynamic/user_by_id?user_id=1

参数: user_id (作为字符串 "1" 传入)

执行: 系统会将其强制转换为 int 类型 (1)，并执行 SQL。

响应: {"code": 0, "message": "查询成功", "data": [...]}

示例 2: 尝试执行被禁用的写入操作

假设尝试注册一个 UPDATE 服务并调用它。

注册 SQL: "UPDATE users SET username = ? WHERE id = ?"

请求: POST http://localhost:8080/api/v1/dynamic/update_user_name

请求体: {"user_id": 2, "new_name": "Alice_Updated"}

响应 (操作被阻止):

{
  "code": 1, 
  "message": "安全限制: 动态服务只允许执行 [SELECT WITH EXPLAIN DESCRIBE DESC ] 查询操作。非查询操作已被阻止。",
  "data": {
    "sql_statement_type": "UPDATE"
  }
}


(注意：系统返回 HTTP 200，但业务代码 code: 1 和明确的消息表明操作已被安全策略拦截，未执行。)

**Deployment & Scripts**

- `DEPLOY.md`: 详尽的部署指南，位于仓库根目录，包含基于 `docker-compose` 的启动、验证与停止步骤。
- `scripts/start.sh`: 启动 helper，会在缺少 `.env` 时从 `.env.example` 复制一份并使用 `docker-compose` 启动服务，同时跟随 `app` 日志。
- `scripts/stop.sh`: 停止并清理容器/镜像/卷的辅助脚本。
- `scripts/status.sh`: 快速展示 `docker-compose ps`、最近的 `app` 日志与宿主端口映射信息。

示例用法：

```bash
# 复制并编辑 .env（仅第一次或需要修改时）
cp .env.example .env
# 启动（脚本会跟随日志）
./scripts/start.sh
# 停止并清理
./scripts/stop.sh
# 查看当前状态与日志摘要
./scripts/status.sh
```

建议：将 `DEPLOY.md` 与这些脚本作为本地开发与 CI 快速验证的参考；在生产环境中使用更成熟的秘密管理与部署机制（例如 Kubernetes + Secrets / HashiCorp Vault）。
Go Gin Gorm 动态 API 项目手册

I. 项目简介 (Project Introduction)

go-gin-gorm-api 是一个基于 Go 语言，使用 Gin 框架和 Gorm ORM 构建的 RESTful API 服务模板。该项目设计旨在提供快速开发的基础结构，并包含一个独特的核心功能：动态 SQL API 服务。

核心特性允许用户通过注册接口配置 SQL 语句，使数据库操作通过标准 HTTP 接口动态暴露，无需修改后端代码即可实现新的数据查询或操作逻辑。

II. 核心功能 (Core Features)

A. 用户管理模块 (User Management)

提供了用户资源的标准的 CRUD（创建、读取、更新、删除）操作接口，是基础业务功能实现的一个示例。

模型: models.User (包含 ID, Username, Email, CreatedAt 等字段，并使用 Gorm 软删除)

操作: 提供了完整的增删改查 API。

B. 动态 API 服务模块 (Dynamic API Service)

这是项目的核心创新功能。它允许开发者通过 HTTP POST 请求注册一个新的 API 服务，该服务与一个自定义的 SQL 语句绑定。

关键特性:

灵活注册: 可注册 GET, POST, PUT, DELETE 等任意 HTTP 方法的服务。

SQL 绑定: 每个服务都绑定一个带有 ? 占位符的原始 SQL 语句。

参数顺序锁定 (ParamKeys): 使用 JSON 数组定义请求参数（如 URL Query 或 Request Body）的名称，顺序必须与 SQL 语句中的 ? 占位符顺序严格一致。

强制类型转换 (ParamTypes): 使用 JSON 数组定义每个参数的预期 Go 类型（如 "int", "string", "float", "bool"）。系统在执行 SQL 前会对参数进行严格的类型转换，确保 SQL 执行的健壮性。

严格查询限制（安全策略）:
为了保护数据库安全，动态服务仅允许执行只读的查询和诊断操作。任何修改数据库状态的语句（如 INSERT, UPDATE, DELETE, DROP 等）都会被系统拦截并记录日志，而不会执行。
允许的语句前缀 (不区分大小写):

SELECT (标准查询)

WITH (CTE，复杂查询)

EXPLAIN (查询执行计划诊断)

DESCRIBE / DESC (查看表结构)

III. 环境搭建与启动 (Setup and Startup)

项目推荐使用 Docker Compose 进行快速、完整的环境搭建。

A. 依赖 (Prerequisites)

Docker

Docker Compose

B. Docker Compose 启动

配置 .env 文件:
确保项目根目录下的 .env 文件配置正确，特别是数据库的连接信息。

MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_USER=user
MYSQL_PASSWORD=password
MYSQL_DATABASE=godb
MYSQL_ROOT_PASSWORD=rootpassword
PORT=8080


启动服务:
在项目根目录下执行 Docker Compose 命令，它将自动构建 Go 应用镜像、启动 MySQL 容器，并运行应用。

docker-compose up --build -d


验证:
服务启动后，可以通过访问健康检查接口验证应用是否正常运行：

curl http://localhost:8080/
# 预期输出: {"message":"Welcome to Go Gin Gorm API"}

C. 配置项与环境变量（运行时）
D. 使用 `.env` 与 Docker Compose 部署（示例）

1. 复制示例环境文件并编辑实际值：

```bash
cp .env.example .env
# 编辑 .env，填写数据库密码等敏感信息
```

2. 使用 `docker-compose` 启动（推荐，包含构建步骤）：

```bash
# 使用默认 .env 文件
docker-compose --env-file .env up --build -d

# 查看日志，确认服务与数据库启动完成
docker-compose logs -f app

# 停止并移除容器
docker-compose down
```

3. 覆盖环境变量示例：

```bash
# 以临时变量覆盖并启动
DYNAMIC_MAX_ROWS=500 DYNAMIC_QUERY_TIMEOUT_SECONDS=10 docker-compose --env-file .env up --build -d
```

说明:
- `.env.example` 包含运行所需的环境变量示例（包括 `DYNAMIC_MAX_ROWS` 和 `DYNAMIC_QUERY_TIMEOUT_SECONDS` 的默认值）。
- 在生产环境中，尽量通过 CI/CD secret 或 Docker secrets 管理敏感配置，不要把生产密码提交到仓库。


项目的一些运行时行为可通过环境变量配置，推荐在 `docker-compose.yml` 或 `.env` 中设置：

- `DYNAMIC_MAX_ROWS`：执行返回的最大行数，默认值 `1000`。当查询结果超过该值时，API 会截断返回并在响应中标注 `truncated`。示例：`DYNAMIC_MAX_ROWS=500`
- `DYNAMIC_QUERY_TIMEOUT_SECONDS`：动态 SQL 执行的超时时间（秒），默认值 `5`。超过该时间查询将被取消并返回业务码 `2`（超时）。示例：`DYNAMIC_QUERY_TIMEOUT_SECONDS=10`

审计日志说明：动态 SQL 的每次执行都会生成一条持久化审计记录（表名 `audits`），记录 `path`, `method`, `client_ip`, `sql`, `args`, `duration_ms`, `rows`, `truncated` 等字段。审计表由应用在启动时通过 GORM 的 `AutoMigrate` 自动创建。


IV. API 接口参考 (API Reference)

API 基础路径为 /api/v1。

A. 用户接口 (User Endpoints)

方法

路径

描述

请求体示例

响应示例

POST

/api/v1/users

创建新用户

{"username": "test_user", "email": "test@example.com"}

{"code": 0, ...}

GET

/api/v1/users

获取所有用户

(无)

{"code": 0, "data": [...]}

GET

/api/v1/users/:id

根据 ID 获取用户

(无)

{"code": 0, "data": {...}}

PUT

/api/v1/users/:id

更新指定用户

{"username": "new_name"}

{"code": 0, "data": {...}}

DELETE

/api/v1/users/:id

软删除指定用户

(无)

{"code": 0, "message": "删除成功"}

B. 动态服务接口 (Dynamic Service Endpoints)

1. 服务注册 (Registration)

方法

路径

描述

请求体示例

POST

/api/v1/dynamic/register

注册一个新的动态 API 服务

见下文示例

注册请求体示例：查询指定 ID 用户

该服务将注册为 GET /api/v1/dynamic/user_by_id，需要一个名为 id 的整数参数。

{
  "name": "GetUserByIdDynamic",
  "method": "GET",
  "path": "/user_by_id",
  "sql": "SELECT id, username, email FROM users WHERE id = ? AND deleted_at IS NULL",
  "param_keys": "[\"user_id\"]",
  "param_types": "[\"int\"]" 
}


2. 服务执行 (Execution)

注册成功后，通过 GET /api/v1/dynamic/*path 或其他方法路径访问。

示例 1: 执行 GET 查询 (使用 URL Query 参数)

假设已注册上述 GetUserByIdDynamic 服务。

请求: GET http://localhost:8080/api/v1/dynamic/user_by_id?user_id=1

参数: user_id (作为字符串 "1" 传入)

执行: 系统会将其强制转换为 int 类型 (1)，并执行 SQL。

响应: {"code": 0, "message": "查询成功", "data": [...]}

示例 2: 尝试执行被禁用的写入操作

假设尝试注册一个 UPDATE 服务并调用它。

注册 SQL: "UPDATE users SET username = ? WHERE id = ?"

请求: POST http://localhost:8080/api/v1/dynamic/update_user_name

请求体: {"user_id": 2, "new_name": "Alice_Updated"}

响应 (操作被阻止):

{
  "code": 1, 
  "message": "安全限制: 动态服务只允许执行 [SELECT WITH EXPLAIN DESCRIBE DESC ] 查询操作。非查询操作已被阻止。",
  "data": {
    "sql_statement_type": "UPDATE"
  }
}


(注意：系统返回 HTTP 200，但业务代码 code: 1 和明确的消息表明操作已被安全策略拦截，未执行。)
