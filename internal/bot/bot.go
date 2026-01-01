package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	pb "github.com/DaikonSushi/bot-platform/api/proto"
	"github.com/DaikonSushi/bot-platform/internal/config"
	"github.com/DaikonSushi/bot-platform/internal/message"
	"github.com/DaikonSushi/bot-platform/internal/plugin"
	"github.com/DaikonSushi/bot-platform/internal/pluginmgr"
)

// Bot represents the main bot instance
type Bot struct {
	config           *config.Config
	httpClient       *http.Client
	wsConn           *websocket.Conn
	wsMu             sync.Mutex // Protects wsConn
	pluginManager    *plugin.Manager
	extPluginManager *pluginmgr.PluginManager
	running          bool
	mu               sync.RWMutex
	stopChan         chan struct{}
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

// SetExternalPluginManager sets the external plugin manager
func (b *Bot) SetExternalPluginManager(mgr *pluginmgr.PluginManager) {
	b.extPluginManager = mgr
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

	b.wsMu.Lock()
	if b.wsConn != nil {
		b.wsConn.Close()
		b.wsConn = nil
	}
	b.wsMu.Unlock()

	log.Println("[Bot] Bot stopped")
}

// connectWebSocket establishes WebSocket connection
func (b *Bot) connectWebSocket() error {
	b.wsMu.Lock()
	defer b.wsMu.Unlock()

	// Close existing connection if any
	if b.wsConn != nil {
		b.wsConn.Close()
		b.wsConn = nil
	}

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
	reconnectAttempt := 0
	maxReconnectDelay := 60 * time.Second
	baseDelay := 1 * time.Second

	for {
		select {
		case <-b.stopChan:
			return
		default:
			b.wsMu.Lock()
			conn := b.wsConn
			b.wsMu.Unlock()

			if conn == nil {
				// No connection, try to reconnect
				b.handleReconnect(&reconnectAttempt, baseDelay, maxReconnectDelay)
				continue
			}

			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[Bot] WebSocket read error: %v", err)
				b.handleReconnect(&reconnectAttempt, baseDelay, maxReconnectDelay)
				continue
			}

			// Reset reconnect counter on successful read
			reconnectAttempt = 0

			go b.handleEvent(msg)
		}
	}
}

// handleReconnect handles reconnection with exponential backoff
func (b *Bot) handleReconnect(attempt *int, baseDelay, maxDelay time.Duration) {
	// Calculate delay with exponential backoff
	delay := time.Duration(math.Min(
		float64(baseDelay)*math.Pow(2, float64(*attempt)),
		float64(maxDelay),
	))

	log.Printf("[Bot] Reconnecting in %v (attempt %d)...", delay, *attempt+1)

	select {
	case <-b.stopChan:
		return
	case <-time.After(delay):
	}

	if err := b.connectWebSocket(); err != nil {
		log.Printf("[Bot] Reconnect failed: %v", err)
		*attempt++
	} else {
		log.Printf("[Bot] Reconnected successfully")
		*attempt = 0
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

	// Dispatch to built-in plugin manager first
	handled := b.pluginManager.HandleEvent(ctx)

	// If not handled and external plugin manager exists, try external plugins
	if !handled && b.extPluginManager != nil {
		b.dispatchToExternalPlugins(ctx)
	}
}

// dispatchToExternalPlugins dispatches the event to external plugins
func (b *Bot) dispatchToExternalPlugins(ctx *plugin.Context) {
	// Check if it's a command
	text := strings.TrimSpace(ctx.Event.GetText())
	if !strings.HasPrefix(text, b.config.Bot.CommandPrefix) {
		// Not a command, dispatch as message to all external plugins
		pbEvent := b.convertToPbMessageEvent(ctx.Event)
		b.extPluginManager.DispatchMessage(context.Background(), pbEvent)
		return
	}

	// Parse command
	text = strings.TrimPrefix(text, b.config.Bot.CommandPrefix)
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	args := parts[1:]

	// Create command event
	cmdEvent := &pb.CommandEvent{
		Message: b.convertToPbMessageEvent(ctx.Event),
		Command: cmd,
		Args:    args,
	}

	// Dispatch to external plugin manager
	handled := b.extPluginManager.DispatchCommand(context.Background(), cmdEvent)
	if !handled {
		log.Printf("[Bot] Command '%s' not handled by any plugin", cmd)
	}
}

// convertToPbMessageEvent converts internal event to protobuf event
func (b *Bot) convertToPbMessageEvent(event *message.Event) *pb.MessageEvent {
	return &pb.MessageEvent{
		MessageId:   fmt.Sprintf("%d", event.MessageID),
		UserId:      event.UserID,
		GroupId:     event.GroupID,
		MessageType: string(event.MessageType),
		RawMessage:  event.RawMessage,
		Timestamp:   event.Time,
		Sender: &pb.UserInfo{
			UserId:   event.Sender.UserID,
			Nickname: event.Sender.Nickname,
		},
	}
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

// SendPrivateText sends a text message to a user (implements MessageSender interface)
func (b *Bot) SendPrivateText(userID int64, text string) error {
	return b.SendPrivateMessage(userID, message.NewMessage().Text(text))
}

// SendGroupText sends a text message to a group (implements MessageSender interface)
func (b *Bot) SendGroupText(groupID int64, text string) error {
	return b.SendGroupMessage(groupID, message.NewMessage().Text(text))
}
