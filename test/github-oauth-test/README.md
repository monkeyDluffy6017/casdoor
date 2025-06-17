# GitHub OAuth 回调处理服务

这是一个专门用于处理 GitHub OAuth 2.0 授权回调的 Go 微服务。

## 功能特性

- 🔗 专注于 GitHub OAuth 回调处理
- 📱 支持 GET 和 POST 请求
- 🔐 安全的授权码处理
- 📊 JSON 格式的响应数据
- 🔍 详细的日志输出便于调试
- 🏥 健康检查接口

## API 接口

### 1. 健康检查
```
GET /health
```

响应：
```json
{
  "status": "ok",
  "timestamp": 1640995200,
  "service": "github-oauth-callback"
}
```

### 2. GitHub OAuth 回调处理
```
GET|POST /auth/github/callback
```

参数：
- `code` (必需) - GitHub 授权码
- `state` (可选) - 状态参数

成功响应：
```json
{
  "success": true,
  "message": "GitHub OAuth 处理成功",
  "user": {
    "id": 12345,
    "login": "username",
    "name": "User Name",
    "email": "user@example.com",
    "avatar_url": "https://avatars.githubusercontent.com/...",
    "company": "Company Name",
    "location": "Location",
    "bio": "User bio",
    "blog": "https://blog.example.com"
  },
  "access_token": "gho_xxxxxxxxxxxx"
}
```

错误响应：
```json
{
  "success": false,
  "message": "处理失败",
  "error": "错误详情"
}
```

## 使用前准备

### 1. 设置环境变量

```bash
export GITHUB_CLIENT_ID="你的Client ID"
export GITHUB_CLIENT_SECRET="你的Client Secret"
```

或者在 Windows 上：

```cmd
set GITHUB_CLIENT_ID=你的Client ID
set GITHUB_CLIENT_SECRET=你的Client Secret
```

### 2. 创建 GitHub OAuth 应用（可选）

如果你需要创建新的 GitHub OAuth 应用：

1. 访问 [GitHub Developer Settings](https://github.com/settings/developers)
2. 点击 "New OAuth App"
3. 填写应用信息
4. 记录 `Client ID` 和 `Client Secret`

## 运行方式

### 方法一：直接运行

```bash
cd test/github-oauth-test
go mod tidy
go run main.go
```

### 方法二：使用启动脚本

```bash
cd test/github-oauth-test
./start.sh
```

### 方法三：编译后运行

```bash
cd test/github-oauth-test
go build -o github-oauth-test
./github-oauth-test
```

## 测试方法

### 1. 健康检查

```bash
curl http://localhost:8080/health
```

### 2. 测试回调接口

#### 使用 GET 请求
```bash
curl "http://localhost:8080/auth/github/callback?code=YOUR_GITHUB_AUTH_CODE"
```

#### 使用 POST 请求
```bash
curl -X POST "http://localhost:8080/auth/github/callback" \
     -d "code=YOUR_GITHUB_AUTH_CODE"
```

#### 使用 curl 发送表单数据
```bash
curl -X POST "http://localhost:8080/auth/github/callback" \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "code=YOUR_GITHUB_AUTH_CODE&state=optional_state"
```

## 调试信息

服务会在控制台输出详细的日志信息，包括：

- 📨 接收到的回调请求信息（请求方法、授权码等）
- ✅ 授权码和访问令牌获取过程
- 👤 用户信息获取过程
- ❌ 任何错误信息

## 获取的用户信息

- 用户 ID
- 用户名 (login)
- 显示名称 (name)
- 邮箱地址（包括私有邮箱）
- 头像 URL
- 公司信息
- 位置信息
- 个人简介
- 博客链接

## 安全说明

- 支持状态参数验证
- 授权码和访问令牌在日志中会被脱敏显示
- 支持获取用户的私有邮箱地址

## 集成示例

### 在其他应用中调用

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
)

func handleGitHubCallback(code string) error {
    // 构造请求数据
    data := url.Values{}
    data.Set("code", code)

    // 发送请求到回调服务
    resp, err := http.Post(
        "http://localhost:8080/auth/github/callback",
        "application/x-www-form-urlencoded",
        bytes.NewBufferString(data.Encode()),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 解析响应
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return err
    }

    if result["success"].(bool) {
        user := result["user"].(map[string]interface{})
        fmt.Printf("用户登录成功: %s\n", user["login"])
    } else {
        fmt.Printf("登录失败: %s\n", result["error"])
    }

    return nil
}
```

### JavaScript/前端调用

```javascript
async function handleGitHubCallback(code) {
    try {
        const response = await fetch('http://localhost:8080/auth/github/callback', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: `code=${encodeURIComponent(code)}`
        });

        const result = await response.json();

        if (result.success) {
            console.log('用户登录成功:', result.user);
            // 处理用户信息
        } else {
            console.error('登录失败:', result.error);
        }
    } catch (error) {
        console.error('请求失败:', error);
    }
}
```

## 故障排除

### 常见问题

1. **环境变量未设置**
   ```
   请设置 GITHUB_CLIENT_ID 和 GITHUB_CLIENT_SECRET 环境变量
   ```

2. **授权码无效**
   - 确保授权码是从 GitHub OAuth 流程中获取的
   - 授权码只能使用一次，过期后需要重新获取

3. **获取用户邮箱失败**
   - 某些用户可能设置了邮箱隐私保护
   - 服务会尝试获取用户的私有邮箱列表

## 依赖包

- `golang.org/x/oauth2`: OAuth 2.0 客户端实现
- `golang.org/x/oauth2/github`: GitHub OAuth 端点配置

## 参考资料

- [GitHub OAuth Apps 文档](https://docs.github.com/en/developers/apps/building-oauth-apps)
- [GitHub REST API 文档](https://docs.github.com/en/rest)
- [OAuth 2.0 规范](https://tools.ietf.org/html/rfc6749)