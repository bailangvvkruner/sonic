#!/bin/bash

# Sonic博客系统重构版 - 一键部署脚本

set -e

echo "=========================================="
echo "  Sonic博客系统 - Docker一键部署"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 函数定义
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker未安装，请先安装Docker"
        print_info "安装指南: https://docs.docker.com/engine/install/"
        exit 1
    fi
    print_success "Docker已安装"
}

# 检查Docker是否可用
check_docker_available() {
    if docker info &> /dev/null; then
        print_success "Docker服务运行正常"
    else
        print_error "Docker服务未运行或权限不足"
        print_info "请确保Docker服务已启动"
        exit 1
    fi
}

# 检查端口占用
check_port() {
    local port=$1
    if netstat -tuln 2>/dev/null | grep ":$port " > /dev/null || ss -tuln 2>/dev/null | grep ":$port " > /dev/null; then
        print_warning "端口 $port 可能被占用，请确认是否有其他服务在运行"
        read -p "是否继续? (y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        print_success "端口 $port 可用"
    fi
}

# 创建必要目录
create_directories() {
    print_info "创建数据目录..."
    
    mkdir -p ../data
    mkdir -p ../logs
    
    # 设置权限（如果需要）
    chmod 755 ../data ../logs 2>/dev/null || true
    
    print_success "目录创建完成"
}

# 检查配置文件
check_config() {
    if [ ! -f "../config_refactored.yaml" ]; then
        print_error "配置文件不存在: config_refactored.yaml"
        exit 1
    fi
    print_success "配置文件检查通过"
}

# 构建镜像
build_image() {
    print_info "构建Docker镜像..."
    
    cd ..
    docker build -f scripts/Dockerfile_refactored -t sonic-blog:latest .
    
    if [ $? -eq 0 ]; then
        print_success "镜像构建成功"
    else
        print_error "镜像构建失败"
        exit 1
    fi
    
    cd scripts
}

# 启动服务
start_service() {
    print_info "启动服务..."
    
    # 停止同名容器（如果存在）
    docker stop sonic-blog 2>/dev/null || true
    docker rm sonic-blog 2>/dev/null || true
    
    # 运行容器
    docker run -d \
        --name sonic-blog \
        -p 8080:8080 \
        -v $(pwd)/../data:/app/data \
        -v $(pwd)/../logs:/app/logs \
        -e TZ=Asia/Shanghai \
        -e PORT=8080 \
        --restart unless-stopped \
        sonic-blog:latest
    
    if [ $? -eq 0 ]; then
        print_success "服务启动成功"
    else
        print_error "服务启动失败"
        exit 1
    fi
}

# 显示欢迎信息
show_welcome() {
    echo ""
    echo "=========================================="
    echo "  部署完成！"
    echo "=========================================="
    echo ""
    echo "访问地址: http://localhost:8080"
    echo "管理后台: http://localhost:8080/admin"
    echo "API接口: http://localhost:8080/api"
    echo ""
    echo "常用命令:"
    echo "  查看日志: docker logs -f sonic-blog"
    echo "  停止服务: docker stop sonic-blog"
    echo "  重启服务: docker restart sonic-blog"
    echo "  删除容器: docker rm sonic-blog"
    echo "  进入容器: docker exec -it sonic-blog sh"
    echo ""
    echo "首次使用:"
    echo "  1. 访问 http://localhost:8080/admin/install"
    echo "  2. 创建管理员账户"
    echo "  3. 开始使用！"
    echo ""
    echo "=========================================="
}

# 主函数
main() {
    echo "开始部署 Sonic博客系统..."
    echo ""
    
    # 检查环境
    check_docker
    check_docker_available
    check_port 8080
    
    # 准备环境
    create_directories
    check_config
    
    # 构建和启动
    build_image
    start_service
    
    # 显示结果
    show_welcome
}

# 运行主函数
main
