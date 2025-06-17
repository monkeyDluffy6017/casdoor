#!/bin/bash

# GitHub OAuth 回调服务测试脚本

BASE_URL="http://localhost:8080"

echo "🧪 GitHub OAuth 回调服务测试"
echo ""

echo "1. 测试健康检查接口..."
curl -s "$BASE_URL/health" | jq . || echo "健康检查失败或服务未启动"
echo ""

echo "2. 测试回调接口（无授权码）..."
curl -s "$BASE_URL/auth/github/callback" | jq . || echo "请求失败"
echo ""

echo "3. 测试回调接口（模拟授权码）..."
curl -s -X POST "$BASE_URL/auth/github/callback" \
     -d "code=mock_auth_code_12345" | jq . || echo "请求失败"
echo ""

echo "💡 使用说明："
echo "   - 如果服务未启动，请先运行: ./start.sh 或 go run main.go"
echo "   - 真实测试需要从 GitHub OAuth 流程获取有效的授权码"
echo "   - 示例: curl -X POST '$BASE_URL/auth/github/callback' -d 'code=YOUR_REAL_CODE'"