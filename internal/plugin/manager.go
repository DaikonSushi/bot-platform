package plugin

import (
	"log"
	"strings"
	"sync"

	"github.com/DaikonSushi/bot-platform/internal/message"
)

// Manager manages all registered plugins
type Manager struct {
	plugins       []Plugin
	commandMap    map[string]Plugin
	commandPrefix string
	mu            sync.RWMutex
}

// NewManager creates a new plugin manager
func NewManager(commandPrefix string) *Manager {
	return &Manager{
		plugins:       make([]Plugin, 0),
		commandMap:    make(map[string]Plugin),
		commandPrefix: commandPrefix,
	}
}

// Register registers a plugin
func (m *Manager) Register(p Plugin) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.plugins = append(m.plugins, p)

	// Register commands
	for _, cmd := range p.Commands() {
		if existing, ok := m.commandMap[cmd]; ok {
			log.Printf("[Plugin] Warning: command '%s' already registered by plugin '%s', overwriting with '%s'",
				cmd, existing.Name(), p.Name())
		}
		m.commandMap[cmd] = p
	}

	log.Printf("[Plugin] Registered plugin: %s (commands: %v)", p.Name(), p.Commands())
}

// HandleEvent processes an incoming event
func (m *Manager) HandleEvent(ctx *Context) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if it's a command
	if cmd, args, ok := m.parseCommand(ctx.Event); ok {
		if plugin, exists := m.commandMap[cmd]; exists {
			log.Printf("[Plugin] Dispatching command '%s' to plugin '%s'", cmd, plugin.Name())
			return plugin.OnCommand(ctx, cmd, args)
		}
		log.Printf("[Plugin] Unknown command: %s", cmd)
		return false
	}

	// Otherwise, let all plugins handle the message
	for _, p := range m.plugins {
		if p.OnMessage(ctx) {
			return true
		}
	}

	return false
}

// parseCommand extracts command and arguments from message
func (m *Manager) parseCommand(event *message.Event) (cmd string, args []string, ok bool) {
	text := strings.TrimSpace(event.GetText())

	if !strings.HasPrefix(text, m.commandPrefix) {
		return "", nil, false
	}

	// Remove prefix
	text = strings.TrimPrefix(text, m.commandPrefix)
	parts := strings.Fields(text)

	if len(parts) == 0 {
		return "", nil, false
	}

	return parts[0], parts[1:], true
}

// GetPlugins returns all registered plugins
func (m *Manager) GetPlugins() []Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins
}

// GetCommands returns all registered commands
func (m *Manager) GetCommands() map[string]Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Plugin)
	for k, v := range m.commandMap {
		result[k] = v
	}
	return result
}
