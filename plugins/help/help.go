package help

import (
	"fmt"
	"strings"

	"github.com/DaikonSushi/bot-platform/internal/message"
	"github.com/DaikonSushi/bot-platform/internal/plugin"
)

// HelpPlugin provides help information
type HelpPlugin struct {
	plugin.BasePlugin
	manager *plugin.Manager
}

// New creates a new help plugin
func New(manager *plugin.Manager) *HelpPlugin {
	return &HelpPlugin{
		BasePlugin: plugin.BasePlugin{
			PluginName:        "help",
			PluginDescription: "Show help information",
			PluginCommands:    []string{"help", "menu"},
		},
		manager: manager,
	}
}

// OnCommand handles help commands
func (p *HelpPlugin) OnCommand(ctx *plugin.Context, cmd string, args []string) bool {
	var sb strings.Builder
	sb.WriteString("📖 Bot Help Menu\n")
	sb.WriteString("================\n\n")

	plugins := p.manager.GetPlugins()
	for _, plug := range plugins {
		sb.WriteString(fmt.Sprintf("【%s】\n", plug.Name()))
		sb.WriteString(fmt.Sprintf("  %s\n", plug.Description()))
		if cmds := plug.Commands(); len(cmds) > 0 {
			sb.WriteString(fmt.Sprintf("  Commands: /%s\n", strings.Join(cmds, ", /")))
		}
		sb.WriteString("\n")
	}

	msg := message.NewMessage().Text(sb.String())
	ctx.Bot.Reply(ctx, msg)
	return true
}
