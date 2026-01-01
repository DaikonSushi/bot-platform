package plugin

import (
	"github.com/DaikonSushi/bot-platform/internal/message"
)

// Context provides plugin access to bot functionality
type Context struct {
	Event   *message.Event
	Bot     BotAPI
	IsAdmin bool
}

// BotAPI interface for plugins to interact with the bot
type BotAPI interface {
	// SendPrivateMessage sends a private message to a user
	SendPrivateMessage(userID int64, msg *message.Message) error
	// SendGroupMessage sends a message to a group
	SendGroupMessage(groupID int64, msg *message.Message) error
	// Reply replies to the current message context
	Reply(ctx *Context, msg *message.Message) error
	// GetLoginInfo gets bot login info
	GetLoginInfo() (*LoginInfo, error)
}

// LoginInfo represents bot login information
type LoginInfo struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
}

// Plugin interface that all plugins must implement
type Plugin interface {
	// Name returns the plugin name
	Name() string
	// Description returns plugin description
	Description() string
	// Commands returns list of commands this plugin handles
	Commands() []string
	// OnMessage is called when a message is received
	// Returns true if the message was handled
	OnMessage(ctx *Context) bool
	// OnCommand is called when a registered command is triggered
	OnCommand(ctx *Context, cmd string, args []string) bool
}

// BasePlugin provides default implementations
type BasePlugin struct {
	PluginName        string
	PluginDescription string
	PluginCommands    []string
}

func (p *BasePlugin) Name() string {
	return p.PluginName
}

func (p *BasePlugin) Description() string {
	return p.PluginDescription
}

func (p *BasePlugin) Commands() []string {
	return p.PluginCommands
}

func (p *BasePlugin) OnMessage(ctx *Context) bool {
	return false
}

func (p *BasePlugin) OnCommand(ctx *Context, cmd string, args []string) bool {
	return false
}
