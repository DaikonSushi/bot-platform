# Plugin Development

External plugins are standalone Go binaries. The platform downloads a release asset, runs it with `--info` to read metadata, then starts it as a gRPC plugin process when an admin runs `/plugin start <name>`.

## Minimal Workflow

1. Start from an existing template:

```bash
git clone https://github.com/DaikonSushi/plugin-fileupload.git plugin-myplugin
cd plugin-myplugin
```

2. Change `go.mod` and the plugin metadata returned by `Info()`.

3. Implement the plugin interface from `github.com/DaikonSushi/bot-platform/pkg/pluginsdk`:

```go
Info() pluginsdk.PluginInfo
OnStart(bot *pluginsdk.BotClient) error
OnStop() error
OnMessage(ctx context.Context, bot *pluginsdk.BotClient, msg *pluginsdk.Message) bool
OnCommand(ctx context.Context, bot *pluginsdk.BotClient, cmd string, args []string, msg *pluginsdk.Message) bool
```

4. Validate locally before publishing:

```bash
go test ./...
go build -o my-plugin .
./my-plugin --info
```

The `--info` output must be JSON with at least:

```json
{
  "name": "myplugin",
  "version": "1.0.0",
  "description": "Short description",
  "author": "DaikonSushi",
  "commands": ["myplugin"],
  "handle_all_messages": false,
  "message_priority": 0,
  "fallback": false
}
```

5. Publish a GitHub release containing a binary named with the target platform suffix, for example `my-plugin_linux_amd64`.

6. Install and operate it from QQ or `botctl`:

```text
/plugin install https://github.com/DaikonSushi/plugin-myplugin
/plugin start myplugin
/plugin update myplugin
/plugin reload myplugin
/plugin list
```

```bash
botctl install --start https://github.com/DaikonSushi/plugin-myplugin
botctl update myplugin
botctl restart myplugin
```

## Agent Plugins

Agent-style plugins should set `handle_all_messages` to `true`, `fallback` to `true`, and usually expose a command such as `agent` or `chat` for explicit control. The platform dispatches non-fallback plugins first, ordered by higher `message_priority`, then fallback plugins. This lets a broad conversational plugin stay available without stealing messages from purpose-built plugins.

Recommended behavior:

- Only answer normal chat when the message clearly mentions the bot, replies to the bot, or matches your plugin's configured chat policy.
- Keep `/agent` commands for provider selection, model settings, memory reset, and diagnostics.
- Load API keys and provider URLs from the plugin's own config file or environment variables, not from source code.
- Return `false` from `OnMessage` when the agent decides not to answer.
- Log provider errors through `bot.Log("error", "...")` so `docker logs qq-bot-all-in-one` can show the failure.

## Useful SDK Calls

```go
bot.Reply(msg, pluginsdk.Text("hello"))
bot.SendGroupMessage(groupID, pluginsdk.Text("hello group"))
bot.SendPrivateMessage(userID, pluginsdk.Text("hello user"))
bot.UploadGroupFile(groupID, "/tmp/report.txt", "report.txt")
bot.CallAPI("get_group_info", map[string]string{"group_id": "123456"})
```

## Release Checklist

- `go test ./...` passes.
- `go build -o <binary> .` succeeds.
- `./<binary> --info` prints valid JSON.
- The release asset name contains the target runtime suffix such as `linux_amd64`.
- The plugin starts with `/plugin start <name>`.
- `/plugin list` shows the expected commands and status.
