// plugin-echo-external - A simple external echo plugin for bot-platform
package main

import (
	"context"
	"strings"

	"github.com/DaikonSushi/bot-platform/pkg/pluginsdk"
)

// EchoPlugin is a simple echo plugin
type EchoPlugin struct {
	bot *pluginsdk.BotClient
}

// Info returns plugin metadata
func (p *EchoPlugin) Info() pluginsdk.PluginInfo {
	return pluginsdk.PluginInfo{
		Name:              "echo-ext",
		Version:           "1.0.0",
		Description:       "A simple external echo plugin - echoes back your message",
		Author:            "hovanzhang",
		Commands:          []string{"echo", "say", "repeat"},
		HandleAllMessages: false,
	}
}

// OnStart is called when the plugin starts
func (p *EchoPlugin) OnStart(bot *pluginsdk.BotClient) error {
	p.bot = bot
	bot.Log("info", "Echo external plugin started!")
	return nil
}

// OnStop is called when the plugin stops
func (p *EchoPlugin) OnStop() error {
	return nil
}

// OnMessage handles incoming messages (not used for this plugin)
func (p *EchoPlugin) OnMessage(ctx context.Context, bot *pluginsdk.BotClient, msg *pluginsdk.Message) bool {
	return false
}

// OnCommand handles commands
func (p *EchoPlugin) OnCommand(ctx context.Context, bot *pluginsdk.BotClient, cmd string, args []string, msg *pluginsdk.Message) bool {
	if len(args) == 0 {
		bot.Reply(msg, 
			pluginsdk.Text("🔊 Echo Plugin\n"),
			pluginsdk.Text("━━━━━━━━━━━━━━\n"),
			pluginsdk.Text("用法:\n"),
			pluginsdk.Text("  /echo <消息> - 回显消息\n"),
			pluginsdk.Text("  /say <消息> - 说话\n"),
			pluginsdk.Text("  /repeat <消息> - 重复消息"),
		)
		return true
	}

	text := strings.Join(args, " ")
	
	var response string
	switch cmd {
	case "echo":
		response = "🔊 " + text
	case "say":
		response = "💬 " + text
	case "repeat":
		response = "🔁 " + text + "\n🔁 " + text
	}

	bot.Reply(msg, pluginsdk.Text(response))
	return true
}

func main() {
	pluginsdk.Run(&EchoPlugin{})
}
