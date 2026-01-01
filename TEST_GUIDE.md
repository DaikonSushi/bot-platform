# 🧪 文件上传功能测试指南

## 📋 测试前准备

### 1. 确认插件已编译 ✅

```bash
ls -lh plugins-bin/
```

应该看到：
- `plugin-filetest` - 文件上传测试插件（14MB）
- `echo-ext-plugin_darwin_arm64` - Echo 插件

### 2. 启动服务

在 `bot-platform` 目录下：

```bash
# 使用 Podman Compose
podman-compose up -d

# 或使用 Docker Compose
docker-compose up -d
```

### 3. 检查服务状态

```bash
# 查看所有容器
podman-compose ps

# 查看 bot-platform 日志
podman-compose logs -f bot-platform

# 查看 napcat 日志
podman-compose logs -f napcat
```

应该看到：
```
bot-platform | Starting bot platform...
bot-platform | Loading plugin: filetest v1.0.0
bot-platform | File test plugin started!
bot-platform | Server started on :8080
```

---

## 🎯 测试场景

### 场景 1: 快速测试（推荐）

这是最简单的测试方式，会自动创建并上传一个测试文件。

**在 QQ 群聊中发送：**
```
/testfile
```

**预期结果：**
1. Bot 回复：`✅ Test file created: /tmp/test_upload.txt`
2. Bot 回复：`Uploading...`
3. Bot 回复：`✅ File uploaded to group successfully!`
4. 群文件中出现 `test_upload.txt`

**在 QQ 私聊中发送：**
```
/testfile
```

**预期结果：**
1. Bot 回复：`✅ Test file created: /tmp/test_upload.txt`
2. Bot 回复：`Uploading...`
3. Bot 回复：`✅ File uploaded to private chat successfully!`
4. 私聊中收到文件

---

### 场景 2: 上传群文件

**命令格式：**
```
/uploadgroup <文件路径> [显示名称] [文件夹]
```

**示例 1: 上传共享目录中的文件**
```
/uploadgroup /shared-data/test.txt
```

**示例 2: 指定显示名称**
```
/uploadgroup /shared-data/test.txt 我的文件.txt
```

**示例 3: 上传到指定文件夹**
```
/uploadgroup /shared-data/test.txt 文档.txt /documents
```

**预期结果：**
- Bot 回复上传进度
- 群文件中出现上传的文件

---

### 场景 3: 上传私聊文件

**命令格式：**
```
/uploadprivate <文件路径> [显示名称]
```

**示例 1: 上传文件**
```
/uploadprivate /shared-data/test.txt
```

**示例 2: 指定显示名称**
```
/uploadprivate /shared-data/test.txt 我的文件.txt
```

**预期结果：**
- Bot 回复上传进度
- 私聊中收到文件

---

## 📁 准备测试文件

### 方法 1: 使用共享目录（推荐）

docker-compose 已经配置了 `/shared-data` 共享目录：

```bash
# 在本地创建测试文件
echo "Hello, this is a test file!" > shared-data/test.txt
echo "测试中文内容" > shared-data/chinese.txt

# 创建一个较大的文件
dd if=/dev/zero of=shared-data/large.bin bs=1M count=10

# 验证文件
ls -lh shared-data/
```

### 方法 2: 在容器内创建

```bash
# 进入 bot-platform 容器
podman-compose exec bot-platform sh

# 创建测试文件
echo "Test from container" > /shared-data/container-test.txt

# 退出容器
exit
```

---

## 🔍 验证测试结果

### 1. 查看 Bot 日志

```bash
podman-compose logs -f bot-platform
```

成功的日志应该包含：
```
[INFO] Uploading file to group: 123456789
[INFO] File uploaded successfully
```

失败的日志可能包含：
```
[ERROR] Failed to upload file: ...
```

### 2. 查看 NapCat 日志

```bash
podman-compose logs -f napcat
```

应该看到 API 调用记录：
```
POST /api/upload_group_file
POST /api/upload_private_file
```

### 3. 在 QQ 中验证

**群文件：**
1. 打开 QQ 群
2. 点击"文件"
3. 查看是否有新上传的文件

**私聊文件：**
1. 打开与 Bot 的私聊
2. 查看聊天记录
3. 应该能看到文件消息

---

## 🐛 故障排查

### 问题 1: Bot 没有响应

**检查步骤：**
```bash
# 1. 检查容器是否运行
podman-compose ps

# 2. 检查 bot-platform 日志
podman-compose logs --tail=50 bot-platform

# 3. 检查 napcat 日志
podman-compose logs --tail=50 napcat

# 4. 重启服务
podman-compose restart bot-platform
```

### 问题 2: 插件未加载

**检查日志：**
```bash
podman-compose logs bot-platform | grep filetest
```

应该看到：
```
Loading plugin: filetest v1.0.0
File test plugin started!
```

