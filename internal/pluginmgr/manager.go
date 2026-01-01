package pluginmgr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	pb "github.com/DaikonSushi/bot-platform/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PluginState represents the runtime state of a plugin
type PluginState struct {
	Info      *PluginMeta
	Process   *os.Process
	Client    pb.PluginServiceClient
	Conn      *grpc.ClientConn
	Port      int
	Status    string // "running", "stopped", "error"
	StartedAt time.Time
	LastError string
}

// PluginMeta represents plugin metadata
type PluginMeta struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Commands    []string `json:"commands"`
	RepoURL     string   `json:"repo_url"`    // GitHub repo URL
	BinaryName  string   `json:"binary_name"` // Binary file name
}

// PortPool manages reusable ports
type PortPool struct {
	mu        sync.Mutex
	available []int
	inUse     map[int]bool
	nextPort  int
	minPort   int
	maxPort   int
}

// NewPortPool creates a new port pool
func NewPortPool(minPort, maxPort int) *PortPool {
	return &PortPool{
		available: make([]int, 0),
		inUse:     make(map[int]bool),
		nextPort:  minPort,
		minPort:   minPort,
		maxPort:   maxPort,
	}
}

// Acquire gets an available port
func (p *PortPool) Acquire() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Try to reuse a released port first
	if len(p.available) > 0 {
		port := p.available[len(p.available)-1]
		p.available = p.available[:len(p.available)-1]
		p.inUse[port] = true
		return port, nil
	}

	// Allocate a new port
	if p.nextPort > p.maxPort {
		return 0, fmt.Errorf("port pool exhausted (min: %d, max: %d)", p.minPort, p.maxPort)
	}

	port := p.nextPort
	p.nextPort++
	p.inUse[port] = true
	return port, nil
}

// Release returns a port to the pool
func (p *PortPool) Release(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.inUse[port] {
		delete(p.inUse, port)
		p.available = append(p.available, port)
	}
}

// PluginManager manages plugin lifecycle
type PluginManager struct {
	mu           sync.RWMutex
	plugins      map[string]*PluginState
	pluginDir    string
	configDir    string
	portPool     *PortPool
	grpcPort     int // BotService gRPC port for plugins to call back
	botService   pb.BotServiceServer
	commandIndex map[string]string // command -> plugin name
	healthTicker *time.Ticker
	stopHealth   chan struct{}
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginDir, configDir string, grpcPort int) *PluginManager {
	pm := &PluginManager{
		plugins:      make(map[string]*PluginState),
		pluginDir:    pluginDir,
		configDir:    configDir,
		portPool:     NewPortPool(50100, 51000), // Allow up to 900 plugins
		grpcPort:     grpcPort,
		commandIndex: make(map[string]string),
		stopHealth:   make(chan struct{}),
	}

	// Start health check goroutine
	pm.startHealthCheck()

	return pm
}

// startHealthCheck starts periodic health checking for all running plugins
func (pm *PluginManager) startHealthCheck() {
	pm.healthTicker = time.NewTicker(30 * time.Second)

	go func() {
		for {
			select {
			case <-pm.stopHealth:
				pm.healthTicker.Stop()
				return
			case <-pm.healthTicker.C:
				pm.checkPluginHealth()
			}
		}
	}()
}

// checkPluginHealth checks health of all running plugins
func (pm *PluginManager) checkPluginHealth() {
	pm.mu.RLock()
	plugins := make([]*PluginState, 0)
	for _, state := range pm.plugins {
		if state.Status == "running" {
			plugins = append(plugins, state)
		}
	}
	pm.mu.RUnlock()

	for _, state := range plugins {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := state.Client.Health(ctx, &pb.Empty{})
		cancel()

		if err != nil {
			log.Printf("[PluginMgr] Plugin %s health check failed: %v", state.Info.Name, err)
			pm.handlePluginCrash(state.Info.Name)
		}
	}
}

