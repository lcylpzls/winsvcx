# Go Windows 服务框架

## 项目概述

这是一个基于 Go 语言开发的 Windows 服务框架，提供了完整的 Windows 服务生命周期管理功能，
包括服务的安装、卸载、启动、停止和重启等操作。该框架使用 `golang.org/x/sys/windows/svc`
包实现 Windows 服务功能，并提供了简单易用的命令行接口进行服务管理。

> 当前状态：**v0.1.0（基座化改造）**。日志使用 logx、错误统一 errx，
> CI 与发布仅限 Windows 平台。

## 功能特性

- **服务生命周期管理**：支持服务的安装、卸载、启动、停止和重启
- **双模式运行**：可作为Windows服务运行，也可作为普通应用程序运行
- **命令行参数**：提供简单的命令行参数用于管理服务
- **用户友好提示**：使用Windows原生消息框提供操作反馈
- **日志管理**：集成 logx 家族日志库，支持文件轮转、压缩和级别控制
- **服务自动恢复**：配置服务在崩溃后自动重启
- **优雅退出**：支持服务的优雅停止和资源清理

## 安装与使用

### 编译

```bash
go build -o MySampleService.exe
```

### 服务管理命令

程序支持以下命令行参数：

- **安装服务**：`MySampleService.exe install`
- **卸载服务**：`MySampleService.exe uninstall`
- **启动服务**：`MySampleService.exe start`
- **停止服务**：`MySampleService.exe stop`
- **重启服务**：`MySampleService.exe restart`

### 运行模式

1. **Windows服务模式**：使用`MySampleService.exe start`命令（推荐方法）启动，或通过Windows服务管理器启动，或使用`net start MySampleService`命令
2. **普通应用程序模式**：直接双击或在命令行中不带参数运行可执行文件

## 代码结构

```
├── lib/                  # 库文件目录
│   ├── app/             # 应用程序核心逻辑
│   │   └── app.go       # 应用程序运行入口
│   ├── config/          # 配置相关
│   │   └── variables.go # 全局变量定义
│   ├── logger/          # 日志处理
│   │   └── logger.go    # 日志初始化和管理
│   ├── service/         # 服务相关功能
│   │   ├── control.go   # 服务控制函数(安装、卸载等)
│   │   └── service.go   # 服务实现和执行
│   └── win32/           # Windows API封装
│       └── messagebox.go # 消息框实现
└── main.go              # 主程序入口
```

### 核心模块说明

#### main.go

主程序入口，负责：
- 初始化日志系统
- 判断当前运行模式（服务模式或普通应用模式）
- 处理命令行参数
- 启动应用程序或服务

#### lib/service

服务管理核心模块，提供：
- `service.go`：实现Windows服务接口，处理服务状态变化
- `control.go`：提供服务控制功能，包括安装、卸载、启动、停止和重启

#### lib/app

应用程序核心逻辑，包含：
- 业务逻辑实现
- 信号处理
- 优雅退出机制

#### lib/logger

日志管理模块，基于logrus实现：
- 支持日志级别控制
- 支持日志文件轮转
- 支持日志压缩

#### lib/win32

Windows API封装，提供：
- 原生Windows消息框功能

## 自定义开发

### 修改服务信息

在`main.go`中修改以下变量：

```go
var serviceName string = "MySampleService"         // 服务名称
var serviceDisplayName string = "MySample测试服务"  // 服务显示名称
var serviceDescription string = "这是一个基于go语言构建的具备Windows服务能力的测试服务，具备方便的命令行参数用于管理服务。" // 服务描述
```

### 修改业务逻辑

在`lib/app/app.go`的`Run`函数中实现您的业务逻辑。

## 开发环境要求

- Go 1.23.0 或更高版本
- Windows 操作系统

## 依赖项

- `golang.org/x/sys/windows/svc`：Windows 服务支持（必需）
- `github.com/lcylpzls/logx`：家族日志库（文件轮转/控制台）
- `github.com/lcylpzls/errx`：家族错误库（统一错误码）

## CI 与发布

- CI 仅运行于 Windows：vet、staticcheck、test、race、coverage，
  并交叉构建 windows/amd64 与 windows/arm64；
- Release 由版本标签触发，仅 Windows 平台，测试全绿后发布。

## 注意事项

1. 服务安装和卸载需要管理员权限
2. 服务名称在系统中必须唯一
3. 日志文件默认保存在程序所在目录的`logs`文件夹中
4. 服务安装后会自动配置为自动启动模式
5. 服务配置了在崩溃后自动重启的恢复策略
