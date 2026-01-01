# Bot Platform

基于 NapCat 的 QQ 机器人消息处理平台，支持插件热加载、Docker 部署、K8s 友好。

## 架构特点

- **插件解耦**：插件编译为独立二进制，通过 gRPC 与主平台通信
- **热加载**：运行时安装、启动、停止插件，无需重启主程序
- **GitHub 集成**：直接从 GitHub Releases 下载安装插件
- **容器友好**：支持 Docker 部署，适合 K8s 编排

## 快速开始

### 1. 编译

```bash
# 安装依赖
make deps

# 生成 protobuf 代码并编译
make all
```

### 2. 配置

编辑 `config.yaml`:

```yaml
napcat:
  http_url: "http://127.0.0.1:3000"
  ws_url: "ws://127.0.0.1:3001"
  token: ""

bot:
  admins:
    - 123456789   # 管理员 QQ 号
  command_prefix: "/"
  debug: false

plugin_manager:
  enabled: true
  plugin_dir: "./plugins-bin"
  config_dir: "./plugins-config"
  grpc_port: 50051
  auto_start: []  # 启动时自动加载的插件

admin_server:
  enabled: true
  addr: ":8080"
```

### 3. 运行

```bash
./bot
```

## 插件管理

### 使用命令行工具 (botctl)

```bash
# 查看已安装插件
./botctl list

# 从 GitHub 安装插件
./botctl install https://github.com/user/plugin-weather

# 启动插件
./botctl start weather

# 停止插件
./botctl stop weather

# 卸载插件
./botctl uninstall weather

# 检查平台状态
./botctl health
```

### 使用 HTTP API

```bash
# 列出所有插件
curl http://127.0.0.1:8080/api/plugins

# 安装插件
curl -X POST http://127.0.0.1:8080/api/plugins/install \
  -H "Content-Type: application/json" \
  -d '{"repo_url": "https://github.com/user/plugin-weather"}'

# 启动插件
curl -X POST http://127.0.0.1:8080/api/plugins/start \
  -H "Content-Type: application/json" \
  -d '{"name": "weather"}'

# 停止插件
curl -X POST http://127.0.0.1:8080/api/plugins/stop \
  -H "Content-Type: application/json" \
  -d '{"name": "weather"}'
```

## 开发插件

### 1. 创建新项目

```bash
mkdir plugin-mybot && cd plugin-mybot
go mod init github.com/user/plugin-mybot
```

### 2. 编写插件

```go
package main

import (
    "context"
    "bot-platform/pkg/pluginsdk"
)

type MyPlugin struct {
    bot *pluginsdk.BotClient
}

func (p *MyPlugin) Info() pluginsdk.PluginInfo {
    return pluginsdk.PluginInfo{
        Name:        "mybot",
        Version:     "1.0.0",
        Description: "My awesome bot plugin",
        Author:      "Your Name",
        Commands:    []string{"hello", "hi"},
    }
}

func (p *MyPlugin) OnStart(bot *pluginsdk.BotClient) error {
    p.bot = bot
    return nil
}

func (p *MyPlugin) OnStop() error {
    return nil
}

func (p *MyPlugin) OnMessage(ctx context.Context, bot *pluginsdk.BotClient, msg *pluginsdk.Message) bool {
    return false
}

func (p *MyPlugin) OnCommand(ctx context.Context, bot *pluginsdk.BotClient, cmd string, args []string, msg *pluginsdk.Message) bool {
    switch cmd {
    case "hello", "hi":
        bot.Reply(msg, pluginsdk.Text("Hello! 👋"))
        return true
    }
    return false
}

func main() {
    pluginsdk.Run(&MyPlugin{})
}
```

### 3. 发布到 GitHub

1. 创建 GitHub 仓库
2. 添加 `.github/workflows/release.yml` (参考 `examples/plugin-weather`)
3. 打标签并推送触发自动构建

```bash
git tag v1.0.0
git push origin v1.0.0
```

### 4. 安装使用

```bash
./botctl install https://github.com/user/plugin-mybot
./botctl start mybot
```

## Docker 部署

### 单独部署 Bot Platform

```bash
docker build -t bot-platform:latest .
docker run -d \
  -p 8080:8080 \
  -p 50051:50051 \
  -v ./config.yaml:/app/config.yaml \
  -v ./plugins-bin:/app/plugins-bin \
  -v ./plugins-config:/app/plugins-config \
  bot-platform:latest
```

### 使用 Docker Compose 部署完整环境

```bash
docker-compose up -d
```

这将启动：
- NapCat (QQ 客户端)
- Bot Platform (消息处理平台)

## 项目结构

```
bot-platform/
├── api/proto/           # gRPC 协议定义
├── cmd/
│   ├── main.go          # 主程序入口
│   └── botctl/          # CLI 管理工具
├── internal/
│   ├── bot/             # 核心 Bot 逻辑
│   ├── config/          # 配置管理
│   ├── message/         # 消息构建
│   ├── plugin/          # 内置插件管理
│   ├── pluginmgr/       # 外部插件管理
│   └── server/          # Admin HTTP API
├── pkg/pluginsdk/       # 插件开发 SDK
├── plugins/             # 内置插件
│   ├── echo/
│   └── help/
└── examples/            # 示例外部插件
    ├── plugin-echo-external/
    └── plugin-weather/
```

## 内置命令

- `/help` - 显示帮助信息（列出所有可用插件和命令）

## 外部插件示例

- `echo-ext` - Echo 插件（支持 `/echo`, `/say`, `/repeat` 命令）
- `plugin-weather` - 天气查询插件

## License

MIT
