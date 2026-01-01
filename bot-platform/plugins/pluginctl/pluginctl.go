package pluginctl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DaikonSushi/bot-platform/internal/message"
	"github.com/DaikonSushi/bot-platform/internal/plugin"
	"github.com/DaikonSushi/bot-platform/internal/pluginmgr"
)

// PluginCtlPlugin provides plugin management through bot commands
type PluginCtlPlugin struct {
	plugin.BasePlugin
	extManager *pluginmgr.PluginManager
}

// New creates a new plugin control plugin
func New(extManager *pluginmgr.PluginManager) *PluginCtlPlugin {
	return &PluginCtlPlugin{
		BasePlugin: plugin.BasePlugin{
			PluginName:        "pluginctl",
			PluginDescription: "Manage external plugins via bot commands",
			PluginCommands:    []string{"plugin", "pm"},
		},
		extManager: extManager,
	}
}

// OnCommand handles plugin management commands
func (p *PluginCtlPlugin) OnCommand(ctx *plugin.Context, cmd string, args []string) bool {
	// Only admins can manage plugins
	if !ctx.IsAdmin {
		msg := message.NewMessage().Text("❌ Permission denied. Only admins can manage plugins.")
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	if p.extManager == nil {
		msg := message.NewMessage().Text("❌ External plugin manager is not enabled.")
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	if len(args) == 0 {
		p.showHelp(ctx)
		return true
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "install":
		return p.handleInstall(ctx, subArgs)
	case "start":
		return p.handleStart(ctx, subArgs)
	case "stop":
		return p.handleStop(ctx, subArgs)
	case "restart":
		return p.handleRestart(ctx, subArgs)
	case "uninstall", "remove":
		return p.handleUninstall(ctx, subArgs)
	case "list", "ls":
		return p.handleList(ctx, subArgs)
	case "info":
		return p.handleInfo(ctx, subArgs)
	default:
		p.showHelp(ctx)
		return true
	}
}

// showHelp displays help information
func (p *PluginCtlPlugin) showHelp(ctx *plugin.Context) {
	help := `🔧 Plugin Management Commands

Usage: /plugin <command> [args]
Alias: /pm <command> [args]

Commands:
  install <repo_url>    Install plugin from GitHub
                        Example: /pm install https://github.com/user/plugin-weather
  
  start <name>          Start an installed plugin
                        Example: /pm start weather
  
  stop <name>           Stop a running plugin
                        Example: /pm stop weather
  
  restart <name>        Restart a plugin
                        Example: /pm restart weather
  
  uninstall <name>      Uninstall a plugin
                        Example: /pm uninstall weather
  
  list                  List all installed plugins
                        Example: /pm list
  
  info <name>           Show detailed info about a plugin
                        Example: /pm info weather

Note: Only administrators can use these commands.`

	msg := message.NewMessage().Text(help)
	ctx.Bot.Reply(ctx, msg)
}

// handleInstall installs a plugin from GitHub
func (p *PluginCtlPlugin) handleInstall(ctx *plugin.Context, args []string) bool {
	if len(args) == 0 {
		msg := message.NewMessage().Text("❌ Usage: /plugin install <repo_url>\nExample: /plugin install https://github.com/user/plugin-weather")
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	repoURL := args[0]

	// Send initial message
	msg := message.NewMessage().Text(fmt.Sprintf("⏳ Installing plugin from %s...", repoURL))
	ctx.Bot.Reply(ctx, msg)

	// Install with timeout
	installCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	meta, err := p.extManager.InstallFromGitHub(installCtx, repoURL)
	if err != nil {
		msg := message.NewMessage().Text(fmt.Sprintf("❌ Installation failed: %v", err))
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	// Success message
	successMsg := fmt.Sprintf("✅ Plugin installed successfully!\n\n"+
		"Name: %s\n"+
		"Version: %s\n"+
		"Description: %s\n"+
		"Commands: /%s\n\n"+
		"Use '/plugin start %s' to start it.",
		meta.Name, meta.Version, meta.Description,
		strings.Join(meta.Commands, ", /"),
		meta.Name)

	msg = message.NewMessage().Text(successMsg)
	ctx.Bot.Reply(ctx, msg)
	return true
}

// handleStart starts a plugin
func (p *PluginCtlPlugin) handleStart(ctx *plugin.Context, args []string) bool {
	if len(args) == 0 {
		msg := message.NewMessage().Text("❌ Usage: /plugin start <name>\nExample: /plugin start weather")
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	name := args[0]

	msg := message.NewMessage().Text(fmt.Sprintf("⏳ Starting plugin '%s'...", name))
	ctx.Bot.Reply(ctx, msg)

	startCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := p.extManager.StartPlugin(startCtx, name)
	if err != nil {
		msg := message.NewMessage().Text(fmt.Sprintf("❌ Failed to start plugin: %v", err))
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	msg = message.NewMessage().Text(fmt.Sprintf("✅ Plugin '%s' started successfully!", name))
	ctx.Bot.Reply(ctx, msg)
	return true
}

// handleStop stops a plugin
func (p *PluginCtlPlugin) handleStop(ctx *plugin.Context, args []string) bool {
	if len(args) == 0 {
		msg := message.NewMessage().Text("❌ Usage: /plugin stop <name>\nExample: /plugin stop weather")
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	name := args[0]

	msg := message.NewMessage().Text(fmt.Sprintf("⏳ Stopping plugin '%s'...", name))
	ctx.Bot.Reply(ctx, msg)

	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := p.extManager.StopPlugin(stopCtx, name)
	if err != nil {
		msg := message.NewMessage().Text(fmt.Sprintf("❌ Failed to stop plugin: %v", err))
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	msg = message.NewMessage().Text(fmt.Sprintf("✅ Plugin '%s' stopped successfully!", name))
	ctx.Bot.Reply(ctx, msg)
	return true
}

// handleRestart restarts a plugin
func (p *PluginCtlPlugin) handleRestart(ctx *plugin.Context, args []string) bool {
	if len(args) == 0 {
		msg := message.NewMessage().Text("❌ Usage: /plugin restart <name>\nExample: /plugin restart weather")
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	name := args[0]

	msg := message.NewMessage().Text(fmt.Sprintf("⏳ Restarting plugin '%s'...", name))
	ctx.Bot.Reply(ctx, msg)

	// Stop first
	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err := p.extManager.StopPlugin(stopCtx, name)
	cancel()

	if err != nil {
		msg := message.NewMessage().Text(fmt.Sprintf("❌ Failed to stop plugin: %v", err))
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	// Wait a bit
	time.Sleep(1 * time.Second)

	// Start again
	startCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = p.extManager.StartPlugin(startCtx, name)
	if err != nil {
		msg := message.NewMessage().Text(fmt.Sprintf("❌ Failed to start plugin: %v", err))
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	msg = message.NewMessage().Text(fmt.Sprintf("✅ Plugin '%s' restarted successfully!", name))
	ctx.Bot.Reply(ctx, msg)
	return true
}

// handleUninstall uninstalls a plugin
func (p *PluginCtlPlugin) handleUninstall(ctx *plugin.Context, args []string) bool {
	if len(args) == 0 {
		msg := message.NewMessage().Text("❌ Usage: /plugin uninstall <name>\nExample: /plugin uninstall weather")
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	name := args[0]

	msg := message.NewMessage().Text(fmt.Sprintf("⏳ Uninstalling plugin '%s'...", name))
	ctx.Bot.Reply(ctx, msg)

	uninstallCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := p.extManager.UninstallPlugin(uninstallCtx, name)
	if err != nil {
		msg := message.NewMessage().Text(fmt.Sprintf("❌ Failed to uninstall plugin: %v", err))
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	msg = message.NewMessage().Text(fmt.Sprintf("✅ Plugin '%s' uninstalled successfully!", name))
	ctx.Bot.Reply(ctx, msg)
	return true
}

// handleList lists all installed plugins
func (p *PluginCtlPlugin) handleList(ctx *plugin.Context, args []string) bool {
	plugins := p.extManager.ListPlugins()

	if len(plugins) == 0 {
		msg := message.NewMessage().Text("📦 No external plugins installed.\n\nUse '/plugin install <repo_url>' to install one.")
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	var sb strings.Builder
	sb.WriteString("📦 Installed Plugins\n")
	sb.WriteString("==================\n\n")

	runningCount := 0
	stoppedCount := 0

	for _, state := range plugins {
		statusIcon := "🔴"
		statusText := "stopped"

		if state.Status == "running" {
			statusIcon = "🟢"
			statusText = "running"
			runningCount++
		} else if state.Status == "error" {
			statusIcon = "🟡"
			statusText = "error"
		} else {
			stoppedCount++
		}

		sb.WriteString(fmt.Sprintf("%s %s (v%s) - %s\n", statusIcon, state.Info.Name, state.Info.Version, statusText))
		sb.WriteString(fmt.Sprintf("   %s\n", state.Info.Description))

		if len(state.Info.Commands) > 0 {
			sb.WriteString(fmt.Sprintf("   Commands: /%s\n", strings.Join(state.Info.Commands, ", /")))
		}

		if state.Status == "running" && !state.StartedAt.IsZero() {
			uptime := time.Since(state.StartedAt).Round(time.Second)
			sb.WriteString(fmt.Sprintf("   Uptime: %s\n", uptime))
		}

		if state.LastError != "" {
			sb.WriteString(fmt.Sprintf("   Error: %s\n", state.LastError))
		}

		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Total: %d plugins (%d running, %d stopped)\n", len(plugins), runningCount, stoppedCount))

	msg := message.NewMessage().Text(sb.String())
	ctx.Bot.Reply(ctx, msg)
	return true
}

// handleInfo shows detailed information about a plugin
func (p *PluginCtlPlugin) handleInfo(ctx *plugin.Context, args []string) bool {
	if len(args) == 0 {
		msg := message.NewMessage().Text("❌ Usage: /plugin info <name>\nExample: /plugin info weather")
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	name := args[0]
	plugins := p.extManager.ListPlugins()

	var targetPlugin *pluginmgr.PluginState
	for _, state := range plugins {
		if state.Info.Name == name {
			targetPlugin = state
			break
		}
	}

	if targetPlugin == nil {
		msg := message.NewMessage().Text(fmt.Sprintf("❌ Plugin '%s' not found.", name))
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 Plugin Information: %s\n", targetPlugin.Info.Name))
	sb.WriteString("========================\n\n")
	sb.WriteString(fmt.Sprintf("Name: %s\n", targetPlugin.Info.Name))
	sb.WriteString(fmt.Sprintf("Version: %s\n", targetPlugin.Info.Version))
	sb.WriteString(fmt.Sprintf("Description: %s\n", targetPlugin.Info.Description))

	if targetPlugin.Info.Author != "" {
		sb.WriteString(fmt.Sprintf("Author: %s\n", targetPlugin.Info.Author))
	}

	if targetPlugin.Info.RepoURL != "" {
		sb.WriteString(fmt.Sprintf("Repository: %s\n", targetPlugin.Info.RepoURL))
	}

	if len(targetPlugin.Info.Commands) > 0 {
		sb.WriteString(fmt.Sprintf("Commands: /%s\n", strings.Join(targetPlugin.Info.Commands, ", /")))
	}

	statusIcon := "🔴"
	if targetPlugin.Status == "running" {
		statusIcon = "🟢"
	} else if targetPlugin.Status == "error" {
		statusIcon = "🟡"
	}

	sb.WriteString(fmt.Sprintf("\nStatus: %s %s\n", statusIcon, targetPlugin.Status))

	if targetPlugin.Status == "running" {
		sb.WriteString(fmt.Sprintf("Port: %d\n", targetPlugin.Port))
		if !targetPlugin.StartedAt.IsZero() {
			sb.WriteString(fmt.Sprintf("Started: %s\n", targetPlugin.StartedAt.Format("2006-01-02 15:04:05")))
			uptime := time.Since(targetPlugin.StartedAt).Round(time.Second)
			sb.WriteString(fmt.Sprintf("Uptime: %s\n", uptime))
		}
	}

	if targetPlugin.LastError != "" {
		sb.WriteString(fmt.Sprintf("\nLast Error: %s\n", targetPlugin.LastError))
	}

	msg := message.NewMessage().Text(sb.String())
	ctx.Bot.Reply(ctx, msg)
	return true
}
