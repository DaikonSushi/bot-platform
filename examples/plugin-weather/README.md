# Bot Plugin Weather

A weather plugin for bot-platform.

## Build

```bash
# Build for your platform
go build -o weather-plugin .

# Cross compile for different platforms
GOOS=linux GOARCH=amd64 go build -o weather-plugin_linux_amd64 .
GOOS=darwin GOARCH=amd64 go build -o weather-plugin_darwin_amd64 .
GOOS=darwin GOARCH=arm64 go build -o weather-plugin_darwin_arm64 .
GOOS=windows GOARCH=amd64 go build -o weather-plugin_windows_amd64.exe .
```

## Usage

Commands:
- `/weather <city>` - Get weather for a city
- `/天气 <城市>` - 获取城市天气

## Release

1. Tag a new version:
```bash
git tag v1.0.0
git push origin v1.0.0
```

2. Create a GitHub release and upload the compiled binaries.

3. In your bot-platform, install via:
```bash
curl -X POST http://localhost:8080/api/plugins/install \
  -H "Content-Type: application/json" \
  -d '{"repo_url": "https://github.com/your-username/plugin-weather"}'
```
