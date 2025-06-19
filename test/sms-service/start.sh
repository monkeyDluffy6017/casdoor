#!/bin/bash

echo "启动短信验证码服务..."
echo "服务地址: http://localhost:8083"
echo "接口地址: http://localhost:8083/oidc_auth/send/sms"
echo "健康检查: http://localhost:8083/health"
echo "按 Ctrl+C 停止服务"
echo "========================"

go run main.go