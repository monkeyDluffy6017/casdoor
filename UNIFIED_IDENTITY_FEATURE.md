# Casdoor 统一身份管理系统 - 新功能总结

## 🚀 功能概述

本次更新为 Casdoor 新增了完整的**统一身份管理系统**，实现了用户账户合并、多认证方式绑定、以及统一的身份认证机制。这是一个全新的特性，支持用户通过不同的认证方式（GitHub OAuth、手机号、邮箱、自定义OAuth等）登录同一个账户。

## 📋 新增功能清单

### 🆕 新增 API 接口

#### 1. 用户账户合并 API
**POST `/api/identity/merge`**

- **功能**：将两个用户账户合并为一个，保留一个账户，删除另一个账户
- **认证**：需要两个有效的 JWT Token
- **请求体**：
```json
{
    "reserved_user_token": "eyJhbGciOiJSUzI1NiIs...",
    "deleted_user_token": "eyJhbGciOiJSUzI1NiIs..."
}
```
- **响应**：
```json
{
    "status": "ok",
    "universal_id": "90ea5f8b-38f8-452b-b4cf-1cd721a2ce27",
    "deleted_user_id": "550e8400-e29b-41d4-a716-446655440001",
    "merged_auth_methods": [
        {
            "auth_type": "phone",
            "auth_value": "+86138000000"
        },
        {
            "auth_type": "github",
            "auth_value": "123456789"
        }
    ]
}
```

#### 2. 身份信息查询 API
**GET `/api/identity/info`**

- **功能**：查询当前用户绑定的所有认证方式
- **认证**：Bearer Token
- **响应**：
```json
{
    "universal_id": "90ea5f8b-38f8-452b-b4cf-1cd721a2ce27",
    "bound_auth_methods": [
        {
            "auth_type": "github",
            "auth_value": "123456789"
        },
        {
            "auth_type": "phone",
            "auth_value": "+86138000000"
        },
        {
            "auth_type": "email",
            "auth_value": "user@example.com"
        }
    ]
}
```

#### 3. 身份绑定管理 API
**POST `/api/identity/bind`**

- **功能**：为当前用户绑定新的认证方式
- **认证**：Bearer Token
- **请求体**：
```json
{
    "auth_type": "email",
    "auth_value": "newuser@example.com"
}
```

**POST `/api/identity/unbind`**

- **功能**：解绑当前用户的指定认证方式
- **认证**：Bearer Token
- **请求体**：
```json
{
    "auth_type": "phone"
}
```

### 🗄️ 数据库变更

#### 1. User 表扩展
```sql
-- 新增 universal_id 字段
ALTER TABLE user ADD COLUMN universal_id VARCHAR(100) INDEX;
```

#### 2. 新增用户身份绑定表
```sql
CREATE TABLE user_identity_binding (
    id VARCHAR(100) PRIMARY KEY,
    universal_id VARCHAR(100) NOT NULL,
    auth_type VARCHAR(50) NOT NULL,
    auth_value VARCHAR(255) NOT NULL,
    created_time VARCHAR(100) NOT NULL,
    INDEX idx_universal_id (universal_id),
    INDEX idx_auth (auth_type, auth_value),
    UNIQUE KEY unique_auth (auth_type, auth_value)
);
```

**字段说明**：
- `universal_id`：统一身份ID，关联到 User 表的 UniversalId 字段
- `auth_type`：认证类型（github、phone、email、password、custom等）
- `auth_value`：认证值（GitHub ID、手机号、邮箱地址等）

### 🔧 核心功能实现

#### 1. JWT Token 增强
在 JWT Token 中新增字段：
```json
{
    "universal_id": "90ea5f8b-38f8-452b-b4cf-1cd721a2ce27",
    "phone_number": "+86138000000",
    "github_account": "123456789",
    // ... 其他原有字段
}
```

#### 2. 统一身份登录机制
- **新增函数**：`GetUserByFieldWithUnifiedIdentity()`
- **功能**：优先通过身份绑定表查找用户，如果找不到则回退到传统方式
- **影响范围**：所有 OAuth 登录流程（GitHub、Google、微信、自定义等）

#### 3. 用户创建流程增强
- **新增函数**：`createIdentityBindings()`
- **功能**：用户创建时自动建立对应的身份绑定记录
- **支持的认证类型**：
  - `password`：用户名密码
  - `phone`：手机号
  - `email`：邮箱
  - `github`：GitHub OAuth
  - `google`：Google OAuth
  - `wechat`：微信登录
  - `custom`：自定义 OAuth 提供商
  - 等等...

#### 4. 用户合并完整流程
- **身份验证**：验证两个用户的 JWT Token
- **数据转移**：将被删除用户的身份绑定转移到保留用户
- **数据清理**：删除被删除用户的所有相关数据：
  - 用户记录
  - Token 记录
  - Session 记录
  - 验证记录
  - 资源记录
  - 支付记录
  - 交易记录
  - 订阅记录
- **事务安全**：使用数据库事务确保操作原子性


### 🎯 业务场景支持

#### 1. 账户合并场景
```
用户A: GitHub登录 (universal_id_A)
用户B: 手机号登录 (universal_id_B)
↓ 用户发现重复账户，申请合并
调用 /api/identity/merge API
↓ 合并结果
保留用户A，删除用户B
用户A现在可以用 GitHub 或手机号登录
```

#### 2. 多方式登录场景
```
用户注册: GitHub OAuth
绑定手机号: 调用 /api/identity/bind
绑定邮箱: 调用 /api/identity/bind
↓ 用户现在可以通过以下方式登录同一账户：
- GitHub OAuth
- 手机号验证码
- 邮箱验证码
```
