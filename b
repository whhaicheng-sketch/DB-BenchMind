#!/bin/bash
# DB-BenchMind 构建脚本
set -e
cd "$(dirname "$0")"

echo "构建前端..."
cd frontend && npm run build && cd ..

echo "构建后端..."
wails build -platform linux/amd64

echo "✓ 构建完成"
