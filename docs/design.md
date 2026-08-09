# winsvcx 设计定版

> 版本：v0.3.0 · Windows 服务生命周期基础库。

## 1. 定位

winsvcx 是基于 `golang.org/x/sys/windows/svc` 的 Windows 服务基础库，
解决服务安装、卸载、启动、停止、重启与优雅退出的重复劳动。
日志与错误统一使用家族库（logx / errx）。

**边界**：

- 只面向 Windows（服务模式、服务管理器、事件日志）；
- 不替代业务逻辑，业务在主循环 `app.Run` 中实现；
- 不处理安装包/分发，只提供服务生命周期能力。

## 2. 依赖

```go
require (
	github.com/lcylpzls/errx v1.3.2 // 错误统一
	github.com/lcylpzls/logx v1.0.1 // 日志统一
	golang.org/x/sys v0.33.0        // Windows 服务必需
)
```

## 3. 数据流

```
命令行/服务管理器
      │
      ▼
main（svc.IsWindowsService 判断运行模式）
      ├─ 服务模式 ──► service.Run ──► svc.Run ──► Service.Execute
      │                                          ├─ 启动 app.Run 主循环
      │                                          ├─ 上报 StartPending/Running
      │                                          └─ 收到 Stop/Shutdown ──► 关闭 stopCh ──► 等待退出
      └─ 应用模式 ──► 系统信号 ──► 关闭 stopCh ──► app 优雅退出
```

## 4. 错误码

统一以 `winsvcx_` 为前缀，通过 errx 注册分类：

| 错误码 | 分类 |
| --- | --- |
| winsvcx_invalid_config | invalid |
| winsvcx_manager_connect | unavailable |
| winsvcx_service_not_found | not_found |
| winsvcx_service_already_exists | already_exists |
| winsvcx_service_control_failed | unavailable |
| winsvcx_service_already_running | conflict |
| winsvcx_service_already_stopped | conflict |
| winsvcx_service_stop_timeout | timeout |
| winsvcx_event_log_failed | unavailable |
| winsvcx_executable_path | unavailable |
| winsvcx_service_run_failed | unavailable |

## 5. 日志

- 入口初始化 logx：文件轮转（默认 20MB × 10 份 × 60 天，超过 1 天压缩）
  + 控制台；
- 全部日志使用简体中文。

## 6. 质量目标

- 控制逻辑分支 100% 单元测试；
- 系统调用适配层（服务管理器/消息框/系统信号）集成测试；
- CI 与 Release 仅 Windows 平台（amd64/arm64 构建）。
