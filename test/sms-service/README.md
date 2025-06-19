# 短信验证码服务

这是一个为Casdoor提供短信验证码功能的HTTP服务。

## 功能特性

- 提供POST接口用于发送短信验证码
- 支持JSON和表单数据格式
- 包含健康检查接口
- 详细的日志记录
- CORS支持

## 接口说明

### 1. 发送短信验证码
- **URL**: `POST /oidc_auth/send/sms`
- **端口**: `8083`
- **请求格式**: JSON或表单数据

#### JSON格式请求示例：
```json
{
    "phone": "13800138000",
    "code": "123456"
}
```

#### 表单数据请求示例：
```bash
curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -d "phone=13800138000&code=123456"
```

#### 响应示例：
```json
{
    "success": true,
    "message": "验证码已成功发送到手机号 13800138000",
    "data": {
        "phone": "13800138000",
        "code": "123456",
        "timestamp": 1672531200
    }
}
```

### 2. 健康检查
- **URL**: `GET /health`
- **响应示例**:
```json
{
    "status": "ok",
    "time": "2024-01-01 12:00:00",
    "service": "SMS验证码服务"
}
```

## 启动服务

```bash
cd test/sms-service
go run main.go
```

服务启动后会在控制台显示：
```
短信验证码服务启动，监听端口: :8083
短信发送接口: POST http://localhost:8083/oidc_auth/send/sms
健康检查接口: GET http://localhost:8083/health
```

## 在Casdoor中配置

在Casdoor管理界面中创建SMS提供商时，配置如下：

- **类型**: Custom HTTP SMS
- **地址节点**: `http://localhost:8083/oidc_auth/send/sms`
- **方法**: POST
- **参数**: code (可选配置phone参数)

## 测试

### 使用curl测试：
```bash
# JSON格式
curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000","code":"123456"}'

# 表单格式
curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -d "phone=13800138000&code=123456"

# 健康检查
curl http://localhost:8083/health
```

## 注意事项

1. 这是一个模拟服务，实际发送短信需要集成真实的短信服务商API（如阿里云、腾讯云等）
2. 当前版本仅记录日志，不会真正发送短信
3. 服务支持CORS，可以被前端直接调用
4. 建议在生产环境中添加更多的安全验证和错误处理