# winsvcx 架构详解

> 版本：v1.1.1（已实现，随版本演进）

## 1. 模块结构

```
winsvcx/
├── main.go                  # 入口：模式判断、日志初始化、命令分发
├── lib/
│   ├── app/                 # 业务主循环与优雅退出
│   ├── config/              # 全局配置（日志器）
│   ├── errors/              # 统一错误码（errx 注册）
│   ├── logger/              # logx 初始化薄封装
│   ├── service/             # 服务生命周期与控制命令
│   │   ├── service.go       # svc.Handler 实现（Execute/Run）
│   │   └── control.go       # 安装/卸载/启动/停止/重启 + mgr 适配层
│   └── win32/               # user32 消息框封装
```

## 2. 服务控制适配层

`control.go` 通过窄接口隔离 `x/sys/windows/svc/mgr`：

- `manager` / `serviceHandle` 接口定义控制所需的最小方法集；
- `managerAdapter` / `serviceAdapter` 适配真实 `*mgr.Mgr` / `*mgr.Service`；
- 测试注入 `connectManager` / `installEventLog` / `removeEventLog` /
  `executablePath` / `stopTimeout`，实现全分支单测。

## 3. 服务生命周期时序

```
Execute:
  StartPending ──► 启动 app.Run ──► Running
  Stop/Shutdown ──► StopPending ──► close(stopCh) ──► wg.Wait ──► return
```

控制命令：

- Install：存在检查 → 创建服务（自动启动）→ 崩溃恢复（5s/10s/15s）→
  事件日志；事件日志失败回滚删除服务；
- Uninstall：存在检查 → 停止（若运行）→ 删除服务 → 删除事件日志；
- Start/Stop/Restart：存在与状态校验，Stop 等待完全停止（默认 30s 超时）。

## 4. 平台说明

- 仅 Windows：`x/sys/windows/svc` 为必需依赖；
- CI 仅 windows-latest：vet / staticcheck / test / race / coverage，
  并交叉构建 windows/amd64 与 windows/arm64；
- Release 由 `v*.*.*` 标签触发，测试全绿后发布。

## 5. 安全与健壮性

- 服务管理器连接失败不会被误报为“服务不存在”；
- 安装过程事件日志失败自动回滚；
- 停止等待有超时上限；日志文件有大小/份数/天数上限。
