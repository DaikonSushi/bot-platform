package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/DaikonSushi/bot-platform/internal/config"
	"github.com/DaikonSushi/bot-platform/internal/message"
	"github.com/DaikonSushi/bot-platform/internal/plugin"
)

// Bot represents the main bot instance
type Bot struct {
	config        *config.Config
	httpClient    *http.Client
	wsConn        *websocket.Conn
	pluginManager *plugin.Manager
	running       bool
	mu            sync.RWMutex
	stopChan      chan struct{}
}

// New creates a new bot instance
func New(cfg *config.Config) *Bot {
	return &Bot{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		pluginManager: plugin.NewManager(cfg.Bot.CommandPrefix),
		stopChan:      make(chan struct{}),
	}
}

// RegisterPlugin registers a plugin with the bot
func (b *Bot) RegisterPlugin(p plugin.Plugin) {
	b.pluginManager.Register(p)
}

// Start starts the bot and connects to NapCat
func (b *Bot) Start() error {
	b.mu.Lock()
	b.running = true
	b.mu.Unlock()

	// Test connection first
	info, err := b.GetLoginInfo()
	if err != nil {
		return fmt.Errorf("failed to connect to NapCat: %w", err)
	}
	log.Printf("[Bot] Connected as %s (%d)", info.Nickname, info.UserID)

	// Connect WebSocket for receiving events
	if err := b.connectWebSocket(); err != nil {
		return fmt.Errorf("failed to connect WebSocket: %w", err)
	}

	// Start event loop
	go b.eventLoop()

	log.Println("[Bot] Bot started successfully")
	return nil
}

// Stop stops the bot
func (b *Bot) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return
	}

	b.running = false
	close(b.stopChan)

	if b.wsConn != nil {
		b.wsConn.Close()
	}

	log.Println("[Bot] Bot stopped")
}

// connectWebSocket establishes WebSocket connection
func (b *Bot) connectWebSocket() error {
	header := http.Header{}
	if b.config.NapCat.Token != "" {
		header.Set("Authorization", "Bearer "+b.config.NapCat.Token)
	}

	conn, _, err := websocket.DefaultDialer.Dial(b.config.NapCat.WsURL, header)
	if err != nil {
		return err
	}

	b.wsConn = conn
	log.Printf("[Bot] WebSocket connected to %s", b.config.NapCat.WsURL)
	return nil
}

// eventLoop processes incoming WebSocket events
func (b *Bot) eventLoop() {
	for {
		select {
		case <-b.stopChan:
			return
		default:
			_, msg, err := b.wsConn.ReadMessage()
			if err != nil {
				log.Printf("[Bot] WebSocket read error: %v", err)
				// Try to reconnect
				time.Sleep(5 * time.Second)
				if err := b.connectWebSocket(); err != nil {
					log.Printf("[Bot] Reconnect failed: %v", err)
				}
				continue
			}

			go b.handleEvent(msg)
		}
	}
}

// handleEvent processes a single event
func (b *Bot) handleEvent(data []byte) {
	event, err := message.ParseEvent(data)
	if err != nil {
		if b.config.Bot.Debug {
			log.Printf("[Bot] Failed to parse event: %v", err)
		}
		return
	}

	// Only handle message events
	if event.PostType != "message" {
		return
	}

	if b.config.Bot.Debug {
		log.Printf("[Bot] Received %s message from %d: %s",
			event.MessageType, event.UserID, event.RawMessage)
	}

	// Create context
	ctx := &plugin.Context{
		Event:   event,
		Bot:     b,
		IsAdmin: b.config.IsAdmin(event.UserID),
	}

	// Dispatch to plugin manager
	b.pluginManager.HandleEvent(ctx)
}

// SendPrivateMessage sends a private message
func (b *Bot) SendPrivateMessage(userID int64, msg *message.Message) error {
	return b.callAPI("send_private_msg", map[string]interface{}{
		"user_id": userID,
		"message": msg.Build(),
	})
}

// SendGroupMessage sends a group message
func (b *Bot) SendGroupMessage(groupID int64, msg *message.Message) error {
	return b.callAPI("send_group_msg", map[string]interface{}{
		"group_id": groupID,
		"message":  msg.Build(),
	})
}

// Reply replies to the current message context
func (b *Bot) Reply(ctx *plugin.Context, msg *message.Message) error {
	if ctx.Event.IsPrivate() {
		return b.SendPrivateMessage(ctx.Event.UserID, msg)
	}
	return b.SendGroupMessage(ctx.Event.GroupID, msg)
}

// GetLoginInfo gets bot login information
func (b *Bot) GetLoginInfo() (*plugin.LoginInfo, error) {
	resp, err := b.callAPIWithResponse("get_login_info", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data plugin.LoginInfo `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

// callAPI makes an API call to NapCat
func (b *Bot) callAPI(action string, params map[string]interface{}) error {
	_, err := b.callAPIWithResponse(action, params)
	return err
}

// callAPIWithResponse makes an API call and returns the response
func (b *Bot) callAPIWithResponse(action string, params map[string]interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", b.config.NapCat.HttpURL, action)

	var body io.Reader
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if b.config.NapCat.Token != "" {
		req.Header.Set("Authorization", "Bearer "+b.config.NapCat.Token)
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	return respBody, nil
}

// GetPluginManager returns the plugin manager
func (b *Bot) GetPluginManager() *plugin.Manager {
	return b.pluginManager
}
