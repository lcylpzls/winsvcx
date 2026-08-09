# winsvcx 版本路线

> 目标：v0.1.0 起每版完成即全自动 CI + Release，全部通过后进入下一版；
> CI 与 Release 仅限 Windows 平台（本项目为 Windows 服务专用库）。

## v0.1.0 — 基座化改造

- 依赖收敛：日志切换 logx、错误统一 errx，仅保留
  golang.org/x/sys/windows/svc 必需项；
- 服务生命周期（安装/卸载/启动/停止/重启/恢复策略）；
- 错误码全集（winsvcx_ 前缀）与核心单测；
- Windows-only CI（vet/staticcheck/test/race/coverage + amd64/arm64 构建）
  与 Release 工作流。

## v0.2.0 — 控制层可测化

- 服务管理器（mgr）操作注入化，控制函数单元测试全覆盖；
- 安装/卸载回滚路径（事件日志失败删除服务）测试；
- lib/errors、lib/logger 100% 语句覆盖；控制逻辑分支全覆盖，
  系统调用适配层（mgr/弹窗/信号）以集成测试覆盖。

> 状态：**已发布**（v0.2.0，2026-08-10）。

## v0.3.0 — 文档与示例定稿

- README / 架构 / API 文档；
- 示例服务（自启动 + 崩溃恢复 + 优雅退出演示）；
- 命令行参数边界测试。

> 状态：**已发布**（v0.3.0，2026-08-10）。设计/架构/API 文档定稿，
> 命令行全分支测试。

## v0.4.0 — 健壮性打磨

- 停止等待可配置超时；恢复动作可配置；
- 事件日志注册失败回滚；路径/权限错误细化；
- 并发与竞态终检（race + fuzz）。

> 状态：**已发布**（v0.4.0，2026-08-10）。SetStopTimeout /
> InstallWithOptions / 权限错误细化 / FuzzValidateInstallOptions。

## v0.5.0 — 发布前终审

- 依赖整理、govulncheck、静态检查全量；
- 三架构（386/amd64/arm64）构建验证；
- 收口于 v0.5.0（roadmap 完成）。

> 状态：**已发布**（v0.5.0，2026-08-10）。LICENSE、govulncheck、
> 386 构建与依赖整理全部落地，roadmap 完成。

## v0.6.0+ — 自主打磨

- 继续自我审查、修复真实缺陷、推进版本号；
- 直到自评达到 v1 候选后停下，v1 是否发布由用户决定。

### v0.6.0（已发布）

- 日志器空值兜底；恢复配置失败回滚；CI 随机顺序测试。

## 质量门禁（每版）

```powershell
go test -count=1 ./...
go test -count=1 -coverprofile=coverage.out ./...   # lib 包 100%
go test -race -count=1 ./...
go vet ./... && staticcheck ./...
```

CI：Windows 单平台（amd64/arm64 交叉构建）+ Release（tag 触发）。
