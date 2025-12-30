#!/bin/bash

# 正确的部署脚本 - 基于原版命令格式

echo "=== Sonic博客系统部署 ==="

# 1. 构建镜像
echo "步骤1: 构建Docker镜像..."
docker build -f scripts/Dockerfile_simple -t sonic:latest .

if [ $? -ne 0 ]; then
    echo "❌ 镜像构建失败"
    exit 1
fi

echo "✅ 镜像构建成功"

# 2. 设置数据目录
SONIC_DIR=/data/sonic
echo "步骤2: 设置数据目录: $SONIC_DIR"
mkdir -p $SONIC_DIR

# 3. 停止并删除旧容器
echo "步骤3: 清理旧容器..."
docker stop sonic 2>/dev/null || true
docker rm sonic 2>/dev/null || true

# 4. 运行容器（使用原版命令格式）
echo "步骤4: 启动容器..."
docker run -d \
    --name sonic \
    --network host \
    -e LOGGING_LEVEL_APP=warn \
    -e SQLITE3_ENABLE=true \
    -v $SONIC_DIR:/sonic \
    sonic:latest

if [ $? -eq 0 ]; then
    echo "✅ 容器启动成功"
    echo ""
    echo "=== 部署完成 ==="
    echo "访问地址: http://localhost:8080"
    echo "数据目录: $SONIC_DIR"
    echo ""
    echo "常用命令:"
    echo "  查看日志: docker logs -f sonic"
    echo "  停止服务: docker stop sonic"
    echo "  重启服务: docker restart sonic"
else
    echo "❌ 容器启动失败"
    echo "请检查端口8080是否被占用"
fi
