# Help Plugin - Before and After Examples

This document shows the difference between the default help output and customized versions.

## Default Configuration (No customization)

### Config:
```yaml
# No help section in config.yaml
```

### Output when user types `/help`:
```
📖 Bot Help Menu
================

【Built-in Plugins】

▸ help
  Show help information
  Commands: /help, /menu

▸ pluginctl
  Manage external plugins via bot commands
  Commands: /plugin, /pm

【External Plugins】

▸ weather (v1.0.0)
  Weather query plugin
  Commands: /weather, /天气
  Author: DaikonSushi

📴 1 external plugin(s) installed but not running
```

---

## Example 1: Custom Title and Description

### Config:
```yaml
help:
  title: "🤖 MyBot Command Center"
  description: |
    Welcome to MyBot! I can help you with various tasks.
    All commands start with /
```

### Output:
```
🤖 MyBot Command Center
=======================

Welcome to MyBot! I can help you with various tasks.
All commands start with /

【Built-in Plugins】

▸ help
  Show help information
  Commands: /help, /menu

▸ pluginctl
  Manage external plugins via bot commands
  Commands: /plugin, /pm

【External Plugins】

▸ weather (v1.0.0)
  Weather query plugin
  Commands: /weather, /天气
  Author: DaikonSushi

📴 1 external plugin(s) installed but not running
Use '/plugin list' to see all plugins
```

---

## Example 2: With Footer

### Config:
```yaml
help:
  title: "🎮 Game Bot Help"
  footer: |
    💡 Pro Tips:
    • Use /plugin info <name> for detailed plugin info
    • Report bugs: https://github.com/user/repo/issues
    • Join our Discord: https://discord.gg/xxxxx
```

### Output:
```
🎮 Game Bot Help
================

【Built-in Plugins】

▸ help
  Show help information
  Commands: /help, /menu

▸ pluginctl
  Manage external plugins via bot commands
  Commands: /plugin, /pm

【External Plugins】

▸ weather (v1.0.0)
  Weather query plugin
  Commands: /weather, /天气
  Author: DaikonSushi

📴 1 external plugin(s) installed but not running
Use '/plugin list' to see all plugins

💡 Pro Tips:
• Use /plugin info <name> for detailed plugin info
• Report bugs: https://github.com/user/repo/issues
• Join our Discord: https://discord.gg/xxxxx
```

---

## Example 3: Only Show External Plugins

### Config:
```yaml
help:
  title: "📦 Available Plugins"
  description: "Here are the community plugins I have installed:"
  show_builtin: false
  show_external: true
```

### Output:
```
📦 Available Plugins
====================

Here are the community plugins I have installed:

【External Plugins】

▸ weather (v1.0.0)
  Weather query plugin
  Commands: /weather, /天气
  Author: DaikonSushi

▸ translate (v2.1.0)
  Multi-language translation
  Commands: /translate, /trans
  Author: TranslateBot

📴 1 external plugin(s) installed but not running
Use '/plugin list' to see all plugins
```

---

## Example 4: Only Show Built-in Commands

### Config:
```yaml
help:
  title: "🔧 System Commands"
  description: "Core bot management commands:"
  show_builtin: true
  show_external: false
```

### Output:
```
🔧 System Commands
==================

Core bot management commands:

【Built-in Plugins】

▸ help
  Show help information
  Commands: /help, /menu

▸ pluginctl
  Manage external plugins via bot commands
  Commands: /plugin, /pm
```

---

## Example 5: Minimal Style

### Config:
```yaml
help:
  title: "Commands"
  show_builtin: true
  show_external: true
```

### Output:
```
Commands
========

【Built-in Plugins】

▸ help
  Show help information
  Commands: /help, /menu

▸ pluginctl
  Manage external plugins via bot commands
  Commands: /plugin, /pm

【External Plugins】

▸ weather (v1.0.0)
  Weather query plugin
  Commands: /weather, /天气
  Author: DaikonSushi

📴 1 external plugin(s) installed but not running
Use '/plugin list' to see all plugins
```

---

## Example 6: Full Customization

### Config:
```yaml
help:
  title: "🌟 SuperBot v2.0 - Command Reference"
  description: |
    👋 Hello! I'm SuperBot, your personal assistant.
    
    📌 Quick Start:
    • Type any command below to get started
    • Use /plugin list to see all available plugins
    • Need help? Contact @admin
  footer: |
    ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
    💡 Tips & Resources:
    • Documentation: https://docs.superbot.com
    • GitHub: https://github.com/user/superbot
    • Support: support@superbot.com
    
    ⭐ Like SuperBot? Star us on GitHub!
  show_builtin: true
  show_external: true
```

### Output:
```
🌟 SuperBot v2.0 - Command Reference
====================================

👋 Hello! I'm SuperBot, your personal assistant.

📌 Quick Start:
• Type any command below to get started
• Use /plugin list to see all available plugins
• Need help? Contact @admin

【Built-in Plugins】

▸ help
  Show help information
  Commands: /help, /menu

▸ pluginctl
  Manage external plugins via bot commands
  Commands: /plugin, /pm

【External Plugins】

▸ weather (v1.0.0)
  Weather query plugin
  Commands: /weather, /天气
  Author: DaikonSushi

📴 1 external plugin(s) installed but not running
Use '/plugin list' to see all plugins

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
💡 Tips & Resources:
• Documentation: https://docs.superbot.com
• GitHub: https://github.com/user/superbot
• Support: support@superbot.com

⭐ Like SuperBot? Star us on GitHub!
```

---

## Tips for Customization

1. **Keep it readable**: Don't make the help message too long
2. **Use emoji wisely**: They make the message more engaging but don't overuse
3. **Provide useful links**: Include documentation, support channels, etc.
4. **Test on mobile**: Make sure it looks good on mobile QQ clients
5. **Update regularly**: Keep the description and footer up-to-date with new features

## Common Use Cases

### For Public Bots
- Show all plugins and provide clear documentation links
- Include support contact information
- Add terms of service or usage guidelines

### For Private/Team Bots
- Customize title with team name
- Hide external plugins if not used
- Add internal wiki or documentation links

### For Gaming Bots
- Use gaming-themed emoji and language
- Highlight most popular commands
- Add Discord/community server links

### For Utility Bots
- Keep it minimal and professional
- Focus on functionality over aesthetics
- Provide clear command examples
