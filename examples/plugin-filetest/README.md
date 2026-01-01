# File Upload Test Plugin

这是一个用于测试 bot-platform 文件上传功能的测试插件。

## 功能

该插件提供了三个命令来测试文件上传功能：

### 1. `/testfile` - 自动测试
自动创建一个测试文件并上传到当前会话（群组或私聊）。

**用法：**
```
/testfile
```

### 2. `/uploadgroup` - 上传文件到群组
上传指定的文件到群组。

**用法：**
```
/uploadgroup <文件路径> [显示名称] [文件夹]
```

**示例：**
```
/uploadgroup /tmp/test.txt
/uploadgroup /tmp/test.txt myfile.txt
/uploadgroup /tmp/test.txt myfile.txt /documents
```

### 3. `/uploadprivate` - 上传文件到私聊
上传指定的文件到私聊。

**用法：**
```
/uploadprivate <文件路径> [显示名称]
```

**示例：**
```
/uploadprivate /tmp/test.txt
/uploadprivate /tmp/test.txt myfile.txt
```

## 编译

在 `examples/plugin-filetest` 目录下运行：

```bash
go build -o filetest-plugin .
```

或者使用交叉编译：

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o filetest-plugin_linux_amd64 .

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o filetest-plugin_linux_arm64 .

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o filetest-plugin_darwin_arm64 .
```

## 部署

1. 将编译好的插件复制到 bot-platform 的 `plugins-bin` 目录
2. 在 `plugins-config` 目录创建配置文件 `filetest.json`：

```json
{
  "name": "filetest",
  "enabled": true,
  "path": "./plugins-bin/filetest-plugin",
  "auto_start": true
}
```

3. 重启 bot-platform 或使用 `/plugin load filetest` 命令加载插件

## 测试步骤

### 测试群组文件上传

1. 在群组中发送 `/testfile` 命令
2. 插件会自动创建一个测试文件并上传到群组
3. 检查群组文件是否成功上传

或者手动测试：

1. 准备一个测试文件，例如 `/tmp/test.txt`
2. 在群组中发送 `/uploadgroup /tmp/test.txt`
3. 检查文件是否上传成功

### 测试私聊文件上传

1. 在私聊中发送 `/testfile` 命令
2. 插件会自动创建一个测试文件并上传到私聊
3. 检查私聊文件是否成功接收

或者手动测试：

1. 准备一个测试文件，例如 `/tmp/test.txt`
2. 在私聊中发送 `/uploadprivate /tmp/test.txt`
3. 检查文件是否上传成功

## 注意事项

- 文件路径必须是绝对路径
- 确保 bot 进程有权限读取指定的文件
- 群组文件上传支持指定文件夹路径
- 私聊文件上传不支持文件夹参数
