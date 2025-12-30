# Sonic博客系统 - 快速开始

## 🚀 极简部署（推荐）

### 方式一：一键脚本部署

```bash
# 进入scripts目录
cd scripts

# 运行简化版部署脚本
chmod +x deploy_simple.sh
./deploy_simple.sh
```

### 方式二：手动Docker命令

```bash
# 1. 设置目录
SONIC_DIR=/data/sonic
mkdir -p $SONIC_DIR

# 2. 构建镜像
docker build -f scripts/Dockerfile_simple -t sonic:latest .

# 3. 运行容器
docker run -d \
    --name sonic \
    --network host \
    -e LOGGING_LEVEL_APP=warn \
    -e SQLITE3_ENABLE=true \
    -v $SONIC_DIR:/sonic \
    sonic:latest
```

## 📋 原版部署命令（保持不变）

```bash
SONIC_DIR=/data/sonic

# 创建目录
mkdir -p $SONIC_DIR

docker run -d \
--name sonic \
--network host \
-e LOGGING_LEVEL_APP=warn \
-e SQLITE3_ENABLE=true \
-v $SONIC_DIR:/sonic \
gosonic/sonic:latest
```

## 🔧 环境变量说明

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `LOGGING_LEVEL_APP` | `warn` | 日志级别 |
| `SQLITE3_ENABLE` | `true` | 启用SQLite数据库 |
| `PORT` | `8080` | 服务端口（可选） |

## 📁 数据目录

- **容器内路径**: `/sonic`
- **宿主机路径**: `/data/sonic`（可自定义）
- **包含内容**: 数据库文件、配置文件、上传文件等

## 🌐 访问系统

- **首页**: http://localhost:8080
- **管理后台**: http://localhost:8080/admin
- **API接口**: http://localhost:8080/api

## 🛠️ 常用命令

### 查看状态
```bash
docker ps | grep sonic
```

### 查看日志
```bash
docker logs -f sonic
```

### 停止服务
```bash
docker stop sonic
```

### 重启服务
```bash
docker restart sonic
```

### 删除容器
```bash
docker rm sonic
```

### 进入容器
```bash
docker exec -it sonic sh
```

## 📦 首次使用步骤

1. **启动服务**
   ```bash
   ./deploy_simple.sh
   ```

2. **访问安装页面**
   - 打开浏览器访问: http://localhost:8080/admin/install
   - 创建管理员账户

3. **开始使用**
   - 登录管理后台: http://localhost:8080/admin
   - 发布文章、管理内容

## 🔄 更新部署

```bash
# 停止旧容器
docker stop sonic

# 删除旧容器
docker rm sonic

# 重新构建
docker build -f scripts/Dockerfile_simple -t sonic:latest .

# 启动新容器
docker run -d \
    --name sonic \
    --network host \
    -e LOGGING_LEVEL_APP=warn \
    -e SQLITE3_ENABLE=true \
    -v /data/sonic:/sonic \
    sonic:latest
```

## 🐛 故障排查

### 容器无法启动
```bash
# 查看日志
docker logs sonic

# 检查端口
netstat -tlnp | grep 8080
```

### 数据库问题
```bash
# 进入容器检查
docker exec -it sonic sh

# 查看数据目录
ls -la /sonic/
```

## 💡 提示

1. **数据持久化**: 所有数据都保存在 `/data/sonic` 目录，删除容器后数据不会丢失
2. **网络模式**: 使用 `--network host` 简化网络配置
3. **日志级别**: 可通过环境变量调整日志详细程度
4. **SQLite**: 默认使用SQLite，无需额外数据库服务

---

**注意**: 这是重构后的简化版本，保持了与原版相同的使用方式，但代码更加简洁高效。
