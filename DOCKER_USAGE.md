# Docker部署说明 - 重构版Sonic博客

## 🐳 Docker部署优势

- **环境隔离**: 完全隔离的运行环境
- **一键部署**: 简单的命令即可启动
- **资源限制**: 可以控制内存和CPU使用
- **持久化**: 数据持久化存储

## 📦 快速开始

### 1. 使用Dockerfile构建

```bash
# 进入项目根目录
cd /path/to/sonic

# 构建镜像
docker build -f scripts/Dockerfile_refactored -t sonic-blog:latest .

# 运行容器
docker run -d \
  --name sonic-blog \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -e TZ=Asia/Shanghai \
  sonic-blog:latest
```

### 2. 使用一键部署脚本（推荐）

```bash
# 进入scripts目录
cd scripts

# 运行部署脚本（Linux/Mac）
chmod +x deploy.sh
./deploy.sh

# Windows用户可以使用PowerShell或Git Bash运行
```

### 3. 手动运行容器

```bash
# 创建数据目录
mkdir -p data logs

# 运行容器
docker run -d \
  --name sonic-blog \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -e TZ=Asia/Shanghai \
  -e PORT=8080 \
  --restart unless-stopped \
  sonic-blog:latest
```

## 📁 目录结构

```
sonic/
├── main_refactored.go          # 主程序
├── config_refactored.yaml      # 配置文件
├── go_refactored.mod           # 依赖文件
├── scripts/
│   ├── Dockerfile_refactored   # Docker构建文件
│   ├── deploy.sh               # 一键部署脚本
│   └── DOCKER_USAGE.md         # Docker使用说明
├── data/                       # 数据持久化目录（运行时创建）
└── logs/                       # 日志目录（运行时创建）
```

## 🔧 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `TZ` | `Asia/Shanghai` | 时区设置 |
| `PORT` | `8080` | 服务端口 |

## 🗄️ 数据持久化

### 数据目录
- `/app/data` - 数据库文件存储位置
- `/app/logs` - 日志文件存储位置

### 备份数据
```bash
# 备份数据库
docker cp sonic-blog:/app/data/sonic.db ./backup/

# 备份配置
docker cp sonic-blog:/app/conf/config.yaml ./backup/
```

### 恢复数据
```bash
# 恢复数据库
docker cp ./backup/sonic.db sonic-blog:/app/data/

# 重启容器生效
docker restart sonic-blog
```

## 🚀 部署示例

### 基础部署（推荐使用一键脚本）
```bash
# 进入scripts目录并运行部署脚本
cd scripts
./deploy.sh
```

### 手动部署
```bash
# 1. 创建必要目录
mkdir -p data logs

# 2. 构建镜像
docker build -f scripts/Dockerfile_refactored -t sonic-blog:latest .

# 3. 运行容器
docker run -d \
  --name sonic-blog \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -e TZ=Asia/Shanghai \
  -e PORT=8080 \
  --restart unless-stopped \
  sonic-blog:latest

# 4. 查看状态
docker ps | grep sonic-blog
```

### 生产环境部署
```bash
# 1. 构建镜像
docker build -f scripts/Dockerfile_refactored -t sonic-blog:latest .

# 2. 运行容器（带资源限制）
docker run -d \
  --name sonic-blog \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -e TZ=Asia/Shanghai \
  -e PORT=8080 \
  --restart unless-stopped \
  --memory=512m \
  --cpus=1.0 \
  sonic-blog:latest

# 3. 查看日志
docker logs -f sonic-blog
```

### 开发环境调试
```bash
# 前台运行查看日志
docker run -it \
  --name sonic-blog-dev \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -e TZ=Asia/Shanghai \
  -e PORT=8080 \
  sonic-blog:latest

# 重建镜像（代码更新后）
docker stop sonic-blog
docker rm sonic-blog
docker build -f scripts/Dockerfile_refactored -t sonic-blog:latest .
docker run -d [上述参数...]
```

## 🔍 常用命令

### 容器管理
```bash
# 查看运行状态
docker ps | grep sonic-blog

# 查看容器日志
docker logs -f sonic-blog

# 进入容器终端
docker exec -it sonic-blog sh

# 重启容器
docker restart sonic-blog

# 停止容器
docker stop sonic-blog

# 删除容器
docker rm sonic-blog

# 强制删除容器（运行中也可删除）
docker rm -f sonic-blog
```

