# 后端认证系统实现文档

## 概述

本文档描述了Xupu AI小说创作平台的后端用户认证系统实现。

## 实现的功能

### 1. 用户模型
- **文件**: `internal/models/user.go`
- **字段**:
  - 基础信息: ID, 用户名, 邮箱, 密码哈希, 手机, 头像
  - 用户等级: free, vip, svip, admin
  - 账户状态: active, suspended, deleted
  - 时间戳: 创建时间, 更新时间, 最后登录时间

### 2. Token模型
- **文件**: `internal/models/auth_tokens.go`
- **用途**: 存储refresh token和密码重置token
- **类型**: access, refresh, reset

### 3. 认证服务
- **文件**: `internal/services/auth/service.go`
- **功能**:
  - 用户登录 (Login)
  - 用户注册 (Register)
  - Token刷新 (RefreshToken)
  - Token验证 (ValidateToken)
  - 密码重置 (GenerateResetToken, ResetPassword)
  - 资料修改 (UpdateProfile)
  - 修改密码 (ChangePassword)

### 4. JWT服务
- **文件**: `internal/services/auth/jwt.go`
- **功能**:
  - 生成Access Token (24小时有效)
  - 生成Refresh Token (30天有效)
  - 验证Token
  - 生成密码重置Token (1小时有效)

### 5. 密码服务
- **文件**: `internal/services/auth/password.go`
- **功能**:
  - 密码哈希 (bcrypt, cost=12)
  - 密码验证
  - 密码强度验证 (最少6个字符)

### 6. 用户仓储
- **文件**: `internal/repositories/user_repo.go`
- **功能**:
  - CRUD操作
  - 按邮箱/用户名查询
  - 分页列表
  - 存在性检查

### 7. 认证Handler
- **文件**: `internal/handlers/auth.go`
- **接口**:
  - `POST /api/v1/auth/login` - 用户登录
  - `POST /api/v1/auth/register` - 用户注册
  - `POST /api/v1/auth/logout` - 用户登出
  - `POST /api/v1/auth/refresh` - 刷新token
  - `POST /api/v1/auth/forgot-password` - 忘记密码
  - `POST /api/v1/auth/reset-password` - 重置密码
  - `GET /api/v1/users/me` - 获取当前用户
  - `PUT /api/v1/users/me/password` - 修改密码

### 8. 认证中间件
- **文件**: `internal/middleware/auth.go`
- **功能**:
  - JWT认证 (Authenticate)
  - 可选认证 (OptionalAuth)
  - 等级验证 (RequireTier)

## 数据库迁移

### 运行迁移

```bash
# 进入PostgreSQL
psql -U postgres -d xupu

# 执行迁移
\i migrations/000001_create_users.up.sql
```

### 回滚迁移

```bash
psql -U postgres -d xupu
\i migrations/000001_create_users.down.sql
```

## 启动API服务器

### 1. 设置环境变量

```bash
# JWT密钥（生产环境必须修改）
export JWT_SECRET="your-secret-key-change-in-production"

# 数据库连接
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="postgres"
export DB_PASSWORD="your-password"
export DB_NAME="xupu"

# 服务端口
export PORT="8080"
```

### 2. 启动服务器

```bash
cd /home/xlei/project/xupu
go run cmd/api/main.go
```

服务器将在 http://localhost:8080 启动

## API测试

### 1. 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "123456"
  }'
```

**响应**:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "username": "testuser",
      "email": "test@example.com",
      "tier": "free",
      "status": "active",
      "email_verified": false
    },
    "tokens": {
      "access_token": "eyJhbGci...",
      "refresh_token": "eyJhbGci..."
    }
  }
}
```

### 2. 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "123456"
  }'
```

### 3. 获取当前用户

```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 4. 刷新Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### 5. 忘记密码

```bash
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'
```

### 6. 重置密码

```bash
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "RESET_TOKEN_FROM_EMAIL",
    "password": "newpassword123"
  }'
```

## 数据库表结构

### users 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| username | VARCHAR(50) | 用户名（唯一） |
| email | VARCHAR(100) | 邮箱（唯一） |
| password_hash | VARCHAR(255) | 密码哈希 |
| phone | VARCHAR(20) | 手机号 |
| avatar_url | TEXT | 头像URL |
| tier | VARCHAR(20) | 用户等级 |
| tier_expires_at | TIMESTAMP | 等级过期时间 |
| status | VARCHAR(20) | 账户状态 |
| email_verified | BOOLEAN | 邮箱是否验证 |
| last_login_at | TIMESTAMP | 最后登录时间 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### auth_tokens 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | UUID | 主键 |
| user_id | UUID | 用户ID（外键） |
| token | VARCHAR(500) | Token字符串 |
| token_type | VARCHAR(20) | Token类型 |
| expires_at | TIMESTAMP | 过期时间 |
| revoked_at | TIMESTAMP | 撤销时间 |
| revoked | BOOLEAN | 是否已撤销 |
| created_at | TIMESTAMP | 创建时间 |

## 安全注意事项

### 1. JWT密钥
- **必须**在生产环境中设置强随机密钥
- 使用环境变量 `JWT_SECRET` 设置
- 建议长度至少32个字符

### 2. 密码哈希
- 使用 bcrypt 算法
- cost factor = 12
- 密码最少6个字符

### 3. Token有效期
- Access Token: 24小时
- Refresh Token: 30天
- 密码重置Token: 1小时

### 4. CORS
- 已启用CORS中间件
- 允许跨域请求

## 前端集成

前端代码已经在 `web/src/services/authApi.ts` 中实现了API调用：

```typescript
// 登录
authApi.login({ email, password })

// 注册
authApi.register({ username, email, password })

// 刷新token
authApi.refreshToken(refreshToken)

// 获取当前用户
authApi.getCurrentUser()

// 修改密码
authApi.changePassword(oldPassword, newPassword)

// 忘记密码
authApi.forgotPassword(email)

// 重置密码
authApi.resetPassword(token, newPassword)
```

## 常见问题

### Q: 如何修改JWT密钥？
A: 设置环境变量 `JWT_SECRET`

### Q: 如何添加新的用户等级？
A: 修改 `models/user.go` 中的Tier约束，并在middleware中更新等级映射

### Q: 如何自定义Token有效期？
A: 修改 `services/auth/jwt.go` 中的 `accessExpiry` 和 `refreshExpiry`

### Q: 密码重置邮件如何发送？
A: 目前只是生成了token，需要集成邮件服务（如SendGrid）来实际发送邮件

## 下一步

- [ ] 集成邮件服务发送密码重置邮件
- [ ] 实现邮箱验证功能
- [ ] 添加OAuth2.0登录（Google, GitHub等）
- [ ] 实现Token黑名单机制
- [ ] 添加登录日志和审计功能
- [ ] 实现用户头像上传
- [ ] 添加手机验证功能
