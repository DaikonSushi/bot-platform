# Plugin Echo External

一个简单的外部 echo 插件，用于演示 bot-platform 的外部插件机制。

## 功能

- `/echo <消息>` - 回显消息 🔊
- `/say <消息>` - 说话 💬
- `/repeat <消息>` - 重复消息两遍 🔁

## 本地编译测试

```bash
# 在本地编译
go build -o echo-ext-plugin .

# 测试插件信息输出
./echo-ext-plugin --info
```

## 发布到 GitHub

1. 创建新的 GitHub 仓库（例如 `plugin-echo-external`）

2. 修改 `go.mod`，将 replace 改为远程依赖：
```go
module github.com/your-username/plugin-echo-external

go 1.21

require bot-platform v0.0.0

// 发布时注释掉这行
// replace bot-platform => ../../
```

3. 推送代码并创建 Release：
```bash
git init
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/your-username/plugin-echo-external.git
git push -u origin main

# 创建标签触发自动构建
git tag v1.0.0
git push origin v1.0.0
```

4. GitHub Actions 会自动构建多平台二进制并附加到 Release

## 在 bot-platform 中使用

```bash
# 安装
./botctl install https://github.com/your-username/plugin-echo-external

# 启动
./botctl start echo-ext

# 停止
./botctl stop echo-ext
```

## License

MIT
