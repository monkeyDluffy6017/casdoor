#!/bin/bash

echo "测试短信验证码服务..."
echo "========================"

# 检查服务是否运行
echo "1. 检查服务健康状态..."
curl -s http://localhost:8083/health | jq '.' 2>/dev/null || curl -s http://localhost:8083/health

echo -e "\n2. 测试JSON格式发送短信..."
curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000","code":"123456"}' | jq '.' 2>/dev/null || curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000","code":"123456"}'

echo -e "\n3. 测试表单格式发送短信..."
curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -d "phone=13900139000&code=654321" | jq '.' 2>/dev/null || curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -d "phone=13900139000&code=654321"

echo -e "\n4. 测试错误情况（缺少手机号）..."
curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -H "Content-Type: application/json" \
  -d '{"code":"123456"}' | jq '.' 2>/dev/null || curl -X POST http://localhost:8083/oidc_auth/send/sms \
  -H "Content-Type: application/json" \
  -d '{"code":"123456"}'

echo -e "\n测试完成！"