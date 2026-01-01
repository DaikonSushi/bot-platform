// Package pluginsdk provides the SDK for developing bot plugins
// Plugins should import this package to create plugins that work with the bot platform
package pluginsdk

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/DaikonSushi/bot-platform/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Plugin is the interface that all plugins must implement
type Plugin interface {
	// Info returns plugin metadata
	Info() PluginInfo

	// OnMessage is called for every incoming message
	OnMessage(ctx context.Context, bot *BotClient, msg *Message) bool

	// OnCommand is called when a registered command is triggered
	OnCommand(ctx context.Context, bot *BotClient, cmd string, args []string, msg *Message) bool

	// OnStart is called when the plugin starts
	OnStart(bot *BotClient) error

	// OnStop is called when the plugin stops
	OnStop() error
}

// PluginInfo contains plugin metadata
type PluginInfo struct {
	Name              string   `json:"name"`
	Version           string   `json:"version"`
	Description       string   `json:"description"`
	Author            string   `json:"author"`
	Commands          []string `json:"commands"`
	HandleAllMessages bool     `json:"handle_all_messages"`
}

// Message represents an incoming message
type Message struct {
	ID        string
	UserID    int64
	GroupID   int64  // 0 for private messages
	Type      string // "private" or "group"
	Text      string // Raw text content
	Segments  []MessageSegment
	Timestamp int64
	Sender    *UserInfo
}

// MessageSegment represents a message segment
type MessageSegment struct {
	Type string
	Data map[string]string
}

// UserInfo contains user information
type UserInfo struct {
	UserID   int64
	Nickname string
	Card     string
	Role     string
}

// BotClient provides methods to interact with the bot
type BotClient struct {
	client pb.BotServiceClient
}

// SendPrivateMessage sends a message to a user
func (b *BotClient) SendPrivateMessage(userID int64, segments ...MessageSegment) (int64, error) {
	pbSegs := make([]*pb.MessageSegment, len(segments))
	for i, seg := range segments {
		pbSegs[i] = &pb.MessageSegment{
			Type: seg.Type,
			Data: seg.Data,
		}
	}

	resp, err := b.client.SendMessage(context.Background(), &pb.SendMessageRequest{
		MessageType: "private",
		UserId:      userID,
		Segments:    pbSegs,
	})
	if err != nil {
		return 0, err
	}
	if resp.Error != "" {
		return 0, fmt.Errorf("%s", resp.Error)
	}
	return resp.MessageId, nil
}

// SendGroupMessage sends a message to a group
func (b *BotClient) SendGroupMessage(groupID int64, segments ...MessageSegment) (int64, error) {
	pbSegs := make([]*pb.MessageSegment, len(segments))
	for i, seg := range segments {
		pbSegs[i] = &pb.MessageSegment{
			Type: seg.Type,
			Data: seg.Data,
		}
	}

	resp, err := b.client.SendMessage(context.Background(), &pb.SendMessageRequest{
		MessageType: "group",
		GroupId:     groupID,
		Segments:    pbSegs,
	})
	if err != nil {
		return 0, err
	}
	if resp.Error != "" {
		return 0, fmt.Errorf("%s", resp.Error)
	}
	return resp.MessageId, nil
}

// Reply replies to a message (auto-detect private/group)
func (b *BotClient) Reply(msg *Message, segments ...MessageSegment) (int64, error) {
	if msg.Type == "group" {
		return b.SendGroupMessage(msg.GroupID, segments...)
	}
	return b.SendPrivateMessage(msg.UserID, segments...)
}

// GetUserInfo gets user information
func (b *BotClient) GetUserInfo(userID int64) (*UserInfo, error) {
	resp, err := b.client.GetUserInfo(context.Background(), &pb.GetUserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}
	return &UserInfo{
		UserID:   resp.UserId,
		Nickname: resp.Nickname,
		Card:     resp.Card,
		Role:     resp.Role,
	}, nil
}

// Log logs a message via the bot platform
func (b *BotClient) Log(level, message string) {
	b.client.Log(context.Background(), &pb.LogRequest{
		Level:   level,
		Message: message,
	})
}

// Text creates a text message segment
func Text(content string) MessageSegment {
	return MessageSegment{
		Type: "text",
		Data: map[string]string{"text": content},
	}
}

// Image creates an image message segment
func Image(url string) MessageSegment {
	return MessageSegment{
		Type: "image",
		Data: map[string]string{"url": url},
	}
}

