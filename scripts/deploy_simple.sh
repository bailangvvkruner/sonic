#!/bin/bash

# 简化版部署脚本 - 保持与原来相同的使用方式

SONIC_DIR=/data/sonic

echo "开始部署 Sonic博客系统..."

# 创建目录
echo "创建目录: $SONIC_DIR"
mkdir -p $SONIC_DIR

# 停止并删除旧容器（如果存在）
echo "清理旧容器..."
docker stop sonic 2>/dev/null || true
docker rm sonic 2>/dev/null || true

# 构建镜像
echo "构建镜像..."
docker build -f scripts/Dockerfile_simple -t sonic:latest .

# 运行容器
echo "启动容器..."
docker run -d \
    --name sonic \
    --network host \
    -e LOGGING_LEVEL_APP=warn \
    -e SQLITE3_ENABLE=true \
    -v $SONIC_DIR:/sonic \
    sonic:latest

echo ""
echo "=========================================="
echo "  部署完成！"
echo "=========================================="
echo ""
echo "访问地址: http://localhost:8080"
echo "数据目录: $SONIC_DIR"
echo ""
echo "常用命令:"
echo "  查看日志: docker logs -f sonic"
echo "  停止服务: docker stop sonic"
echo "  重启服务: docker restart sonic"
echo ""
echo "=========================================="
