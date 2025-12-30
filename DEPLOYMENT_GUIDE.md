# Sonic博客系统 - 完整部署指南

## 🚀 快速部署（推荐）

### 方式一：一键脚本部署

```bash
# 1. 进入scripts目录
cd scripts

# 2. 运行正确部署脚本
chmod +x deploy_correct.sh
./deploy_correct.sh
```

### 方式二：手动部署

#### 步骤1: 构建镜像
```bash
# 在项目根目录执行
docker build -f scripts/Dockerfile_simple -t sonic:latest .
```

#### 步骤2: 创建数据目录
```bash
mkdir -p /data/sonic
```

#### 步骤3: 运行容器
```bash
docker run -d \
    --name sonic \
    --network host \
    -e LOGGING_LEVEL_APP=warn \
    -e SQLITE3_ENABLE=true \
    -v /data/sonic:/sonic \
    sonic:latest
```

## 🔧 原版部署命令（保持兼容）

如果您想使用与原版完全相同的命令格式，可以使用以下方式：

```bash
# 设置环境变量
SONIC_DIR=/data/sonic
mkdir -p $SONIC_DIR

# 运行容器（与原版命令完全一致）
docker run -d \
--name sonic \
--network host \
-e LOGGING_LEVEL_APP=warn \
-e SQLITE3_ENABLE=true \
-v $SONIC_DIR:/sonic \
gosonic/sonic:latest
```

**注意**: 这里使用的是官方镜像 `gosonic/sonic:latest`，而不是我们构建的镜像。

## 📋 部署前的准备

### 1. 环境要求
- Docker 已安装
- 端口 8080 可用
- 至少 512MB 内存

### 2. 项目结构
```
sonic/
├── main_refactored_v2.go    # 主程序（完整Fiber v2特性）
├── config_refactored.yaml   # 配置文件
├── go_refactored.mod        # 依赖文件
├── scripts/
│   ├── Dockerfile_simple    # Docker构建文件
│   ├── deploy_correct.sh    # 正确部署脚本
│   └── deploy_simple.sh     # 简化部署脚本
└── 其他文件...
```

## 🐳 Dockerfile 说明

### 简化版 Dockerfile (`scripts/Dockerfile_simple`)
```dockerfile
FROM golang:alpine as builder
WORKDIR /build
COPY . .
RUN apk add --no-cache git ca-certificates build-base \
    && go mod download \
    && CGO_ENABLED=0 GOOS=linux go build -o sonic main_refactored.go

FROM alpine:latest
RUN apk add --no-cache tzdata ca-certificates
WORKDIR /sonic
COPY --from=builder /build/sonic .
COPY --from=builder /build/config_refactored.yaml config.yaml
RUN mkdir -p data logs
EXPOSE 8080
CMD ["./sonic", "-config", "config.yaml"]
```

### 构建过程
1. **构建阶段**: 使用Go 1.25编译代码
2. **运行阶段**: 使用Alpine Linux，包含时区和证书
3. **复制文件**: 二进制文件 + 配置文件
4. **启动**: 运行sonic服务

## 🌐 访问系统

部署成功后，可以通过以下地址访问：

- **首页**: http://localhost:8080
- **管理后台**: http://localhost:8080/admin
- **API接口**: http://localhost:8080/api
- **安装页面**: http://localhost:8080/admin/install

## 🛠️ 常用管理命令

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

### 查看容器信息
```bash
docker inspect sonic
```

### 检查资源使用
```bash
docker stats sonic
```

## 📊 环境变量说明

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `LOGGING_LEVEL_APP` | `warn` | 日志级别 (debug, info, warn, error) |
| `SQLITE3_ENABLE` | `true` | 启用SQLite数据库 |
| `PORT` | `8080` | 服务端口 |
| `TZ` | `Asia/Shanghai` | 时区设置 |

## 🗄️ 数据持久化

### 数据目录结构
```
/data/sonic/
├── sonic.db          # SQLite数据库文件
├── config.yaml       # 配置文件（可选）
├── data/             # 上传文件目录
└── logs/             # 日志文件目录
```

### 备份数据
```bash
# 备份整个目录
tar -czf sonic-backup-$(date +%Y%m%d).tar.gz /data/sonic

# 只备份数据库
cp /data/sonic/sonic.db ./sonic-db-backup-$(date +%Y%m%d).db
```

### 恢复数据
```bash
# 恢复整个目录
tar -xzf sonic-backup-20240101.tar.gz -C /

# 恢复数据库
cp ./sonic-db-backup-20240101.db /data/sonic/sonic.db
```

## 🔍 故障排查

### 问题1: 容器启动失败
```bash
# 查看详细日志
docker logs sonic

# 常见原因：
# 1. 端口8080被占用
netstat -tlnp | grep 8080

# 2. 数据目录权限问题
ls -la /data/sonic
chmod 755 /data/sonic
```

### 问题2: 数据库连接失败
```bash
# 进入容器检查
docker exec -it sonic sh

# 检查数据库文件
ls -la /sonic/sonic.db

# 检查配置文件
cat /sonic/config.yaml
```

### 问题3: 网络问题
```bash
# 测试容器网络
docker exec -it sonic ping -c 3 baidu.com

# 检查端口映射
docker port sonic
```

## 📝 首次使用流程

1. **启动服务**
   ```bash
   ./deploy_correct.sh
   ```

2. **访问安装页面**
   - 打开浏览器访问: http://localhost:8080/admin/install
   - 填写管理员信息（用户名、密码、邮箱）

3. **登录管理后台**
   - 访问: http://localhost:8080/admin
   - 使用刚才创建的管理员账户登录

4. **开始使用**
   - 发布文章
   - 管理分类和标签
   - 配置系统设置

## 🎯 部署场景

### 场景1: 开发测试
```bash
# 前台运行，查看实时日志
docker run -it --rm \
    --name sonic-dev \
    --network host \
    -e LOGGING_LEVEL_APP=debug \
    -v /data/sonic-dev:/sonic \
    sonic:latest
```

### 场景2: 生产环境
```bash
# 后台运行，资源限制
docker run -d \
    --name sonic \
    --network host \
    --restart unless-stopped \
    --memory=512m \
    --cpus=1.0 \
    -e LOGGING_LEVEL_APP=warn \
    -e SQLITE3_ENABLE=true \
    -v /data/sonic:/sonic \
    sonic:latest
```

### 场景3: 与Nginx配合
```bash
# 使用自定义网络
docker network create sonic-net

# 启动Sonic
docker run -d \
    --name sonic \
    --network sonic-net \
    -p 8080:8080 \
    -v /data/sonic:/sonic \
    sonic:latest

# 启动Nginx
docker run -d \
    --name nginx \
    --network sonic-net \
    -p 80:80 \
    -v /path/to/nginx.conf:/etc/nginx/nginx.conf \
    nginx:latest
```

## 📚 参考资料

- [Fiber官方文档](https://docs.gofiber.io/)
- [Docker官方文档](https://docs.docker.com/)
- [SQLite文档](https://www.sqlite.org/docs.html)

## 💡 提示

1. **首次启动**: 系统会自动创建数据库和必要目录
2. **数据安全**: 定期备份 `/data/sonic` 目录
3. **性能优化**: 生产环境建议使用资源限制
4. **日志管理**: 定期清理日志文件
5. **安全考虑**: 生产环境建议使用SSL证书

---

**注意**: 本部署指南专为重构后的Sonic博客系统设计，保持了与原版相同的使用方式，同时提供了更简化的部署流程。
