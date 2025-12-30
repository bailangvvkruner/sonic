@echo off
REM Sonic博客系统重构版 - Windows一键部署脚本

setlocal enabledelayedexpansion

echo ==========================================
echo   Sonic博客系统 - Docker一键部署
echo ==========================================
echo.

REM 颜色定义 (Windows批处理颜色有限，使用简单提示)
set INFO=[INFO]
set SUCCESS=[SUCCESS]
set WARNING=[WARNING]
set ERROR=[ERROR]

REM 检查Docker是否安装
echo %INFO% 检查Docker环境...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo %ERROR% Docker未安装，请先安装Docker
    echo %INFO% 安装指南: https://docs.docker.com/engine/install/
    pause
    exit /b 1
)
echo %SUCCESS% Docker已安装

REM 检查Docker Compose是否安装
echo %INFO% 检查Docker Compose...
docker compose version >nul 2>&1
if %errorlevel% neq 0 (
    docker-compose --version >nul 2>&1
    if %errorlevel% neq 0 (
        echo %ERROR% Docker Compose未安装，请先安装
        echo %INFO% 安装指南: https://docs.docker.com/compose/install/
        pause
        exit /b 1
    )
)
echo %SUCCESS% Docker Compose已安装

REM 检查端口占用
echo %INFO% 检查端口8080...
netstat -ano | findstr ":8080" >nul 2>&1
if %errorlevel% equ 0 (
    echo %WARNING% 端口8080可能被占用，请确认是否有其他服务在运行
    set /p "continue=是否继续? (y/n): "
    if /i "!continue!" neq "y" (
        exit /b 1
    )
) else (
    echo %SUCCESS% 端口8080可用
)

REM 创建必要目录
echo %INFO% 创建数据目录...
if not exist "..\data" mkdir "..\data"
if not exist "..\logs" mkdir "..\logs"
echo %SUCCESS% 目录创建完成

REM 检查配置文件
echo %INFO% 检查配置文件...
if not exist "..\config_refactored.yaml" (
    echo %ERROR% 配置文件不存在: config_refactored.yaml
    pause
    exit /b 1
)
echo %SUCCESS% 配置文件检查通过

REM 构建镜像
echo %INFO% 构建Docker镜像...
cd ..
docker build -f scripts/Dockerfile_refactored -t sonic-blog:latest .
if %errorlevel% neq 0 (
    echo %ERROR% 镜像构建失败
    cd scripts
    pause
    exit /b 1
)
echo %SUCCESS% 镜像构建成功
cd scripts

REM 启动服务
echo %INFO% 启动服务...
