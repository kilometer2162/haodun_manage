#!/usr/bin/env bash
set -e

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
DIST_DIR="$PROJECT_ROOT/dist"

echo "====== 开始打包 ======"

# 1. 清理旧产物
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# 2. 编译后端
echo ">>> 编译后端 ..."
cd "$BACKEND_DIR"
go mod tidy
GOOS=linux GOARCH=amd64 go build -o "$DIST_DIR/backend" .

# 3. 构建前端
echo ">>> 构建前端 ..."
cd "$FRONTEND_DIR"
npm ci
npm run build
cp -r dist/* "$DIST_DIR/"

# 4. 复制静态资源（可选）
echo ">>> 复制静态资源 ..."
if [ -d "$BACKEND_DIR/uploads" ]; then
  cp -r "$BACKEND_DIR/uploads" "$DIST_DIR/"
fi
if [ -d "$BACKEND_DIR/template" ]; then
  cp -r "$BACKEND_DIR/template" "$DIST_DIR/"
fi

echo "====== 打包完成 ======"
echo "产物目录：$DIST_DIR"