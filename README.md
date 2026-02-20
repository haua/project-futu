# 浮图 - Desktop GIF Widget

一款极简的桌面减压小组件，可在所有窗口上方显示动图。

![Futu](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue)

## 特性

- 轻量级单文件可执行程序（约 8-12MB）
- 窗口置顶显示，始终在所有窗口上方
- 支持透明度调节
- 支持缩放
- 可拖拽移动位置
- 右键点击关闭，或按 ESC 退出
- 无需安装，免配置运行

### todo

1. 真正无边框桌面贴纸效果+拖拽移动
2. 使用CGO的方式点击穿透
3. 设置界面，可以设置开机自启
4. 播放器 goroutine 安全退出（防泄漏）
5. 添加应用版本信息（公司名 / 描述 / 版本号）
6. 写一个CI 跨平台自动编译脚本，dist 目录一键打包，自动发布 bat（含 git tag）
7. exe 内嵌图标（.ico）
8. 拖拽文件到浮图直接播放（这个要设置开关）
9. 右键托盘 → 最近 5 个 GIF
10. [done]文件选择框要记住上次位置
11. 现在打包不会把图标打包进去，把打包后的exe挪个位置，就找不到图标了。

## AI生成提示语

```text
用go的Fyne写一个桌面小组件，可以在电脑桌面上循环播放动图（GIF、WebP），项目名叫浮图，英文叫Futu。要求windows系统免安装即可使用。在系统托盘图标处右键可更换图片。生成一个项目，不要单文件项目。
```

## 构建

### 环境安装

1. 安装fyne：https://docs.fyne.io/started/quick/ （注意，安装后确保PATH变量中，`C:\msys64\mingw64\bin`在`~\go\bin`前面）

### Windows

```bash
# 32位
set GOARCH=386
go build -ldflags="-s -w" -o futu-386.exe ./cmd/futu

# 64位
set GOARCH=amd64
go build -ldflags="-s -w" -o futu-amd64.exe ./cmd/futu
```

### 交叉编译

从 Linux/Mac 编译 Windows 版本:

```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o futu.exe ./cmd/futu
```

## 优化体积

使用以下选项可以进一步减小程序体积:

```bash
go build -ldflags="-s -w -H=windowsgui" -o futu.exe ./cmd/futu
```

参数说明:
- `-s -w`: 去除调试信息，减小体积
- `-H=windowsgui`: 编译为 GUI 程序，不显示控制台窗口

## 项目结构

```
futu/
├── cmd/futu/          # 主程序入口
├── internal/
│   ├── window/        # Windows API 窗口管理
│   ├── gif/           # GIF 播放器
│   └── config/        # 配置管理
├── assets/            # 默认 GIF 资源
├── go.mod
└── README.md
```

## 技术栈

- **语言**: Go 1.21+
- **GUI**: Win32 API (syscall)
- **GIF 解码**: 标准库 image/gif

## 许可证

MIT License
