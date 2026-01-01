# Docker/Podman 部署更新指南

本指南支持 Docker 和 Podman 两种容器运行时。

## 📦 构建并推送镜像到 DockerHub

### 快速开始

在 bot-platform 目录下运行：

```bash
./build-and-push.sh
```

脚本会自动检测你使用的是 Docker 还是 Podman，并引导你完成以下步骤：
1. 检查容器运行时和登录状态
2. 选择构建类型（单平台或多平台）
3. 输入版本标签
4. 构建并推送镜像

### 构建选项

#### 选项 1: 快速构建（推荐用于测试）
- 只构建当前平台（Mac ARM64）
- 速度快，适合快速迭代
- 如果远程服务器是 x86_64，需要选择选项 2

#### 选项 2: 多平台构建（推荐用于生产）
- 同时构建 linux/amd64 和 linux/arm64
- 兼容性好，适合不同架构的服务器
- 构建时间较长

### 手动构建命令

如果你想手动控制构建过程：

#### 使用 Podman

**单平台构建**
```bash
# 构建
podman build -t daikonsushi/bot-platform:latest .

# 推送
podman push daikonsushi/bot-platform:latest
```

**多平台构建**
```bash
# 创建 manifest
podman manifest create daikonsushi/bot-platform:latest

# 构建 amd64
podman build --platform linux/amd64 \
  --manifest daikonsushi/bot-platform:latest .

# 构建 arm64
podman build --platform linux/arm64 \
  --manifest daikonsushi/bot-platform:latest .

# 推送 manifest
podman manifest push daikonsushi/bot-platform:latest \
  docker://daikonsushi/bot-platform:latest
```

#### 使用 Docker

**单平台构建**
```bash
# 构建
docker build -t daikonsushi/bot-platform:latest .

# 推送
docker push daikonsushi/bot-platform:latest
```

**多平台构建**
```bash
# 创建 buildx builder（首次需要）
docker buildx create --name multiplatform --use
docker buildx inspect --bootstrap

# 构建并推送
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag daikonsushi/bot-platform:latest \
  --tag daikonsushi/bot-platform:v1.0.0 \
  --push \
  .
```

## 🚀 更新远程服务器

### 使用 Podman Compose

SSH 到远程服务器后：

```bash
# 进入项目目录
cd /path/to/your/napcat

# 拉取最新镜像
podman-compose pull bot-platform

# 重启服务
podman-compose up -d bot-platform

# 查看日志
podman-compose logs -f bot-platform
```

**一键更新**
```bash
podman-compose pull bot-platform && \
podman-compose up -d bot-platform && \
podman-compose logs -f bot-platform
```

**完全重建**
```bash
# 停止并删除容器
podman-compose down bot-platform

# 拉取最新镜像
podman-compose pull bot-platform

# 启动新容器
podman-compose up -d bot-platform

# 查看日志
podman-compose logs -f bot-platform
```

### 使用 Docker Compose

SSH 到远程服务器后：

```bash
# 进入项目目录
cd /path/to/your/napcat

# 拉取最新镜像
docker-compose pull bot-platform

# 重启服务
docker-compose up -d bot-platform

# 查看日志
docker-compose logs -f bot-platform
```

**一键更新**
```bash
docker-compose pull bot-platform && \
docker-compose up -d bot-platform && \
docker-compose logs -f bot-platform
```

**完全重建**
```bash
# 停止并删除容器
docker-compose down bot-platform

# 拉取最新镜像
docker-compose pull bot-platform

# 启动新容器
docker-compose up -d bot-platform

# 查看日志
docker-compose logs -f bot-platform
```

### 使用原生 Podman（不使用 compose）

```bash
# 停止并删除旧容器
podman stop bot-platform
podman rm bot-platform

# 拉取最新镜像
podman pull daikonsushi/bot-platform:latest

# 运行新容器（根据你的配置调整参数）
podman run -d \
  --name bot-platform \
  --network napcat_bot-network \
  -v ./bot-platform/config.yaml:/app/config.yaml:ro \
  -v ./bot-platform/plugins-bin:/app/plugins-bin:ro \
  -v ./bot-platform/plugins-config:/app/plugins-config \
  -p 8080:8080 \
  daikonsushi/bot-platform:latest

# 查看日志
podman logs -f bot-platform
```

## 🔍 验证更新

### 1. 检查容器状态

**使用 Podman Compose:**
```bash
podman-compose ps
```

**使用 Docker Compose:**
```bash
docker-compose ps
```

**使用原生 Podman:**
```bash
podman ps | grep bot-platform
```

应该看到 bot-platform 容器状态为 `Up`。

### 2. 查看日志

**使用 Compose:**
```bash
# Podman
podman-compose logs bot-platform

# Docker
docker-compose logs bot-platform
```

**使用原生命令:**
```bash
# Podman
podman logs bot-platform

# Docker
docker logs bot-platform
```

