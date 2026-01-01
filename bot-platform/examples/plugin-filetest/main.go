// plugin-filetest - A test plugin for file upload functionality
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DaikonSushi/bot-platform/pkg/pluginsdk"
)

// FileTestPlugin tests file upload functionality
type FileTestPlugin struct {
	bot *pluginsdk.BotClient
}

// Info returns plugin metadata
func (p *FileTestPlugin) Info() pluginsdk.PluginInfo {
	return pluginsdk.PluginInfo{
		Name:              "filetest",
		Version:           "1.0.0",
		Description:       "Test plugin for file upload functionality",
		Author:            "hovanzhang",
		Commands:          []string{"uploadgroup", "uploadprivate", "testfile"},
		HandleAllMessages: false,
	}
}

// OnStart is called when the plugin starts
func (p *FileTestPlugin) OnStart(bot *pluginsdk.BotClient) error {
	p.bot = bot
	bot.Log("info", "File test plugin started!")
	return nil
}

// OnStop is called when the plugin stops
func (p *FileTestPlugin) OnStop() error {
	return nil
}

// OnMessage handles incoming messages
func (p *FileTestPlugin) OnMessage(ctx context.Context, bot *pluginsdk.BotClient, msg *pluginsdk.Message) bool {
	return false
}

// OnCommand handles commands
func (p *FileTestPlugin) OnCommand(ctx context.Context, bot *pluginsdk.BotClient, cmd string, args []string, msg *pluginsdk.Message) bool {
	switch cmd {
	case "testfile":
		p.handleTestFile(bot, msg)
		return true
	case "uploadgroup":
		p.handleUploadGroup(bot, args, msg)
		return true
	case "uploadprivate":
		p.handleUploadPrivate(bot, args, msg)
		return true
	}
	return false
}

// handleTestFile creates a test file and uploads it
func (p *FileTestPlugin) handleTestFile(bot *pluginsdk.BotClient, msg *pluginsdk.Message) {
	// Create a temporary test file
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "test_upload.txt")
	
	content := fmt.Sprintf("This is a test file created at %s\nFor testing file upload functionality", 
		filepath.Base(testFile))
	
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		bot.Reply(msg, pluginsdk.Text(fmt.Sprintf("❌ Failed to create test file: %v", err)))
		return
	}
	
	bot.Reply(msg, pluginsdk.Text(fmt.Sprintf("✅ Test file created: %s\nUploading...", testFile)))
	
	// Upload to group or private based on message type
	if msg.GroupID > 0 {
		err = bot.UploadGroupFile(msg.GroupID, testFile, "test_upload.txt")
		if err != nil {
			bot.Reply(msg, pluginsdk.Text(fmt.Sprintf("❌ Upload failed: %v", err)))
		} else {
			bot.Reply(msg, pluginsdk.Text("✅ File uploaded to group successfully!"))
		}
	} else {
		err = bot.UploadPrivateFile(msg.UserID, testFile, "test_upload.txt")
		if err != nil {
			bot.Reply(msg, pluginsdk.Text(fmt.Sprintf("❌ Upload failed: %v", err)))
		} else {
			bot.Reply(msg, pluginsdk.Text("✅ File uploaded to private chat successfully!"))
		}
	}
	
	// Clean up
	os.Remove(testFile)
}

// handleUploadGroup uploads a file to group
func (p *FileTestPlugin) handleUploadGroup(bot *pluginsdk.BotClient, args []string, msg *pluginsdk.Message) {
	if msg.GroupID == 0 {
		bot.Reply(msg, pluginsdk.Text("❌ This command can only be used in groups"))
		return
	}
	
	if len(args) < 1 {
		bot.Reply(msg, 
			pluginsdk.Text("📤 Upload Group File\n"),
			pluginsdk.Text("━━━━━━━━━━━━━━\n"),
			pluginsdk.Text("用法:\n"),
			pluginsdk.Text("  /uploadgroup <文件路径> [显示名称] [文件夹]\n"),
			pluginsdk.Text("\n示例:\n"),
			pluginsdk.Text("  /uploadgroup /tmp/test.txt\n"),
			pluginsdk.Text("  /uploadgroup /tmp/test.txt myfile.txt\n"),
			pluginsdk.Text("  /uploadgroup /tmp/test.txt myfile.txt /documents"),
		)
		return
	}
	
	filePath := args[0]
	fileName := filepath.Base(filePath)
	if len(args) > 1 {
		fileName = args[1]
	}
	
	folder := "/"
	if len(args) > 2 {
		folder = args[2]
	}
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		bot.Reply(msg, pluginsdk.Text(fmt.Sprintf("❌ File not found: %s", filePath)))
		return
	}
	
	bot.Reply(msg, pluginsdk.Text(fmt.Sprintf("📤 Uploading file: %s\nAs: %s\nTo folder: %s", filePath, fileName, folder)))
	
	err := bot.UploadGroupFile(msg.GroupID, filePath, fileName, folder)
	if err != nil {
		bot.Reply(msg, pluginsdk.Text(fmt.Sprintf("❌ Upload failed: %v", err)))
	} else {
		bot.Reply(msg, pluginsdk.Text("✅ File uploaded successfully!"))
	}
}

// handleUploadPrivate uploads a file to private chat
func (p *FileTestPlugin) handleUploadPrivate(bot *pluginsdk.BotClient, args []string, msg *pluginsdk.Message) {
	if msg.GroupID > 0 {
		bot.Reply(msg, pluginsdk.Text("❌ This command can only be used in private chats"))
		return
	}
	
	if len(args) < 1 {
		bot.Reply(msg, 
			pluginsdk.Text("📤 Upload Private File\n"),
			pluginsdk.Text("━━━━━━━━━━━━━━\n"),
			pluginsdk.Text("用法:\n"),
			pluginsdk.Text("  /uploadprivate <文件路径> [显示名称]\n"),
			pluginsdk.Text("\n示例:\n"),
			pluginsdk.Text("  /uploadprivate /tmp/test.txt\n"),
			pluginsdk.Text("  /uploadprivate /tmp/test.txt myfile.txt"),
		)
		return
	}
	
	filePath := args[0]
	fileName := filepath.Base(filePath)
	if len(args) > 1 {
		fileName = args[1]
	}
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		bot.Reply(msg, pluginsdk.Text(fmt.Sprintf("❌ File not found: %s", filePath)))
		return
	}
	
	bot.Reply(msg, pluginsdk.Text(fmt.Sprintf("📤 Uploading file: %s\nAs: %s", filePath, fileName)))
	
	err := bot.UploadPrivateFile(msg.UserID, filePath, fileName)
	if err != nil {
		bot.Reply(msg, pluginsdk.Text(fmt.Sprintf("❌ Upload failed: %v", err)))
	} else {
		bot.Reply(msg, pluginsdk.Text("✅ File uploaded successfully!"))
	}
}

func main() {
	pluginsdk.Run(&FileTestPlugin{})
}
