package help

import (
	"fmt"
	"strings"

	"github.com/DaikonSushi/bot-platform/internal/config"
	"github.com/DaikonSushi/bot-platform/internal/message"
	"github.com/DaikonSushi/bot-platform/internal/plugin"
	"github.com/DaikonSushi/bot-platform/internal/pluginmgr"
)

// HelpPlugin provides help information
type HelpPlugin struct {
	plugin.BasePlugin
	manager    *plugin.Manager
	extManager *pluginmgr.PluginManager
	helpConfig *config.HelpConfig
}

// New creates a new help plugin
func New(manager *plugin.Manager, extManager *pluginmgr.PluginManager, helpConfig *config.HelpConfig) *HelpPlugin {
	return &HelpPlugin{
		BasePlugin: plugin.BasePlugin{
			PluginName:        "help",
			PluginDescription: "Show help information",
			PluginCommands:    []string{"help", "menu"},
		},
		manager:    manager,
		extManager: extManager,
		helpConfig: helpConfig,
	}
}

// OnCommand handles help commands
func (p *HelpPlugin) OnCommand(ctx *plugin.Context, cmd string, args []string) bool {
	var sb strings.Builder
	
	// Custom title or default
	title := p.helpConfig.Title
	if title == "" {
		title = "📖 Bot Help Menu"
	}
	sb.WriteString(title + "\n")
	sb.WriteString(strings.Repeat("=", len([]rune(title))) + "\n\n")

	// Custom description/header
	if p.helpConfig.Description != "" {
		sb.WriteString(p.helpConfig.Description + "\n\n")
	}

	// Show built-in plugins
	if p.helpConfig.ShowBuiltin {
		sb.WriteString("【Built-in Plugins】\n\n")
		plugins := p.manager.GetPlugins()
		for _, plug := range plugins {
			sb.WriteString(fmt.Sprintf("▸ %s\n", plug.Name()))
			sb.WriteString(fmt.Sprintf("  %s\n", plug.Description()))
			if cmds := plug.Commands(); len(cmds) > 0 {
				sb.WriteString(fmt.Sprintf("  Commands: /%s\n", strings.Join(cmds, ", /")))
			}
			sb.WriteString("\n")
		}
	}

	// Show external plugins if available
	if p.helpConfig.ShowExternal && p.extManager != nil {
		extPlugins := p.extManager.GetRunningPlugins()
		if len(extPlugins) > 0 {
			sb.WriteString("【External Plugins】\n\n")
			for _, state := range extPlugins {
				sb.WriteString(fmt.Sprintf("▸ %s (v%s)\n", state.Info.Name, state.Info.Version))
				sb.WriteString(fmt.Sprintf("  %s\n", state.Info.Description))
				if len(state.Info.Commands) > 0 {
					sb.WriteString(fmt.Sprintf("  Commands: /%s\n", strings.Join(state.Info.Commands, ", /")))
				}
				if state.Info.Author != "" {
					sb.WriteString(fmt.Sprintf("  Author: %s\n", state.Info.Author))
				}
				sb.WriteString("\n")
			}
		}

		// Show stopped plugins
		allPlugins := p.extManager.ListPlugins()
		stoppedCount := 0
		for _, state := range allPlugins {
			if state.Status != "running" {
				stoppedCount++
			}
		}
		if stoppedCount > 0 {
			sb.WriteString(fmt.Sprintf("📴 %d external plugin(s) installed but not running\n", stoppedCount))
			sb.WriteString("Use '/plugin list' to see all plugins\n\n")
		}
	}

	// Custom footer
	if p.helpConfig.Footer != "" {
		sb.WriteString(p.helpConfig.Footer + "\n")
	}

	msg := message.NewMessage().Text(sb.String())
	ctx.Bot.Reply(ctx, msg)
	return true
}