// handlePluginCrash handles a crashed plugin
func (pm *PluginManager) handlePluginCrash(name string) {
	pm.mu.Lock()
	state, exists := pm.plugins[name]
	if !exists || state.Status != "running" {
		pm.mu.Unlock()
		return
	}

	// Mark as error
	state.Status = "error"
	state.LastError = "plugin crashed or became unresponsive"

	// Release port
	if state.Port > 0 {
		pm.portPool.Release(state.Port)
	}

	// Clean up connection
	if state.Conn != nil {
		state.Conn.Close()
	}

	// Remove command index
	for _, cmd := range state.Info.Commands {
		delete(pm.commandIndex, cmd)
	}

	pm.mu.Unlock()

	log.Printf("[PluginMgr] Plugin %s crashed, attempting restart...", name)

	// Attempt to restart the plugin
	go func() {
		time.Sleep(5 * time.Second)
		if err := pm.StartPlugin(context.Background(), name); err != nil {
			log.Printf("[PluginMgr] Failed to restart plugin %s: %v", name, err)
		} else {
			log.Printf("[PluginMgr] Successfully restarted plugin %s", name)
		}
	}()
}

// SetBotService sets the bot service for plugins to call back
func (pm *PluginManager) SetBotService(svc pb.BotServiceServer) {
	pm.botService = svc
}

// InstallFromGitHub downloads and installs a plugin from GitHub releases
func (pm *PluginManager) InstallFromGitHub(ctx context.Context, repoURL string) (*PluginMeta, error) {
	// Parse repo URL: https://github.com/owner/repo
	parts := strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid GitHub URL: %s", repoURL)
	}
	owner, repo := parts[0], parts[1]

	// Get latest release info
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, _ := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	// Find the right binary for current OS/arch
	osName := runtime.GOOS
	archName := runtime.GOARCH
	expectedSuffix := fmt.Sprintf("%s_%s", osName, archName)
	if osName == "windows" {
		expectedSuffix += ".exe"
	}

	var downloadURL, binaryName string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, expectedSuffix) {
			downloadURL = asset.BrowserDownloadURL
			binaryName = asset.Name
			break
		}
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("no binary found for %s/%s", osName, archName)
	}

	// Download binary
	log.Printf("[PluginMgr] Downloading %s...", binaryName)

	pluginPath := filepath.Join(pm.pluginDir, binaryName)
	if err := pm.downloadFile(ctx, downloadURL, pluginPath); err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}

	// Make executable
	if osName != "windows" {
		os.Chmod(pluginPath, 0755)
	}

	// Get plugin info by running with --info flag
	meta, err := pm.getPluginInfo(pluginPath)
	if err != nil {
		os.Remove(pluginPath)
		return nil, fmt.Errorf("failed to get plugin info: %w", err)
	}
	meta.RepoURL = repoURL
	meta.BinaryName = binaryName

	// Save plugin meta
	metaPath := filepath.Join(pm.configDir, meta.Name+".json")
	metaFile, _ := os.Create(metaPath)
	json.NewEncoder(metaFile).Encode(meta)
	metaFile.Close()

	log.Printf("[PluginMgr] Installed plugin: %s v%s", meta.Name, meta.Version)
	return meta, nil
}

