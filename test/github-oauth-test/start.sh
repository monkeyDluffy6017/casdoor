#!/bin/bash

# GitHub OAuth 测试工具启动脚本

echo "🚀 启动 GitHub OAuth 测试工具"
echo ""

# 检查环境变量
if [ -z "$GITHUB_CLIENT_ID" ]; then
    echo "❌ 请设置 GITHUB_CLIENT_ID 环境变量"
    echo "   export GITHUB_CLIENT_ID=\"你的Client ID\""
    exit 1
fi

if [ -z "$GITHUB_CLIENT_SECRET" ]; then
    echo "❌ 请设置 GITHUB_CLIENT_SECRET 环境变量"
    echo "   export GITHUB_CLIENT_SECRET=\"你的Client Secret\""
    exit 1
fi

echo "✅ 环境变量检查通过"
echo "   GITHUB_CLIENT_ID: $GITHUB_CLIENT_ID"
echo "   GITHUB_CLIENT_SECRET: ${GITHUB_CLIENT_SECRET:0:4}****${GITHUB_CLIENT_SECRET: -4}"
echo ""

# 切换到项目目录
cd "$(dirname "$0")"

# 初始化 Go 模块依赖
echo "📦 安装依赖..."
go mod tidy

# 启动服务
echo ""
echo "🚀 启动服务器..."
go run main.go