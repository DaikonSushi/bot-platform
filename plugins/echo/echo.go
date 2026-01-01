package echo

import (
	"strings"

	"github.com/DaikonSushi/bot-platform/internal/message"
	"github.com/DaikonSushi/bot-platform/internal/plugin"
)

// EchoPlugin is a simple echo plugin for testing
type EchoPlugin struct {
	plugin.BasePlugin
}

// New creates a new echo plugin
func New() *EchoPlugin {
	return &EchoPlugin{
		BasePlugin: plugin.BasePlugin{
			PluginName:        "echo",
			PluginDescription: "Echo back messages (for testing)",
			PluginCommands:    []string{"echo", "say"},
		},
	}
}

// OnCommand handles echo commands
func (p *EchoPlugin) OnCommand(ctx *plugin.Context, cmd string, args []string) bool {
	if len(args) == 0 {
		msg := message.NewMessage().Text("Usage: /echo <message>")
		ctx.Bot.Reply(ctx, msg)
		return true
	}

	text := strings.Join(args, " ")
	msg := message.NewMessage().Text(text)
	ctx.Bot.Reply(ctx, msg)
	return true
}
