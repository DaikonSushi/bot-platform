package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	pb "github.com/DaikonSushi/bot-platform/api/proto"
	"github.com/DaikonSushi/bot-platform/internal/bot"
	"github.com/DaikonSushi/bot-platform/internal/botservice"
	"github.com/DaikonSushi/bot-platform/internal/config"
	"github.com/DaikonSushi/bot-platform/internal/pluginmgr"
	"github.com/DaikonSushi/bot-platform/internal/server"
	"github.com/DaikonSushi/bot-platform/plugins/echo"
	"github.com/DaikonSushi/bot-platform/plugins/help"
	"github.com/DaikonSushi/bot-platform/plugins/pluginctl"
)

func main() {
	// Parse flags
	configPath := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("[Main] Config loaded successfully")

	// Create bot instance
	b := bot.New(cfg)

	// Start BotService gRPC server for external plugins to call back
	grpcPort := cfg.PluginManager.GRPCPort
	botSvc := botservice.NewService(b)
	grpcServer := grpc.NewServer()
	pb.RegisterBotServiceServer(grpcServer, botSvc)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %d: %v", grpcPort, err)
	}

	go func() {
		log.Printf("[Main] BotService gRPC server starting on port %d", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("[Main] gRPC server error: %v", err)
		}
	}()

	// Initialize external plugin manager if enabled
	var extPluginMgr *pluginmgr.PluginManager
	if cfg.PluginManager.Enabled {
		// Create plugin directories
		os.MkdirAll(cfg.PluginManager.PluginDir, 0755)
		os.MkdirAll(cfg.PluginManager.ConfigDir, 0755)

		extPluginMgr = pluginmgr.NewPluginManager(
			cfg.PluginManager.PluginDir,
			cfg.PluginManager.ConfigDir,
			grpcPort,
		)

		// Load installed plugins
		if err := extPluginMgr.LoadInstalledPlugins(); err != nil {
			log.Printf("[Main] Warning: failed to load installed plugins: %v", err)
		}

		// Auto-start plugins
		if len(cfg.PluginManager.AutoStart) > 0 {
			extPluginMgr.AutoStartPlugins(context.Background(), cfg.PluginManager.AutoStart)
		}

		// Set external plugin manager to bot
		b.SetExternalPluginManager(extPluginMgr)

		log.Println("[Main] External plugin manager initialized")
	}

	// Register built-in plugins based on config
	enabledPlugins := make(map[string]bool)
	for _, name := range cfg.Plugins.Enabled {
		enabledPlugins[name] = true
	}

	// If no plugins specified in config, enable all by default
	if len(enabledPlugins) == 0 {
		enabledPlugins["echo"] = true
		enabledPlugins["help"] = true
		enabledPlugins["pluginctl"] = true
	}

	if enabledPlugins["echo"] {
		b.RegisterPlugin(echo.New())
		log.Println("[Main] Registered built-in plugin: echo")
	}

	if enabledPlugins["help"] {
		b.RegisterPlugin(help.New(b.GetPluginManager(), extPluginMgr, &cfg.Help))
		log.Println("[Main] Registered built-in plugin: help")
	}

	if enabledPlugins["pluginctl"] && extPluginMgr != nil {
		b.RegisterPlugin(pluginctl.New(extPluginMgr))
		log.Println("[Main] Registered built-in plugin: pluginctl")
	}

	// Start admin server if enabled
	if cfg.AdminServer.Enabled && extPluginMgr != nil {
		adminSrv := server.NewAdminServer(cfg.AdminServer.Addr, extPluginMgr)
		go func() {
			log.Printf("[Main] Admin server starting on %s", cfg.AdminServer.Addr)
			if err := adminSrv.Start(); err != nil {
				log.Printf("[Main] Admin server error: %v", err)
			}
		}()
	}

	// Start bot
	if err := b.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("[Main] Bot is running. Press Ctrl+C to stop.")
	<-sigChan

	// Graceful shutdown
	log.Println("[Main] Shutting down...")

	// Stop gRPC server
	grpcServer.GracefulStop()

	// Stop external plugins
	if extPluginMgr != nil {
		extPluginMgr.Shutdown()
	}

	b.Stop()
	log.Println("[Main] Goodbye!")
}