如果没有，检查：
```bash
# 1. 插件文件是否存在
ls -l plugins-bin/plugin-filetest

# 2. 插件是否有执行权限
chmod +x plugins-bin/plugin-filetest

# 3. 重启服务
podman-compose restart bot-platform
```

### 问题 3: 文件上传失败

**可能原因：**

1. **文件不存在**
   ```bash
   # 检查文件路径
   podman-compose exec bot-platform ls -l /shared-data/
   ```

2. **NapCat 未连接**
   ```bash
   # 检查网络连接
   podman-compose exec bot-platform ping napcat
   
   # 检查 NapCat 状态
   curl http://localhost:3000/api/status
   ```

3. **权限问题**
   ```bash
   # 检查文件权限
   ls -l shared-data/test.txt
   
   # 修复权限
   chmod 644 shared-data/test.txt
   ```

4. **NapCat API 错误**
   ```bash
   # 查看 NapCat 日志
   podman-compose logs napcat | grep -i error
   ```

### 问题 4: 命令格式错误

**获取帮助：**
```
/uploadgroup
/uploadprivate
```

Bot 会回复正确的命令格式和示例。

---

## 📊 测试检查清单

使用这个清单确保所有功能都测试过：

- [ ] **服务启动**
  - [ ] bot-platform 容器运行正常
  - [ ] napcat 容器运行正常
  - [ ] filetest 插件加载成功

- [ ] **快速测试**
  - [ ] 群聊中 `/testfile` 成功
  - [ ] 私聊中 `/testfile` 成功

- [ ] **群文件上传**
  - [ ] 上传文本文件
  - [ ] 上传中文文件名
  - [ ] 上传到指定文件夹
  - [ ] 指定显示名称

- [ ] **私聊文件上传**
  - [ ] 上传文本文件
  - [ ] 上传中文文件名
  - [ ] 指定显示名称

- [ ] **错误处理**
  - [ ] 文件不存在时的错误提示
  - [ ] 命令格式错误时的帮助信息
  - [ ] 在错误场景使用命令（如群聊用私聊命令）

- [ ] **日志验证**
  - [ ] bot-platform 日志正常
  - [ ] napcat 日志正常
  - [ ] 无错误或警告

---

## 🚀 快速测试脚本

创建一个测试脚本自动化测试：

```bash
#!/bin/bash
# test-file-upload.sh

echo "🧪 Starting file upload test..."
echo ""

# 1. 创建测试文件
echo "📝 Creating test files..."
mkdir -p shared-data
echo "Hello from test script!" > shared-data/auto-test.txt
echo "测试文件" > shared-data/中文测试.txt

# 2. 检查服务状态
echo "🔍 Checking service status..."
podman-compose ps

# 3. 查看插件加载
echo "🔌 Checking plugin loading..."
podman-compose logs bot-platform | grep filetest

# 4. 等待用户测试
echo ""
echo "✅ Setup complete!"
echo ""
echo "📱 Now test in QQ:"
echo "  1. Send: /testfile"
echo "  2. Send: /uploadgroup /shared-data/auto-test.txt"
echo "  3. Send: /uploadprivate /shared-data/中文测试.txt"
echo ""
echo "📊 Watch logs with:"
echo "  podman-compose logs -f bot-platform"
```

保存并运行：
```bash
chmod +x test-file-upload.sh
./test-file-upload.sh
```

---

## 📝 测试报告模板

测试完成后，记录结果：

```
测试日期: 2026-01-02
测试人员: hovanzhang
环境: 本地 Podman

测试结果:
✅ 服务启动正常
✅ 插件加载成功
✅ /testfile 命令 - 群聊
✅ /testfile 命令 - 私聊
✅ /uploadgroup 命令
✅ /uploadprivate 命令
✅ 错误处理正常
✅ 日志输出正常

问题记录:
- 无

备注:
- 所有功能正常工作
- 准备部署到生产环境
```

---

## 🎉 测试成功后

如果所有测试通过：

1. **提交代码**
   ```bash
   git add .
   git commit -m "feat: add file upload functionality with tests"
   git push
   ```

2. **更新远程服务器**
   - 参考 [QUICK_UPDATE.md](QUICK_UPDATE.md)
   - 使用 `podman-compose pull && podman-compose up -d`

3. **监控生产环境**
   ```bash
   ssh user@server
   podman-compose logs -f bot-platform
   ```

---

**祝测试顺利！** 🚀

如有问题，查看：
- [DOCKER_DEPLOYMENT.md](DOCKER_DEPLOYMENT.md) - 部署文档
- [FILE_UPLOAD_REVIEW.md](FILE_UPLOAD_REVIEW.md) - 代码审查
- [examples/plugin-filetest/README.md](examples/plugin-filetest/README.md) - 插件文档