// ImageFile creates an image message segment from file path
func ImageFile(path string) MessageSegment {
	return MessageSegment{
		Type: "image",
		Data: map[string]string{"file": "file://" + path},
	}
}

// At creates an at message segment
func At(userID int64) MessageSegment {
	return MessageSegment{
		Type: "at",
		Data: map[string]string{"qq": fmt.Sprintf("%d", userID)},
	}
}

// AtAll creates an at-all message segment
func AtAll() MessageSegment {
	return MessageSegment{
		Type: "at",
		Data: map[string]string{"qq": "all"},
	}
}

// Face creates a QQ face emoji segment
func Face(id int) MessageSegment {
	return MessageSegment{
		Type: "face",
		Data: map[string]string{"id": fmt.Sprintf("%d", id)},
	}
}

// pluginServer implements the gRPC PluginService
type pluginServer struct {
	pb.UnimplementedPluginServiceServer
	plugin Plugin
	bot    *BotClient
}

func (s *pluginServer) GetInfo(ctx context.Context, _ *pb.Empty) (*pb.PluginInfo, error) {
	info := s.plugin.Info()
	return &pb.PluginInfo{
		Name:              info.Name,
		Version:           info.Version,
		Description:       info.Description,
		Author:            info.Author,
		Commands:          info.Commands,
		HandleAllMessages: info.HandleAllMessages,
	}, nil
}

func (s *pluginServer) OnMessage(ctx context.Context, event *pb.MessageEvent) (*pb.HandleResult, error) {
	msg := convertMessage(event)
	handled := s.plugin.OnMessage(ctx, s.bot, msg)
	return &pb.HandleResult{Handled: handled}, nil
}

func (s *pluginServer) OnCommand(ctx context.Context, event *pb.CommandEvent) (*pb.HandleResult, error) {
	msg := convertMessage(event.Message)
	handled := s.plugin.OnCommand(ctx, s.bot, event.Command, event.Args, msg)
	return &pb.HandleResult{Handled: handled}, nil
}

func (s *pluginServer) Health(ctx context.Context, _ *pb.Empty) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Healthy: true,
		Status:  "running",
	}, nil
}

func (s *pluginServer) Shutdown(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	go func() {
		s.plugin.OnStop()
		os.Exit(0)
	}()
	return &pb.Empty{}, nil
}

func convertMessage(event *pb.MessageEvent) *Message {
	segments := make([]MessageSegment, len(event.Segments))
	for i, seg := range event.Segments {
		segments[i] = MessageSegment{
			Type: seg.Type,
			Data: seg.Data,
		}
	}

	var sender *UserInfo
	if event.Sender != nil {
		sender = &UserInfo{
			UserID:   event.Sender.UserId,
			Nickname: event.Sender.Nickname,
			Card:     event.Sender.Card,
			Role:     event.Sender.Role,
		}
	}

	return &Message{
		ID:        event.MessageId,
		UserID:    event.UserId,
		GroupID:   event.GroupId,
		Type:      event.MessageType,
		Text:      event.RawMessage,
		Segments:  segments,
		Timestamp: event.Timestamp,
		Sender:    sender,
	}
}

// Run starts the plugin and connects to the bot platform
func Run(plugin Plugin) {
	var (
		port     int
		coreAddr string
		showInfo bool
	)

	flag.IntVar(&port, "port", 50100, "Port to listen on")
	flag.StringVar(&coreAddr, "core-addr", "127.0.0.1:50051", "Bot core gRPC address")
	flag.BoolVar(&showInfo, "info", false, "Print plugin info as JSON and exit")
	flag.Parse()

	// If --info flag is set, print plugin info and exit
	if showInfo {
		info := plugin.Info()
		data, _ := json.Marshal(info)
		fmt.Println(string(data))
		os.Exit(0)
	}

	// Connect to bot core
	conn, err := grpc.Dial(coreAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to bot core: %v", err)
	}
	defer conn.Close()

	botClient := &BotClient{
		client: pb.NewBotServiceClient(conn),
	}

	// Initialize plugin
	if err := plugin.OnStart(botClient); err != nil {
		log.Fatalf("Plugin start failed: %v", err)
	}

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPluginServiceServer(grpcServer, &pluginServer{
		plugin: plugin,
		bot:    botClient,
	})

	log.Printf("[Plugin:%s] Started on port %d", plugin.Info().Name, port)

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Printf("[Plugin:%s] Shutting down...", plugin.Info().Name)
		plugin.OnStop()
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}
