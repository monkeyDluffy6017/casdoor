#!/bin/bash

echo "🔧 GitHub OAuth 问题诊断脚本"
echo "============================="

# 检查环境变量
echo "📋 当前环境变量:"
echo "GITHUB_CLIENT_ID: ${GITHUB_CLIENT_ID}"
if [ -n "${GITHUB_CLIENT_SECRET}" ]; then
    SECRET_LEN=${#GITHUB_CLIENT_SECRET}
    echo "GITHUB_CLIENT_SECRET: ${GITHUB_CLIENT_SECRET:0:4}...${GITHUB_CLIENT_SECRET: -4} (长度: $SECRET_LEN)"
else
    echo "GITHUB_CLIENT_SECRET: 未设置"
fi
echo ""

# 检查网络连接
echo "🌐 测试网络连接:"
if curl -s --max-time 10 https://api.github.com/user > /dev/null; then
    echo "✅ GitHub API 连接正常"
else
    echo "❌ GitHub API 连接失败"
fi

if curl -s --max-time 10 https://github.com > /dev/null; then
    echo "✅ GitHub 主站连接正常"
else
    echo "❌ GitHub 主站连接失败"
fi
echo ""

# 运行Go调试程序
echo "🐛 运行详细诊断:"
go run debug_test.go -c 'RunDebugTests()'

echo ""
echo "🔍 额外检查项目:"

# 检查端口占用
echo "1. 检查端口占用情况:"
if command -v lsof > /dev/null; then
    echo "   端口8000: $(lsof -ti:8000 | wc -l) 个进程"
    echo "   端口8080: $(lsof -ti:8080 | wc -l) 个进程"
else
    echo "   lsof 命令不可用，跳过端口检查"
fi

# 检查系统时间
echo "2. 系统时间: $(date)"

# 检查DNS解析
echo "3. DNS解析测试:"
if command -v dig > /dev/null; then
    DIG_RESULT=$(dig +short api.github.com)
    if [ -n "$DIG_RESULT" ]; then
        echo "   ✅ api.github.com 解析正常: $DIG_RESULT"
    else
        echo "   ❌ api.github.com 解析失败"
    fi
else
    echo "   dig 命令不可用，跳过DNS检查"
fi

echo ""
echo "📝 解决建议:"
echo "1. 如果所有检查都通过，问题可能是授权码过期"
echo "2. 尝试清除浏览器缓存并重新授权"
echo "3. 确保GitHub OAuth应用状态为激活状态"
echo "4. 如果网络有问题，检查防火墙和代理设置"