// downloadFile downloads a file from URL
func (pm *PluginManager) downloadFile(ctx context.Context, url, dest string) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// getPluginInfo runs the plugin binary with --info to get metadata
func (pm *PluginManager) getPluginInfo(binaryPath string) (*PluginMeta, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "--info")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var meta PluginMeta
	if err := json.Unmarshal(output, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// StartPlugin starts a plugin by name
func (pm *PluginManager) StartPlugin(ctx context.Context, name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if already running
	if state, exists := pm.plugins[name]; exists && state.Status == "running" {
		return fmt.Errorf("plugin %s is already running", name)
	}

	// Load plugin meta
	metaPath := filepath.Join(pm.configDir, name+".json")
	metaFile, err := os.Open(metaPath)
	if err != nil {
		return fmt.Errorf("plugin %s not found, install it first", name)
	}
	defer metaFile.Close()

	var meta PluginMeta
	if err := json.NewDecoder(metaFile).Decode(&meta); err != nil {
		return fmt.Errorf("invalid plugin meta: %w", err)
	}

	// Find binary
	binaryPath := filepath.Join(pm.pluginDir, meta.BinaryName)
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin binary not found: %s", binaryPath)
	}

	// Allocate port from pool
	port, err := pm.portPool.Acquire()
	if err != nil {
		return fmt.Errorf("failed to allocate port: %w", err)
	}

	// Start plugin process
	cmd := exec.Command(binaryPath,
		"--port", fmt.Sprintf("%d", port),
		"--core-addr", fmt.Sprintf("127.0.0.1:%d", pm.grpcPort),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		pm.portPool.Release(port)
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	// Wait for plugin to be ready with retries
	var conn *grpc.ClientConn
	var client pb.PluginServiceClient

	maxRetries := 10
	retryInterval := 200 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		time.Sleep(retryInterval)

		// Use context with timeout instead of deprecated grpc.WithTimeout
		dialCtx, dialCancel := context.WithTimeout(ctx, 2*time.Second)
		conn, err = grpc.DialContext(
			dialCtx,
			fmt.Sprintf("127.0.0.1:%d", port),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		dialCancel()

		if err != nil {
			if i == maxRetries-1 {
				cmd.Process.Kill()
				pm.portPool.Release(port)
				return fmt.Errorf("failed to connect to plugin after %d retries: %w", maxRetries, err)
			}
			continue
		}

		client = pb.NewPluginServiceClient(conn)

		// Verify connection with health check
		healthCtx, healthCancel := context.WithTimeout(ctx, 2*time.Second)
		healthResp, err := client.Health(healthCtx, &pb.Empty{})
		healthCancel()

		if err != nil || !healthResp.Healthy {
			conn.Close()
			if i == maxRetries-1 {
				cmd.Process.Kill()
				pm.portPool.Release(port)
				return fmt.Errorf("plugin health check failed after %d retries", maxRetries)
			}
			continue
		}

		// Success!
		break
	}

	// Register plugin
	state := &PluginState{
		Info:      &meta,
		Process:   cmd.Process,
		Client:    client,
		Conn:      conn,
		Port:      port,
		Status:    "running",
		StartedAt: time.Now(),
	}
	pm.plugins[name] = state

	// Index commands
	for _, cmd := range meta.Commands {
		pm.commandIndex[cmd] = name
	}

	log.Printf("[PluginMgr] Started plugin: %s on port %d", name, port)
	return nil
}

// StopPlugin stops a running plugin
func (pm *PluginManager) StopPlugin(ctx context.Context, name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	state, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if state.Status != "running" {
		return fmt.Errorf("plugin %s is not running", name)
	}

	// Send shutdown command
	if state.Client != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		state.Client.Shutdown(shutdownCtx, &pb.Empty{})
		cancel()
	}

	// Wait for process to exit gracefully
	done := make(chan struct{})
	go func() {
		if state.Process != nil {
			state.Process.Wait()
		}
		close(done)
	}()

	select {
	case <-done:
		// Process exited gracefully
	case <-time.After(5 * time.Second):
		// Timeout, force kill
		if state.Process != nil {
			state.Process.Kill()
			state.Process.Wait()
		}
	}

	// Close connection
	if state.Conn != nil {
		state.Conn.Close()
	}

	// Release port back to pool
	if state.Port > 0 {
		pm.portPool.Release(state.Port)
	}

	// Remove command index
	for _, cmd := range state.Info.Commands {
		delete(pm.commandIndex, cmd)
	}

	state.Status = "stopped"
	log.Printf("[PluginMgr] Stopped plugin: %s", name)
	return nil
}

// Shutdown stops all plugins and cleans up
func (pm *PluginManager) Shutdown() {
	// Stop health check
	close(pm.stopHealth)

	// Stop all running plugins
	pm.mu.RLock()
	names := make([]string, 0)
	for name, state := range pm.plugins {
		if state.Status == "running" {
			names = append(names, name)
		}
	}
	pm.mu.RUnlock()

	for _, name := range names {
		pm.StopPlugin(context.Background(), name)
	}
}

// UninstallPlugin removes a plugin
func (pm *PluginManager) UninstallPlugin(ctx context.Context, name string) error {
	// Stop if running
	pm.StopPlugin(ctx, name)

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Load meta to get binary name
	metaPath := filepath.Join(pm.configDir, name+".json")
	metaFile, err := os.Open(metaPath)
	if err == nil {
		var meta PluginMeta
		json.NewDecoder(metaFile).Decode(&meta)
		metaFile.Close()

		// Remove binary
		binaryPath := filepath.Join(pm.pluginDir, meta.BinaryName)
		os.Remove(binaryPath)
	}

	// Remove meta file
	os.Remove(metaPath)

	// Remove from map
	delete(pm.plugins, name)

	log.Printf("[PluginMgr] Uninstalled plugin: %s", name)
	return nil
}