### 镜像管理
```bash
# 查看镜像
docker images | grep sonic-blog

# 删除镜像
docker rmi sonic-blog:latest

# 重新构建
docker build -f scripts/Dockerfile_refactored -t sonic-blog:latest .
```

### 一键更新
```bash
# 停止并删除旧容器
docker stop sonic-blog 2>/dev/null || true
docker rm sonic-blog 2>/dev/null || true

# 重新构建镜像
docker build -f scripts/Dockerfile_refactored -t sonic-blog:latest .

# 启动新容器
docker run -d \
  --name sonic-blog \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -e TZ=Asia/Shanghai \
  -e PORT=8080 \
  --restart unless-stopped \
  sonic-blog:latest
```

## 📊 资源监控

### 查看资源使用
```bash
# 查看CPU和内存使用
docker stats sonic-blog

# 查看详细信息
docker inspect sonic-blog
```

### 设置资源限制
编辑 `docker-compose-refactored.yml`：
```yaml
deploy:
  resources:
    limits:
      cpus: '1.0'      # 最多使用1个CPU核心
      memory: 512M     # 最多使用512MB内存
    reservations:
      cpus: '0.5'      # 保证0.5个CPU核心
      memory: 128M     # 保证128MB内存
```

## 🔄 更新部署

### 代码更新后
```bash
# 1. 停止服务
docker-compose -f scripts/docker-compose-refactored.yml down

# 2. 重新构建
docker-compose -f scripts/docker-compose-refactored.yml build --no-cache

# 3. 启动服务
docker-compose -f scripts/docker-compose-refactored.yml up -d
```

### 配置更新后
```bash
# 更新配置文件
vim config_refactored.yaml

# 重启容器
docker restart sonic-blog
```

## 🐛 故障排查

### 容器无法启动
```bash
# 查看启动日志
docker logs sonic-blog

# 检查端口占用
netstat -tlnp | grep 8080

# 检查文件权限
ls -la data/ logs/
```

### 数据库问题
```bash
# 进入容器检查
docker exec -it sonic-blog sh

# 检查数据库文件
ls -la /app/data/

# 查看数据库大小
du -h /app/data/sonic.db
```

### 网络问题
```bash
# 测试容器网络
docker exec -it sonic-blog ping -c 3 baidu.com

# 检查端口映射
docker port sonic-blog
```

## 📝 部署检查清单

- [ ] 创建数据目录 `data/`
- [ ] 创建日志目录 `logs/`
- [ ] 配置文件已准备
- [ ] 端口8080未被占用
- [ ] Docker环境已安装
- [ ] 防火墙已配置（如需要）
- [ ] 域名解析已配置（如需要）
- [ ] SSL证书已配置（如需要）

## 🎯 部署场景

### 场景1: 个人博客
```bash
# 简单部署，使用默认配置
docker-compose -f scripts/docker-compose-refactored.yml up -d
```

### 场景2: 多环境部署
```bash
# 开发环境
cp config_refactored.yaml config.dev.yaml
# 修改配置...
docker build -f scripts/Dockerfile_refactored -t sonic-blog:dev .
docker run -d --name sonic-blog-dev -p 8081:8080 -v $(pwd)/data-dev:/app/data sonic-blog:dev

# 生产环境
cp config_refactored.yaml config.prod.yaml
# 修改配置...
docker build -f scripts/Dockerfile_refactored -t sonic-blog:prod .
docker run -d --name sonic-blog-prod -p 8080:8080 -v $(pwd)/data-prod:/app/data sonic-blog:prod
```

### 场景3: 反向代理部署
```bash
# 配置Nginx反向代理
# nginx.conf 示例:
# server {
#     listen 80;
#     server_name your-domain.com;
#     location / {
#         proxy_pass http://localhost:8080;
#         proxy_set_header Host $host;
#         proxy_set_header X-Real-IP $remote_addr;
#     }
# }
```

## 📚 参考资料

- [Docker官方文档](https://docs.docker.com/)
- [Docker Compose官方文档](https://docs.docker.com/compose/)
- [Sonic博客系统文档](README_REFACTORED.md)

## 💡 提示

1. **首次启动**: 第一次启动时，系统会自动创建数据库和必要的目录
2. **数据安全**: 定期备份 `data/` 目录
3. **性能优化**: 根据实际需求调整Docker资源限制
4. **日志管理**: 定期清理日志文件，避免磁盘空间不足
5. **安全考虑**: 生产环境建议使用SSL证书和防火墙

---

**注意**: 本Docker配置专为重构后的Sonic博客系统设计，相比原版更加轻量和高效。
