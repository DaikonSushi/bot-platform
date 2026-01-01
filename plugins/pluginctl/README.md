# Plugin Management via Bot Commands

The `pluginctl` plugin allows you to manage external plugins directly through bot messages, without needing to use the command-line tool or HTTP API.

## Prerequisites

- You must be configured as an admin in `config.yaml`
- External plugin manager must be enabled in the configuration

## Commands

All commands can use either `/plugin` or `/pm` as the prefix.

### Install a Plugin

Install a plugin from a GitHub repository:

```
/plugin install https://github.com/user/plugin-weather
/pm install https://github.com/user/plugin-weather
```

The bot will:
1. Download the latest release binary for your OS/architecture
2. Extract plugin metadata
3. Save it to the plugins directory
4. Show you the plugin information

### Start a Plugin

Start an installed plugin:

```
/plugin start weather
/pm start weather
```

### Stop a Plugin

Stop a running plugin:

```
/plugin stop weather
/pm stop weather
```

### Restart a Plugin

Restart a plugin (stop then start):

```
/plugin restart weather
/pm restart weather
```

### List All Plugins

View all installed plugins and their status:

```
/plugin list
/pm list
/pm ls
```

This shows:
- Plugin name and version
- Status (running/stopped/error)
- Description
- Available commands
- Uptime (for running plugins)
- Any error messages

### Show Plugin Info

Get detailed information about a specific plugin:

```
/plugin info weather
/pm info weather
```

This displays:
- Full metadata (name, version, author, description)
- Repository URL
- Available commands
- Current status
- Port (if running)
- Start time and uptime
- Last error (if any)

### Uninstall a Plugin

Remove a plugin completely:

```
/plugin uninstall weather
/pm uninstall weather
```

This will:
1. Stop the plugin if it's running
2. Delete the binary file
3. Remove the configuration

## Example Workflow

```
# Install a plugin
/pm install https://github.com/DaikonSushi/plugin-weather

# Start it
/pm start weather

# Check if it's running
/pm list

# Use the plugin's commands
/weather Beijing

# Get detailed info
/pm info weather

# Restart if needed
/pm restart weather

# Stop when not needed
/pm stop weather

# Uninstall completely
/pm uninstall weather
```

## Permissions

Only users listed in the `admins` section of `config.yaml` can use plugin management commands. Regular users will receive a "Permission denied" message.

## Notes

- Plugin installation requires internet access to GitHub
- Plugins must follow the bot-platform plugin specification
- The bot will automatically restart crashed plugins
- All plugin operations have timeouts to prevent hanging
