package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
)

const defaultAddr = "http://127.0.0.1:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	addr := os.Getenv("BOT_ADMIN_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	switch os.Args[1] {
	case "list", "ls":
		listPlugins(addr)
	case "install", "i":
		if len(os.Args) < 3 {
			fmt.Println("Usage: botctl install [--start] <github_repo_url>")
			os.Exit(1)
		}
		// Parse --start flag
		autoStart := false
		repoURL := ""
		for _, arg := range os.Args[2:] {
			if arg == "--start" || arg == "-s" {
				autoStart = true
			} else if !strings.HasPrefix(arg, "-") {
				repoURL = arg
			}
		}
		if repoURL == "" {
			fmt.Println("Usage: botctl install [--start] <github_repo_url>")
			os.Exit(1)
		}
		installPlugin(addr, repoURL, autoStart)
	case "start":
		if len(os.Args) < 3 {
			fmt.Println("Usage: botctl start <plugin_name>")
			os.Exit(1)
		}
		startPlugin(addr, os.Args[2])
	case "stop":
		if len(os.Args) < 3 {
			fmt.Println("Usage: botctl stop <plugin_name>")
			os.Exit(1)
		}
		stopPlugin(addr, os.Args[2])
	case "uninstall", "rm":
		if len(os.Args) < 3 {
			fmt.Println("Usage: botctl uninstall <plugin_name>")
			os.Exit(1)
		}
		uninstallPlugin(addr, os.Args[2])
	case "health":
		checkHealth(addr)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`botctl - Bot Platform Plugin Manager

Usage:
  botctl <command> [arguments]

Commands:
  list, ls                      List all installed plugins
  install, i [--start] <url>    Install plugin from GitHub repo URL
                                --start, -s: Auto-start after install
  start <name>                  Start a plugin
  stop <name>                   Stop a running plugin
  uninstall, rm <name>          Uninstall a plugin
  health                        Check platform health
  help                          Show this help

Environment:
  BOT_ADMIN_ADDR        Admin API address (default: http://127.0.0.1:8080)

Examples:
  botctl list
  botctl install https://github.com/user/plugin-weather
  botctl install --start DaikonSushi/plugin-echo
  botctl start weather
  botctl stop weather
  botctl uninstall weather`)
}

func listPlugins(addr string) {
	resp, err := http.Get(addr + "/api/plugins")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    []struct {
			Name        string   `json:"name"`
			Version     string   `json:"version"`
			Description string   `json:"description"`
			Author      string   `json:"author"`
			Commands    []string `json:"commands"`
			Status      string   `json:"status"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if result.Code != 0 {
		fmt.Printf("Error: %s\n", result.Message)
		os.Exit(1)
	}

	if len(result.Data) == 0 {
		fmt.Println("No plugins installed.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tSTATUS\tCOMMANDS\tDESCRIPTION")
	fmt.Fprintln(w, "----\t-------\t------\t--------\t-----------")
	for _, p := range result.Data {
		cmds := ""
		for i, c := range p.Commands {
			if i > 0 {
				cmds += ", "
			}
			cmds += "/" + c
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", p.Name, p.Version, p.Status, cmds, p.Description)
	}
	w.Flush()
}

func installPlugin(addr, repoURL string, autoStart bool) {
	fmt.Printf("Installing plugin from %s...\n", repoURL)

	body, _ := json.Marshal(map[string]interface{}{"repo_url": repoURL, "auto_start": autoStart})
	resp, err := http.Post(addr+"/api/plugins/install", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	printInstallResult(resp.Body, autoStart)
}

func startPlugin(addr, name string) {
	fmt.Printf("Starting plugin %s...\n", name)

	body, _ := json.Marshal(map[string]string{"name": name})
	resp, err := http.Post(addr+"/api/plugins/start", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	printResult(resp.Body)
}

func stopPlugin(addr, name string) {
	fmt.Printf("Stopping plugin %s...\n", name)

	body, _ := json.Marshal(map[string]string{"name": name})
	resp, err := http.Post(addr+"/api/plugins/stop", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	printResult(resp.Body)
}

func uninstallPlugin(addr, name string) {
	fmt.Printf("Uninstalling plugin %s...\n", name)

	body, _ := json.Marshal(map[string]string{"name": name})
	resp, err := http.Post(addr+"/api/plugins/uninstall", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	printResult(resp.Body)
}

func checkHealth(addr string) {
	resp, err := http.Get(addr + "/api/health")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Status         string `json:"status"`
			RunningPlugins int    `json:"running_plugins"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Status: %s\n", result.Data.Status)
	fmt.Printf("Running Plugins: %d\n", result.Data.RunningPlugins)
}

func printResult(body io.Reader) {
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if result.Code != 0 {
		fmt.Printf("❌ Error: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Printf("✅ %s\n", result.Message)
}

func printInstallResult(body io.Reader, autoStart bool) {
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Started bool   `json:"started"`
		} `json:"data"`
	}

	if err := json.NewDecoder(body).Decode(&result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if result.Code != 0 {
		fmt.Printf("❌ Error: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Printf("✅ %s\n", result.Message)
	fmt.Printf("   Name: %s\n", result.Data.Name)
	fmt.Printf("   Version: %s\n", result.Data.Version)
	if autoStart {
		if result.Data.Started {
			fmt.Printf("   Status: 🟢 started\n")
		} else {
			fmt.Printf("   Status: 🔴 failed to start\n")
		}
	}
}
