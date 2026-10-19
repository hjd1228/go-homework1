# 个人博客系统后端（Gin + GORM + JWT）

一个用 Go 实现的博客后端，提供用户认证、文章 CRUD、评论管理能力。

## 功能

- **用户认证**：注册（密码 bcrypt 加密存储）、登录，登录后签发 JWT
- **文章管理**：列表（分页）、详情、创建、更新、删除；写操作需登录，且只有作者本人能改删
- **评论**：对文章发表评论、按文章查询评论列表
- **统一错误处理**：所有错误返回 `{"error": "..."}`，并带有语义正确的 HTTP 状态码
- **日志**：请求日志（方法/路径/状态码/耗时）+ GORM 慢查询与错误日志

## 运行环境

| 项 | 要求 |
|---|---|
| Go | 1.20 及以上（`go.mod` 声明 `go 1.20`） |
| 数据库 | SQLite（默认，零配置）；也支持换成 MySQL |
| 依赖 | 见 `go.mod`，由 `go mod download` 自动拉取 |

主要依赖：`gin-gonic/gin` v1.10.0、`gorm.io/gorm` v1.31.2、`gorm.io/driver/sqlite` v1.6.0、
`golang-jwt/jwt/v5` v5.2.1、`golang.org/x/crypto` v0.23.0（bcrypt）。

## 目录结构

```
blog/
├── main.go                       入口：加载配置、连库、迁移、启动 HTTP
├── go.mod / go.sum
├── scripts/smoke_test.sh         接口冒烟测试脚本
├── api.http                      REST Client / Postman 风格的用例集合
└── internal/
    ├── config/config.go          环境变量配置
    ├── database/database.go      数据库连接与日志配置
    ├── models/models.go          User / Post / Comment 模型
    ├── auth/jwt.go               JWT 签发与校验
    ├── middleware/
    │   ├── logger.go             请求日志
    │   └── auth.go               JWT 认证中间件
    ├── response/response.go      统一错误响应
    ├── handlers/                 auth / post / comment 处理函数
    └── router/router.go          路由注册
```

## 快速开始

```bash
cd blog
go mod download
go run .
```

默认监听 `:8080`，数据库文件为当前目录下的 `blog.db`。

### 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `JWT_SECRET` | 随机生成 | JWT 签名密钥。**未设置时每次启动随机生成**，服务重启后旧 token 失效；生产环境必须显式设置 |
| `BLOG_ADDR` | `:8080` | 监听地址 |
| `BLOG_DB` | `blog.db` | SQLite 文件路径 |

```bash
JWT_SECRET=your-strong-secret BLOG_ADDR=:9000 go run .
```

### 切换到 MySQL

```go
// internal/database/database.go
import "gorm.io/driver/mysql"

func Connect(dsn string) (*gorm.DB, error) {
	// ...
	return gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger})
}
```

DSN 形如 `user:pass@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=True`。
模型、查询、钩子代码无需改动。

## 接口文档

统一约定：请求与响应均为 JSON；需要认证的接口通过 `Authorization: Bearer <token>` 传 JWT。

| 方法 | 路径 | 认证 | 说明 |
|---|---|---|---|
| POST | `/api/register` | 否 | 注册。body：`username`(3-32)、`password`(6-64)、`email` |
| POST | `/api/login` | 否 | 登录，返回 `token` 与 `expires_in`（秒） |
| GET | `/api/posts` | 否 | 文章列表，支持 `page`、`page_size`（默认 1 / 10，上限 100） |
| GET | `/api/posts/:id` | 否 | 文章详情（含作者信息） |
| POST | `/api/posts` | 是 | 创建文章。body：`title`、`content` |
| PUT | `/api/posts/:id` | 是 | 更新文章，仅作者本人 |
| DELETE | `/api/posts/:id` | 是 | 删除文章（连带删除其评论），仅作者本人 |
| POST | `/api/posts/:id/comments` | 是 | 发表评论。body：`content` |
| GET | `/api/posts/:id/comments` | 否 | 某篇文章的评论列表 |

### 状态码约定

| 码 | 含义 |
|---|---|
| 200 / 201 | 成功 / 创建成功 |
| 400 | 请求参数不合法（含校验失败） |
| 401 | 未认证：缺少或格式错误的 Authorization 头、token 无效或过期、用户名密码错误 |
| 403 | 已认证但无权操作（非文章作者） |
| 404 | 文章或接口不存在 |
| 409 | 用户名或邮箱已被注册 |
| 500 | 服务端错误（细节只进日志，不返回给调用方） |

### 示例

