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

示例 2: 执行 POST 写入操作 (使用 JSON Body 参数)

假设注册了一个更新用户名的服务：

{
  "name": "UpdateUsernameDynamic",
  "method": "POST",
  "path": "/update_user_name",
  "sql": "UPDATE users SET username = ? WHERE id = ?",
  "param_keys": "[\"new_name\", \"user_id\"]",
  "param_types": "[\"string\", \"int\"]" 
}


请求: POST http://localhost:8080/api/v1/dynamic/update_user_name

请求体: {"user_id": 2, "new_name": "Alice_Updated"}

注意: 即使请求体中的顺序是 user_id 在前，但系统会严格按照 ParamKeys ("new_name", "user_id") 的顺序将参数 Alice_Updated 和 2 绑定到 SQL 的 ? 占位符上。

响应: {"code": 0, "message": "SQL 执行成功，影响行数: 1", "data": {"rows_affected": 1}} (因为是非查询操作)
