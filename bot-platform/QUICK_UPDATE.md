# 🚀 快速更新指南

## 本地构建并推送（已完成 ✅）

```bash
cd /Users/hovanzhang/git_repo/napcat/bot-platform
./build-and-push.sh
```

**已推送的镜像：**
- `daikonsushi/bot-platform:latest`
- `daikonsushi/bot-platform:dev-20260102-001605`

**支持的平台：**
- linux/amd64 (x86_64)
- linux/arm64 (ARM64)

---

## 远程服务器更新步骤

### 1️⃣ SSH 到远程服务器

```bash
ssh user@your-server
```

### 2️⃣ 进入项目目录

```bash
cd /path/to/your/napcat
```

### 3️⃣ 更新镜像并重启

**使用 Podman Compose:**
```bash
podman-compose pull bot-platform && \
podman-compose up -d bot-platform && \
podman-compose logs -f bot-platform
```

**使用 Docker Compose:**
```bash
docker-compose pull bot-platform && \
docker-compose up -d bot-platform && \
docker-compose logs -f bot-platform
```

### 4️⃣ 验证更新

查看日志确认服务正常启动：
```
bot-platform | Starting bot platform...
bot-platform | Loading plugins...
bot-platform | Server started on :8080
```

### 5️⃣ 测试文件上传功能

在 QQ 中发送：
```
/testfile
```

应该能看到文件上传成功的消息。

---

## 🔧 常用命令

### 查看容器状态
```bash
# Podman
podman-compose ps

# Docker
docker-compose ps
```

### 查看日志
```bash
# Podman
podman-compose logs -f bot-platform

# Docker
docker-compose logs -f bot-platform
```

### 重启服务
```bash
# Podman
podman-compose restart bot-platform

# Docker
docker-compose restart bot-platform
```

### 完全重建
```bash
# Podman
podman-compose down bot-platform
podman-compose pull bot-platform
podman-compose up -d bot-platform

# Docker
docker-compose down bot-platform
docker-compose pull bot-platform
docker-compose up -d bot-platform
```

---

## 📋 更新内容

本次更新添加了以下功能：

### ✨ 新增功能
- **文件上传支持**：插件现在可以通过 gRPC 调用上传文件到 NapCatQQ
- **新的 Proto 定义**：添加了 `UploadFileRequest` 和 `UploadFileResponse`
- **示例插件**：`plugin-filetest` 演示如何使用文件上传功能

### 🔧 技术细节
- 支持多种文件类型（图片、视频、语音等）
- 支持 base64 编码的文件数据
- 支持指定文件名和类型
- 完整的错误处理

### 📦 镜像信息
- **基础镜像**：golang:1.24-bookworm (构建), alpine:3.19 (运行)
- **架构支持**：linux/amd64, linux/arm64
- **镜像大小**：约 20MB (压缩后)

---

## 🆘 故障排查

### 问题：容器无法启动
```bash
# 查看详细日志
podman-compose logs --tail=100 bot-platform

# 检查配置文件
cat bot-platform/config.yaml
```

### 问题：文件上传失败
```bash
# 检查 NapCatQQ 是否运行
podman-compose ps napcat

# 测试网络连接
podman-compose exec bot-platform ping napcat
```

### 问题：镜像拉取失败
```bash
# 手动拉取
podman pull daikonsushi/bot-platform:latest

# 检查登录状态
podman login docker.io
```

---

## 📚 相关文档

- [DOCKER_DEPLOYMENT.md](DOCKER_DEPLOYMENT.md) - 完整部署文档
- [FILE_UPLOAD_REVIEW.md](FILE_UPLOAD_REVIEW.md) - 文件上传功能代码审查
- [examples/plugin-filetest/README.md](examples/plugin-filetest/README.md) - 测试插件文档

---

## 💡 提示

1. **首次部署**：确保 `docker-compose.yaml` 中的镜像名称正确
2. **版本管理**：可以使用特定版本标签而不是 `latest`
3. **备份配置**：更新前备份 `plugins-config` 目录
4. **日志监控**：使用 `-f` 参数实时查看日志
5. **SELinux**：如果使用 Podman 且遇到权限问题，在卷挂载后添加 `:Z` 标志

---

**更新时间**: 2026-01-02 00:16:05  
**镜像版本**: dev-20260102-001605  
**构建工具**: Podman (multi-platform)