```bash
# 注册
curl -X POST http://127.0.0.1:8080/api/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123","email":"alice@example.com"}'

# 登录并取 token
TOKEN=$(curl -s -X POST http://127.0.0.1:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}' \
  | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

# 创建文章
curl -X POST http://127.0.0.1:8080/api/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"第一篇文章","content":"Gin + GORM + JWT"}'
```

## 接口测试

两种方式，任选其一。

### 方式一：一键冒烟测试（推荐）

脚本会自动编译并启动服务（端口 18080、临时数据库），跑完全部用例后关闭服务，
不影响你手动启动的实例。

```bash
cd blog
bash scripts/smoke_test.sh
```

### 方式二：api.http

用 VS Code 的 REST Client 扩展或 Postman 导入 [api.http](api.http) 逐个执行。

### 测试结果

`bash scripts/smoke_test.sh` 实际运行结果（21 个用例，全部符合预期）：

| # | 用例 | 期望 | 实际 | 结论 |
|---|---|---|---|---|
| 1 | 注册用户 A | 201 | 201 | ✅ |
| 2 | 重复注册同一用户名 | 409 | 409 | ✅ |
| 3 | 注册密码过短 | 400 | 400 | ✅ |
| 4 | 登录成功并取得 token | 200 | 200 | ✅ |
| 5 | 登录密码错误 | 401 | 401 | ✅ |
| 6 | 未携带 token 创建文章 | 401 | 401 | ✅ |
| 7 | token 伪造 | 401 | 401 | ✅ |
| 8 | 用户 A 创建文章 | 201 | 201 | ✅ |
| 9 | 获取文章列表 | 200 | 200 | ✅ |
| 10 | 获取文章详情 | 200 | 200 | ✅ |
| 11 | 作者本人更新文章 | 200 | 200 | ✅ |
| 12 | 注册用户 B | 201 | 201 | ✅ |
| 12b | 用户 B 登录 | 200 | 200 | ✅ |
| 13 | 用户 B 更新他人文章 | 403 | 403 | ✅ |
| 14 | 用户 B 删除他人文章 | 403 | 403 | ✅ |
| 15 | 用户 B 发表评论 | 201 | 201 | ✅ |
| 16 | 获取评论列表 | 200 | 200 | ✅ |
| 17 | 对不存在的文章发评论 | 404 | 404 | ✅ |
| 18 | 获取不存在的文章 | 404 | 404 | ✅ |
| 19 | 作者本人删除文章 | 200 | 200 | ✅ |
| 20 | 删除后再查 | 404 | 404 | ✅ |

## 设计说明

### 与作业示例的两处有意偏离

1. **JWT 库用 `golang-jwt/jwt/v5` 而非示例中的 `dgrijalva/jwt-go`**。
   后者已于 2020 年归档不再维护，且存在 CVE-2020-26160（`aud` 声明校验缺陷）。
   `golang-jwt/jwt` 是社区接手的官方延续版本，API 基本一致。
2. **模型不内嵌 `gorm.Model`，改为显式字段**。
   `gorm.Model` 会带出 `DeletedAt` 开启软删除，并把该字段暴露到 JSON 里，
   与作业要求的字段清单（id/title/content/user_id/created_at/updated_at）也不完全对应。

### 安全相关的处理

- 密码用 **bcrypt** 加密存储；`User.Password` 带 `json:"-"`，任何接口都不会泄露密码哈希。
- 登录失败时**不区分**"用户不存在"与"密码错误"，统一返回 401，避免账号枚举。
- JWT 校验时**强制检查签名算法**必须为 HMAC，防止把 `alg` 改成 `none` 或非对称算法绕过签名。
- 密钥不硬编码：从 `JWT_SECRET` 读取，未设置则随机生成并打印警告。
- 所有数据库查询均为参数化查询（GORM `?` 占位符），无 SQL 拼接。

### 其他工程细节

- **`Preload("Author")`**：一对多/多对一关联用一条额外的 `IN` 查询批量加载，
  避免 N+1（列表接口从 `1+N` 次查询降到 2 次）。
- **更新文章时 `Omit("Author")`**：`findPost` 未预加载作者，`post.Author` 是零值，
  若直接 `Save` 会让 GORM 连带 upsert 一条 id 为 0 的假用户。
- **删除文章用事务**：先删评论再删文章，避免 SQLite 未开启外键约束时留下孤儿评论。
- **GORM 日志降噪**：默认日志会把 `record not found` 当错误打印，
  而"查不到文章"是正常的 404 分支，因此改为只记录慢查询与真实错误。

## 已知限制

- 未实现文章的分页总数缓存、标签、点赞等功能。
- JWT 无刷新机制，过期后需重新登录。
- 未编写单元测试，接口正确性依赖 `scripts/smoke_test.sh` 的端到端验证。
