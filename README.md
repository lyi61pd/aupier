# Aupier - 看电影学英语音频截取工具

一个用 Go 开发的 Windows 桌面应用，帮助你在看电影学英语时快速截取音频片段并循环播放。

## 功能特性

- ✅ **全局热键控制** - 无需切换窗口即可录音
- ✅ **快速录音** - 按下热键开始/停止录音
- ✅ **自动播放** - 录音完成后自动播放
- ✅ **循环播放** - 支持重复播放某段音频
- ✅ **音量调节** - 可调节播放音量
- ✅ **片段管理** - 列表显示所有录音片段
- ✅ **WAV 格式** - 无损音质（16bit PCM）

## 技术栈

- **语言**: Go 1.21+
- **GUI**: Fyne v2 - 跨平台图形界面框架
- **音频**: PortAudio - 专业音频 I/O 库
- **热键**: Windows API - 全局热键注册

## 系统要求

- Windows 10/11 (64位)
- PortAudio 运行时库

## 安装依赖

### 在 Windows 上开发

#### 1. 安装 Go

下载并安装 Go 1.21 或更高版本：https://go.dev/dl/

#### 2. 安装 PortAudio

**方式一：使用 MSYS2 (推荐)**

```bash
# 安装 MSYS2: https://www.msys2.org/

# 在 MSYS2 终端中执行
pacman -S mingw-w64-x86_64-portaudio
pacman -S mingw-w64-x86_64-gcc

# 将 MSYS2 的 mingw64/bin 添加到系统 PATH
# 例如: C:\msys64\mingw64\bin
```

**方式二：使用预编译 DLL**

从 PortAudio 官网下载预编译的 DLL 文件，放到系统 PATH 或程序目录。

#### 3. 下载 Go 依赖

```bash
cd aupier
go mod tidy
```

### 在 Linux 上交叉编译到 Windows

#### 1. 安装交叉编译工具链

**Ubuntu/Debian:**

```bash
# 安装 MinGW-w64 交叉编译器
sudo apt-get update
sudo apt-get install -y gcc-mingw-w64-x86-64 g++-mingw-w64-x86-64

# 安装 PortAudio 开发库（本地编译用）
sudo apt-get install -y portaudio19-dev

# 下载 Windows 版 PortAudio（用于交叉编译）
# 需要手动下载或使用 MSYS2 的包
```

**Arch Linux:**

```bash
# 安装 MinGW-w64 工具链
sudo pacman -S mingw-w64-gcc

# 安装 PortAudio
sudo pacman -S portaudio

# 安装 Windows 版 PortAudio（AUR）
yay -S mingw-w64-portaudio
```

**使用 Docker（最简单）:**

```bash
# 使用预配置的交叉编译环境
docker pull ghcr.io/fyne-io/fyne-cross:latest

# 在项目目录运行
docker run --rm -v $(pwd):/app \
  -w /app \
  ghcr.io/fyne-io/fyne-cross:latest \
  windows-amd64
```

#### 2. 配置 CGO 环境变量

```bash
# 设置交叉编译环境
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++

# 设置 PortAudio 的头文件和库路径（如果需要）
export CGO_CFLAGS="-I/path/to/portaudio/include"
export CGO_LDFLAGS="-L/path/to/portaudio/lib -lportaudio"
```

#### 3. 下载 Go 依赖

```bash
cd aupier
go mod tidy
```

#### 4. 使用 Fyne-Cross（推荐）

这是最简单的方式，自动处理所有交叉编译依赖：

```bash
# 安装 fyne-cross
go install github.com/fyne-io/fyne-cross@latest

# 交叉编译到 Windows
fyne-cross windows -arch=amd64

# 生成的 exe 在 fyne-cross/dist/windows-amd64/ 目录
```

## 构建项目

### 在 Windows 上直接构建

```bash
go build -o aupier.exe ./cmd/app
```

### 在 Linux 上交叉编译

**方法一：使用 fyne-cross（推荐）**

```bash
# 一键交叉编译，自动处理所有依赖
fyne-cross windows -arch=amd64

# 输出文件在: fyne-cross/dist/windows-amd64/aupier.exe
```

**方法二：手动交叉编译**

```bash
# 设置环境变量
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc

# 编译
go build -o aupier.exe ./cmd/app
```

### 发布构建（优化大小）

```bash
# Windows 上
go build -ldflags="-s -w -H windowsgui" -o aupier.exe ./cmd/app

# Linux 交叉编译
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags="-s -w -H windowsgui" -o aupier.exe ./cmd/app
```

