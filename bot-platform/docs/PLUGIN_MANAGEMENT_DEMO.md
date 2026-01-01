# 通过消息管理插件功能演示

## 场景：在 QQ 中直接管理插件

### 1. 查看帮助信息

**发送消息：**
```
/plugin
```

**Bot 回复：**
```
🔧 Plugin Management Commands

Usage: /plugin <command> [args]
Alias: /pm <command> [args]

Commands:
  install <repo_url>    Install plugin from GitHub
  start <name>          Start an installed plugin
  stop <name>           Stop a running plugin
  restart <name>        Restart a plugin
  uninstall <name>      Uninstall a plugin
  list                  List all installed plugins
  info <name>           Show detailed info about a plugin

Note: Only administrators can use these commands.
```

---

### 2. 安装插件

**发送消息：**
```
/pm install https://github.com/DaikonSushi/plugin-weather
```

**Bot 回复：**
```
⏳ Installing plugin from https://github.com/DaikonSushi/plugin-weather...

✅ Plugin installed successfully!

Name: weather
Version: 1.0.0
Description: Weather query plugin
Commands: /weather, /天气

Use '/plugin start weather' to start it.
```

---

### 3. 启动插件

**发送消息：**
```
/pm start weather
```

**Bot 回复：**
```
⏳ Starting plugin 'weather'...

✅ Plugin 'weather' started successfully!
```

---

### 4. 查看所有插件

**发送消息：**
```
/pm list
```

**Bot 回复：**
```
📦 Installed Plugins
==================

🟢 weather (v1.0.0) - running
   Weather query plugin
   Commands: /weather, /天气
   Uptime: 2m30s

🔴 echo-ext (v1.0.0) - stopped
   Echo plugin with multiple commands
   Commands: /echo, /say, /repeat

Total: 2 plugins (1 running, 1 stopped)
```

---

### 5. 使用插件功能

**发送消息：**
```
/weather Beijing
```

**Bot 回复：**
```
🌤️ Beijing Weather

Temperature: 15°C
Condition: Partly Cloudy
Humidity: 45%
Wind: 12 km/h
```

---

### 6. 查看插件详情

**发送消息：**
```
/pm info weather
```

**Bot 回复：**
```
📋 Plugin Information: weather
========================

Name: weather
Version: 1.0.0
Description: Weather query plugin
Author: DaikonSushi
Repository: https://github.com/DaikonSushi/plugin-weather
Commands: /weather, /天气

Status: 🟢 running
Port: 50100
Started: 2026-01-01 22:45:30
Uptime: 5m15s
```

---

### 7. 重启插件

**发送消息：**
```
/pm restart weather
```

**Bot 回复：**
```
⏳ Restarting plugin 'weather'...

✅ Plugin 'weather' restarted successfully!
```

---

### 8. 停止插件

**发送消息：**
```
/pm stop weather
```

**Bot 回复：**
```
⏳ Stopping plugin 'weather'...

✅ Plugin 'weather' stopped successfully!
```

---

### 9. 卸载插件

**发送消息：**
```
/pm uninstall weather
```

**Bot 回复：**
```
⏳ Uninstalling plugin 'weather'...

✅ Plugin 'weather' uninstalled successfully!
```

---

## 权限控制

如果非管理员用户尝试使用插件管理命令：

**发送消息：**
```
/pm list
```

**Bot 回复：**
```
❌ Permission denied. Only admins can manage plugins.
```

---

## 配置管理员

在 `config.yaml` 中配置管理员 QQ 号：

```yaml
bot:
  admins:
    - 123456789    # 你的 QQ 号
    - 987654321    # 其他管理员的 QQ 号
  command_prefix: "/"
  debug: false
```

---

## 优势

相比命令行工具和 HTTP API：

✅ **更方便** - 无需登录服务器或使用 curl  
✅ **更直观** - 直接在聊天界面操作  
✅ **实时反馈** - 立即看到操作结果  
✅ **移动友好** - 手机上也能轻松管理  
✅ **权限控制** - 自动验证管理员身份  

---

## 注意事项

1. 确保 `plugin_manager.enabled` 设置为 `true`
2. 只有配置的管理员才能使用这些命令
3. 插件安装需要网络连接到 GitHub
4. 所有操作都有超时保护，避免长时间等待
