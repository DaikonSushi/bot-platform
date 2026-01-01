# Help Plugin Configuration Guide

The help plugin now supports customization through the `config.yaml` file. You can customize the title, description, footer, and control which plugins to display.

## Configuration Options

Add the following section to your `config.yaml`:

```yaml
help:
  # Custom title for help menu (supports emoji)
  title: "🤖 My Bot Help"
  
  # Custom description/header text (optional)
  description: "Welcome! Here are all available commands:"
  
  # Custom footer text (optional)
  footer: "💡 Tip: Use /plugin list to manage external plugins"
  
  # Show built-in plugins (default: true)
  show_builtin: true
  
  # Show external plugins (default: true)
  show_external: true
```

## Configuration Fields

### `title` (string)
- **Default**: `"📖 Bot Help Menu"`
- **Description**: The main title displayed at the top of the help message
- **Supports**: Emoji and Unicode characters
- **Example**: `"🤖 My Awesome Bot"`, `"Help Center"`

### `description` (string)
- **Default**: Empty (no description)
- **Description**: Additional text shown below the title, before the plugin list
- **Use case**: Welcome message, usage instructions, or important notes
- **Example**: 
  ```yaml
  description: |
    Welcome to MyBot! 
    Use the commands below to interact with me.
    For support, contact @admin
  ```

### `footer` (string)
- **Default**: Empty (no footer)
- **Description**: Text shown at the bottom of the help message
- **Use case**: Tips, links, credits, or additional information
- **Example**: 
  ```yaml
  footer: |
    💡 Pro tip: Type /plugin list to see all plugins
    🔗 Documentation: https://example.com/docs
  ```

### `show_builtin` (boolean)
- **Default**: `true`
- **Description**: Whether to display built-in plugins in the help menu
- **Use case**: Hide built-in plugins if you only want to show external ones

### `show_external` (boolean)
- **Default**: `true`
- **Description**: Whether to display external plugins in the help menu
- **Use case**: Hide external plugins if you only want to show built-in ones

## Examples

### Example 1: Minimal Configuration

```yaml
help:
  title: "🎮 Game Bot Commands"
```

**Result:**
```
🎮 Game Bot Commands
====================

【Built-in Plugins】
...
```

---

### Example 2: Full Customization

```yaml
help:
  title: "🤖 MyBot Help Center"
  description: |
    Welcome to MyBot! I'm here to help you.
    All commands start with /
  footer: |
    💡 Need more help? Use /plugin info <name>
    🐛 Report bugs: https://github.com/user/repo/issues
  show_builtin: true
  show_external: true
```

**Result:**
```
🤖 MyBot Help Center
====================

Welcome to MyBot! I'm here to help you.
All commands start with /

【Built-in Plugins】
...

【External Plugins】
...

💡 Need more help? Use /plugin info <name>
🐛 Report bugs: https://github.com/user/repo/issues
```

---

### Example 3: Only Show External Plugins

```yaml
help:
  title: "📦 Available Plugins"
  description: "Here are the installed plugins:"
  show_builtin: false
  show_external: true
```

**Result:**
```
📦 Available Plugins
====================

Here are the installed plugins:

【External Plugins】
...
```

---

### Example 4: Only Show Built-in Plugins

```yaml
help:
  title: "🔧 Core Commands"
  show_builtin: true
  show_external: false
```

**Result:**
```
🔧 Core Commands
================

【Built-in Plugins】
...
```

---

## Multi-line Text

For `description` and `footer`, you can use YAML's multi-line syntax:

```yaml
help:
  description: |
    Line 1
    Line 2
    Line 3
  
  # Or use folded style (joins lines with spaces)
  footer: >
    This is a long footer text
    that will be joined into
    a single line.
```

## Default Behavior

If you don't specify the `help` section in your config, the plugin will use these defaults:

```yaml
help:
  title: "📖 Bot Help Menu"
  description: ""
  footer: ""
  show_builtin: true
  show_external: true
```

## Tips

1. **Keep it concise**: Help messages are sent as chat messages, so avoid very long text
2. **Use emoji**: Emoji can make your help menu more visually appealing
3. **Test your config**: After changing the config, restart the bot and test with `/help`
4. **Localization**: You can create different config files for different languages

## Troubleshooting

### Help menu shows default title even after config change
- Make sure you've restarted the bot after modifying `config.yaml`
- Check that your YAML syntax is correct (proper indentation)

### Multi-line text not working
- Use the `|` (literal) or `>` (folded) indicator for multi-line strings
- Ensure proper indentation (2 spaces per level)

### Emoji not displaying correctly
- Make sure your terminal/chat client supports Unicode emoji
- Some older systems may not render emoji properly
