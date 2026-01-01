package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DaikonSushi/bot-platform/internal/bot"
	"github.com/DaikonSushi/bot-platform/internal/config"
	"github.com/DaikonSushi/bot-platform/internal/pluginmgr"
	"github.com/DaikonSushi/bot-platform/internal/server"
	"github.com/DaikonSushi/bot-platform/plugins/echo"
	"github.com/DaikonSushi/bot-platform/plugins/help"
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

	// Register built-in plugins
	b.RegisterPlugin(echo.New())
	b.RegisterPlugin(help.New(b.GetPluginManager()))

	// Initialize external plugin manager if enabled
	var extPluginMgr *pluginmgr.PluginManager
	if cfg.PluginManager.Enabled {
		// Create plugin directories
		os.MkdirAll(cfg.PluginManager.PluginDir, 0755)
		os.MkdirAll(cfg.PluginManager.ConfigDir, 0755)

		extPluginMgr = pluginmgr.NewPluginManager(
			cfg.PluginManager.PluginDir,
			cfg.PluginManager.ConfigDir,
		)

		// Load installed plugins
		if err := extPluginMgr.LoadInstalledPlugins(); err != nil {
			log.Printf("[Main] Warning: failed to load installed plugins: %v", err)
		}

		// Auto-start plugins
		if len(cfg.PluginManager.AutoStart) > 0 {
			extPluginMgr.AutoStartPlugins(context.Background(), cfg.PluginManager.AutoStart)
		}

		log.Println("[Main] External plugin manager initialized")
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

	// Stop external plugins
	if extPluginMgr != nil {
		for _, p := range extPluginMgr.GetRunningPlugins() {
			extPluginMgr.StopPlugin(context.Background(), p.Info.Name)
		}
	}

	b.Stop()
	log.Println("[Main] Goodbye!")
}