应该看到类似的启动日志：
```
bot-platform | Starting bot platform...
bot-platform | Loading plugins...
bot-platform | Server started on :8080
```

### 3. 测试文件上传功能

在 QQ 中发送命令测试：
```
/testfile
```

如果你部署了 filetest 插件，应该能看到文件上传成功的消息。

## 📝 插件更新

如果你需要更新插件（比如添加文件上传功能到现有插件）：

### 1. 在本地编译插件
```bash
cd /path/to/your/plugin
GOOS=linux GOARCH=amd64 go build -o plugin-name .
```

### 2. 上传到服务器
```bash
scp plugin-name user@server:/path/to/napcat/bot-platform/plugins-bin/
```

### 3. 重启 bot-platform

**使用 Compose:**
```bash
# Podman
podman-compose restart bot-platform

# Docker
docker-compose restart bot-platform
```

**使用原生命令:**
```bash
# Podman
podman restart bot-platform

# Docker
docker restart bot-platform
```

## 🐛 故障排查

### 问题 1: 镜像拉取失败

**使用 Podman:**
```bash
# 检查网络连接
ping docker.io

# 手动拉取镜像
podman pull daikonsushi/bot-platform:latest

# 检查登录状态
podman login docker.io
```

**使用 Docker:**
```bash
# 检查网络连接
ping docker.io

# 手动拉取镜像
docker pull daikonsushi/bot-platform:latest

# 如果还是失败，检查 Docker Hub 状态
```

### 问题 2: 容器启动失败

**使用 Compose:**
```bash
# Podman
podman-compose logs --tail=100 bot-platform

# Docker
docker-compose logs --tail=100 bot-platform
```

**检查配置文件:**
```bash
cat bot-platform/config.yaml
```

**检查卷挂载:**
```bash
# Podman
podman-compose config

# Docker
docker-compose config
```

### 问题 3: 文件上传功能不工作

**检查 NapCatQQ 是否正常运行:**
```bash
# Podman
podman-compose ps napcat

# Docker
docker-compose ps napcat
```

**检查网络连接:**
```bash
# Podman
podman-compose exec bot-platform ping napcat

# Docker
docker-compose exec bot-platform ping napcat
```

**查看 bot-platform 日志:**
```bash
# Podman
podman-compose logs -f bot-platform

# Docker
docker-compose logs -f bot-platform
```

### 问题 4: Podman 特定问题

**SELinux 权限问题:**
```bash
# 如果遇到权限问题，可能需要添加 :Z 标志
# 在 docker-compose.yaml 中修改卷挂载：
volumes:
  - ./bot-platform/config.yaml:/app/config.yaml:ro,Z
  - ./bot-platform/plugins-bin:/app/plugins-bin:ro,Z
```

**Rootless Podman 端口绑定:**
```bash
# 如果使用 rootless podman 且端口 < 1024
# 需要允许绑定低端口
echo "net.ipv4.ip_unprivileged_port_start=80" | sudo tee /etc/sysctl.d/99-podman.conf
sudo sysctl --system
```

## 📊 版本管理

### 查看当前运行的版本
```bash
docker-compose exec bot-platform ./bot --version
```

### 回滚到特定版本
```bash
# 修改 docker-compose.yaml
# 将 image: daikonsushi/bot-platform:latest
# 改为 image: daikonsushi/bot-platform:v1.0.0

# 重启服务
docker-compose up -d bot-platform
```

### 保留多个版本
在推送时使用版本标签：
```bash
docker tag daikonsushi/bot-platform:latest daikonsushi/bot-platform:v1.0.0
docker push daikonsushi/bot-platform:v1.0.0
```

## 🔐 安全建议

1. **不要在 docker-compose.yaml 中硬编码敏感信息**
   - 使用环境变量文件 `.env`
   - 不要提交 `.env` 到 git

2. **定期更新基础镜像**
   ```bash
   docker pull golang:1.24-bookworm
   docker pull alpine:3.19
   ```

3. **限制容器权限**
   - 移除不必要的 `privileged: true`
   - 使用非 root 用户运行

4. **备份重要数据**
   ```bash
   # 备份插件配置
   tar -czf backup-$(date +%Y%m%d).tar.gz \
     bot-platform/plugins-config \
     bot-platform/config.yaml
   ```

## 📚 相关文档

- [FILE_UPLOAD_REVIEW.md](FILE_UPLOAD_REVIEW.md) - 文件上传功能代码审查
- [examples/plugin-filetest/README.md](examples/plugin-filetest/README.md) - 测试插件使用说明
- [Dockerfile](Dockerfile) - Docker 构建配置
- [docker-compose.yaml](docker-compose.yaml) - Docker Compose 配置

## 🆘 获取帮助

如果遇到问题：
1. 查看日志：`docker-compose logs -f bot-platform`
2. 检查容器状态：`docker-compose ps`
3. 进入容器调试：`docker-compose exec bot-platform sh`
4. 查看网络：`docker network inspect napcat_bot-network`