// ListPlugins returns all installed plugins
func (pm *PluginManager) ListPlugins() []*PluginState {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Load all installed plugins from config dir
	files, _ := filepath.Glob(filepath.Join(pm.configDir, "*.json"))

	result := make([]*PluginState, 0)
	for _, f := range files {
		name := strings.TrimSuffix(filepath.Base(f), ".json")

		if state, exists := pm.plugins[name]; exists {
			result = append(result, state)
		} else {
			// Plugin installed but not running
			metaFile, _ := os.Open(f)
			var meta PluginMeta
			json.NewDecoder(metaFile).Decode(&meta)
			metaFile.Close()

			result = append(result, &PluginState{
				Info:   &meta,
				Status: "stopped",
			})
		}
	}
	return result
}

// GetPluginByCommand finds plugin that handles a command
func (pm *PluginManager) GetPluginByCommand(cmd string) *PluginState {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pluginName, exists := pm.commandIndex[cmd]
	if !exists {
		return nil
	}
	return pm.plugins[pluginName]
}

// GetRunningPlugins returns all running plugins
func (pm *PluginManager) GetRunningPlugins() []*PluginState {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]*PluginState, 0)
	for _, state := range pm.plugins {
		if state.Status == "running" {
			result = append(result, state)
		}
	}
	return result
}

// GetAllCommands returns all registered commands from running plugins
func (pm *PluginManager) GetAllCommands() map[string]*PluginMeta {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]*PluginMeta)
	for cmd, pluginName := range pm.commandIndex {
		if state, exists := pm.plugins[pluginName]; exists && state.Status == "running" {
			result[cmd] = state.Info
		}
	}
	return result
}

// DispatchMessage dispatches a message to all running plugins
func (pm *PluginManager) DispatchMessage(ctx context.Context, event *pb.MessageEvent) {
	pm.mu.RLock()
	plugins := make([]*PluginState, 0)
	for _, state := range pm.plugins {
		if state.Status == "running" {
			plugins = append(plugins, state)
		}
	}
	pm.mu.RUnlock()

	for _, plugin := range plugins {
		go func(p *PluginState) {
			_, err := p.Client.OnMessage(ctx, event)
			if err != nil {
				log.Printf("[PluginMgr] Plugin %s OnMessage error: %v", p.Info.Name, err)
			}
		}(plugin)
	}
}

// DispatchCommand dispatches a command to the appropriate plugin
func (pm *PluginManager) DispatchCommand(ctx context.Context, event *pb.CommandEvent) bool {
	log.Printf("[PluginMgr] DispatchCommand: looking for command '%s', commandIndex: %v", event.Command, pm.commandIndex)
	plugin := pm.GetPluginByCommand(event.Command)
	if plugin == nil {
		log.Printf("[PluginMgr] No plugin found for command: %s", event.Command)
		return false
	}

	log.Printf("[PluginMgr] Dispatching command '%s' to plugin '%s'", event.Command, plugin.Info.Name)
	result, err := plugin.Client.OnCommand(ctx, event)
	if err != nil {
		log.Printf("[PluginMgr] Plugin %s OnCommand error: %v", plugin.Info.Name, err)
		return false
	}
	return result.Handled
}

// LoadInstalledPlugins loads all installed plugins at startup
func (pm *PluginManager) LoadInstalledPlugins() error {
	files, err := filepath.Glob(filepath.Join(pm.configDir, "*.json"))
	if err != nil {
		return err
	}

	for _, f := range files {
		metaFile, err := os.Open(f)
		if err != nil {
			continue
		}

		var meta PluginMeta
		if err := json.NewDecoder(metaFile).Decode(&meta); err != nil {
			metaFile.Close()
			continue
		}
		metaFile.Close()

		// Register in plugins map (but don't start)
		pm.plugins[meta.Name] = &PluginState{
			Info:   &meta,
			Status: "stopped",
		}
	}
	return nil
}

// AutoStartPlugins starts all plugins marked for auto-start
func (pm *PluginManager) AutoStartPlugins(ctx context.Context, autoStart []string) {
	for _, name := range autoStart {
		if err := pm.StartPlugin(ctx, name); err != nil {
			log.Printf("[PluginMgr] Failed to auto-start %s: %v", name, err)
		}
	}
}