**说明:**
- `-s -w`: 去除调试信息，减小文件大小
- `-H windowsgui`: 隐藏控制台窗口（GUI 程序）

## 运行程序

```bash
./aupier.exe
```

或直接双击 `aupier.exe`

## 使用说明

### 快捷键

- **Ctrl+Shift+R** - 开始/停止录音
- **Ctrl+Shift+P** - 播放最近的录音

### 录音流程

1. 打开程序
2. 播放你想学习的电影/视频
3. 按 `Ctrl+Shift+R` 开始录音
4. 再按一次 `Ctrl+Shift+R` 停止录音
5. 录音自动保存并播放
6. 勾选"循环播放"可以重复听

### 界面操作

- **播放** - 播放选中的音频片段
- **停止** - 停止当前播放
- **循环播放** - 勾选后片段会自动循环
- **音量** - 调节播放音量（0-2倍）
- **删除所选** - 删除选中的音频片段
- **打开文件夹** - 在资源管理器中查看录音文件

## 配置文件

程序会在当前目录创建 `config.json`：

```json
{
  "output_dir": "./recordings",
  "record_hotkey": "Ctrl+Shift+R",
  "play_last_clip_hotkey": "Ctrl+Shift+P",
  "sample_rate": 44100,
  "channels": 2
}
```

可以手动修改配置后重启程序。

## 录音源说明

### 当前实现：麦克风录音

程序目前使用**默认麦克风**作为录音源。这是 PortAudio 的默认行为。

### 如何录制系统声音（高级）

要录制电影/视频的声音（而不是麦克风），有以下方案：

#### 方案一：使用虚拟音频设备

1. 安装虚拟音频设备（如 [VB-CABLE](https://vb-audio.com/Cable/)）
2. 在 Windows 声音设置中：
   - 将系统默认播放设备设为 "CABLE Input"
   - 在程序中选择 "CABLE Output" 作为录音设备
3. 使用扬声器监听：将 CABLE Output 设置为"侦听此设备"

#### 方案二：修改代码使用 WASAPI Loopback

在 `internal/audio/recorder.go` 的 `Start()` 方法中：

```go
// 当前代码使用默认输入设备
stream, err := portaudio.OpenDefaultStream(...)

// 改为使用 WASAPI Loopback 需要：
// 1. 枚举音频设备找到 loopback 设备
// 2. 使用 OpenStream 而非 OpenDefaultStream
// 3. 指定 WASAPI Host API
```

参考 PortAudio 文档：https://github.com/gordonklaus/portaudio

#### 方案三：Windows 立体声混音

某些声卡支持"立体声混音"功能：

1. 打开 Windows 声音设置
2. 在"录音"选项卡中启用"立体声混音"
3. 将其设为默认录音设备
4. 程序会自动使用该设备

## 项目结构

```
aupier/
├── cmd/app/           # 主程序入口
│   └── main.go
├── internal/
│   ├── audio/         # 音频录制和播放
│   │   ├── recorder.go
│   │   └── player.go
│   ├── config/        # 配置管理
│   │   └── config.go
│   ├── hotkey/        # 全局热键
│   │   └── hotkey.go
│   └── ui/            # GUI 界面
│       └── mainwindow.go
├── config.json        # 配置文件
├── go.mod
└── README.md
```

## 常见问题

### 1. 无法启动程序

**错误**: "找不到 portaudio DLL"

**解决**: 确保安装了 PortAudio 并且其 DLL 在系统 PATH 中。

### 2. 热键不工作

**原因**: 
- 热键可能被其他程序占用
- 需要管理员权限

**解决**: 
- 修改 `config.json` 中的热键
- 以管理员身份运行程序

### 3. 录不到系统声音

**原因**: 当前使用麦克风作为录音源

**解决**: 参考"录音源说明"部分的解决方案

### 4. 编译错误

**错误**: "gcc: command not found"

**解决**: 
- 确保安装了 MSYS2 和 GCC
- 将 `C:\msys64\mingw64\bin` 添加到 PATH

## 扩展建议

### 添加更多热键

在 `internal/hotkey/hotkey.go` 的 `parseHotkey()` 中添加：

```go
case "Ctrl+Shift+S":
    modifiers = MOD_CONTROL | MOD_SHIFT
    key = 'S'
```

### 支持更多音频格式

可以集成 `github.com/hajimehoshi/go-mp3` 等库来支持 MP3 导出。

### 添加设置界面

在 `internal/ui/` 中添加 `settings.go` 创建设置窗口。

## 许可证

MIT License

## 作者

Aupier 开发团队

---

**享受学习英语的乐趣！**
