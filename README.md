# Bot Platform

基于 NapCat 的 QQ 机器人消息处理平台，支持插件热加载。

> 🤖 本项目多数代码由 Claude vibe coding 生成，请自行解决网络问题。

## 快速部署

推荐使用 Docker Compose 部署（跨平台）：

```bash
git clone https://github.com/DaikonSushi/bot-platform.git
cd bot-platform

# 配置管理员 QQ 号
cp config.example.yaml config.yaml
vim config.yaml

# 启动
docker-compose up -d

# 扫码登录: http://localhost:6099
```

## 插件管理

在 QQ 中给 Bot 发送命令（仅管理员）：

```
/plugin install <repo_url>   # 安装插件
/plugin start <name>         # 启动
/plugin stop <name>          # 停止
/plugin list                 # 查看所有插件
/plugin uninstall <name>     # 卸载
```

### 示例：安装 ShowMeJM 插件

```
/plugin install https://github.com/DaikonSushi/plugin-showmejm
/plugin start showmejm
```

## 开发插件

clone [plugin-fileupload](https://github.com/DaikonSushi/plugin-fileupload) 作为模板：

```bash
git clone https://github.com/DaikonSushi/plugin-fileupload.git plugin-myplugin
cd plugin-myplugin

# 1. 修改 go.mod 模块名
# 2. 编写插件逻辑
# 3. 打 tag 发布（GitHub Actions 自动构建）
git tag v1.0.0
git push origin v1.0.0
```

详细开发文档见 [docs/PLUGIN_DEVELOPMENT.md](docs/PLUGIN_DEVELOPMENT.md)

## 其他文档

- [Docker 部署](docs/DEPLOYMENT_SUMMARY.md)
- [多架构支持](docs/MULTI_ARCH.md)
- [Help 配置](docs/HELP_CONFIGURATION.md)

## License

MIT
