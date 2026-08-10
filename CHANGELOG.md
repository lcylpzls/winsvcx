# 更新日志

## [v0.15.1] - 2026-08-10

### 变更

- 服务退出码错误统一 errx 化（CodeServiceRunFailed），对外错误带结构化 code/kind，消息保持原语义。

### 质量

- 覆盖率维持基线；race / vet / staticcheck / govulncheck 全绿。

## [v0.15.0] - 2026-08-10

### 变更

- 安装选项校验统一迁移至家族 `validx`：注册
  `winsvcx_install_options` 全局规则（启动类型、恢复动作），
  调用点走 `validx.ValidateField`；
- 默认值补齐逻辑保留在 `validateInstallOptions`，errx 错误码
  保持 winsvcx 语义，行为不变。

### 质量

- race / vet / staticcheck / govulncheck 全绿；覆盖率维持平台库基线。

## [v0.14.2] - 2026-08-10

### 变更

- go 指令与 CI/Release 工作流统一为 Go 1.26.5；
- README Go 版本徽章同步更新。

## [v0.14.1] - 2026-08-10

### 变更

- 家族统一 Go 1.21：全部 go.mod 与 CI/Release 工作流版本号对齐 1.21；
- testx 依赖升级 v1.2.1。

## [v0.14.0] - 2026-08-10

### 变更

- 家族测试底座接入：根包与 lib 子包测试改用语义等价的
  testx 断言（含 Require* 致命断言）；
- 默认构建版本号同步为 `v0.14.0`；
- 测试依赖新增 `testx v1.2.0`，errx 同步升级 v1.4.0。

### 质量

- race / vet / staticcheck 全绿。

## [v0.13.0] - 2026-08-10

### 新增

- `TraceHook` 服务生命周期链路追踪钩子（零依赖接口）：
  `RunWithHook(name, h)` 注入，服务会话自动埋点
  （winsvcx.service_name 属性），由 tracex 等外部适配器接入；
- 非零退出码映射为链路错误。

## [v0.12.0] - 2026-08-10

### 新增

- 参数解析下沉到库包 `lib/cli`：`cli.Parse` 统一检测安静模式、
  控制命令与版本参数，示例入口直接复用；
- 版本参数：`-V` / `--version` / `/V` / `-v`，输出
  `winsvcx <版本>` 并返回 0（版本号可用 ldflags 覆盖）；
- 未知开关参数返回配置错误码（不再静默忽略）。

### 质量

- lib/cli 语句覆盖率 95%；race / vet / staticcheck / shuffle 全绿。

## [v0.11.0] - 2026-08-10

### 新增

- 安静模式参数：`-quiet` / `--quiet` / `/quiet` / `-q`
  （位置任意）：关闭消息框与控制台输出，仅保留文件日志与退出码；
- 命令退出码语义：0 成功 / 1 操作失败 / 2 无效命令，
  便于脚本与静默部署；
- 服务模式检测失败改为返回非零退出码。

### 质量

- 主包语句覆盖率 97.3%；race / vet / staticcheck / shuffle /
  fuzz / 三架构构建全绿。

## [v0.10.0] - 2026-08-10

### 终轮打磨

- 系统信号处理后取消注册（signal.Stop），避免残留监听；
- 新增库方式使用示例文档（examples/README.md）；
- 最终审计：三架构构建、race、shuffle、fuzz、vet、
  staticcheck、依赖校验全绿。

> **自评：winsvcx 已达到 v1 候选标准**，v1 是否发布由用户决定。

## [v0.9.0] - 2026-08-10

### 打磨

- 服务运行失败同步写入 Windows 事件日志（事件查看器可见），
  写入失败仅告警不阻断；
- 事件日志写入注入化并补测试。

### 质量

- race / vet / staticcheck 全绿。

## [v0.8.0] - 2026-08-10

### 打磨

- 库级 `Uninstall` 自动先停止运行中的服务，避免删除失败；
- 状态查询失败时容忍并继续卸载（入口已自行处理停止）。

### 质量

- race / vet / staticcheck 全绿；主包 96.4%。

## [v0.7.0] - 2026-08-10

### 打磨

- 主入口流程可测化：`runMain` 全分支测试（可执行文件路径失败 /
  服务检测失败 / 服务模式 / 命令模式 / 应用模式）；
- 修复真实缺陷：日志目录不可用时 logx Build 失败导致日志器为
  空的问题，现降级为 stderr 输出；
- 主包语句覆盖率提升至 96.4%（仅剩入口 main 一行）。

### 质量

- race / vet / staticcheck / shuffled 全绿。

## [v0.6.0] - 2026-08-10

### 打磨

- 日志器空值兜底：库方式使用时 config.Log 未注入也能降级
  logx 全局日志器，不再 panic；
- 安装回滚增强：恢复动作配置失败同样删除已创建的服务；
- CI 增加随机顺序测试（-shuffle=on），消除用例顺序依赖。

### 质量

- lib/service 覆盖率提升至 91.1%；race / vet / staticcheck /
  shuffled 全绿。

## [v0.5.0] - 2026-08-10

### 发布前终审

- 新增 MIT LICENSE；
- CI 增加 govulncheck 作业与 windows/386 构建；
- Release 增加 windows/386 构建验证；
- 依赖整理（go mod tidy/verify）与三架构（386/amd64/arm64）
  构建验证；
- 静态检查、race、fuzz 全量复核。

> roadmap 至此完成，后续版本为自主打磨。

## [v0.4.0] - 2026-08-10

### 新增

- 停止等待超时可配置：`SetStopTimeout`（非法值返回配置错误码）；
- 安装选项可配置：`InstallWithOptions` + `DefaultInstallOptions`
  （启动类型/恢复动作/重置周期/事件日志类别，含校验）；
- 权限错误细化：访问被拒绝（需要管理员）映射
  `winsvcx_access_denied`（forbidden 分类）；
- `FuzzValidateInstallOptions` 模糊目标并接入 Windows CI。

### 质量

- lib/service 控制逻辑 90.9%（剩余为真实服务管理器适配层集成覆盖）；
- race / vet / staticcheck / fuzz 全绿。

## [v0.3.0] - 2026-08-10

### 新增

- 文档定稿：design / architecture / api 三份文档；
- 主入口命令注入化，install/uninstall/start/stop/restart 全部
  命令行分支边界测试（消息框断言）；
- README 结构、依赖与 CI 说明补全。

### 质量

- 主包命令分发覆盖；lib/errors、lib/logger 100%；
- race / vet / staticcheck 全绿。

## [v0.2.0] - 2026-08-10

### 新增

- 服务控制层可测化：服务管理器/服务句柄窄接口 + 适配层，
  安装/卸载/启动/停止/重启全分支单元测试；
- `IsServiceExist` 改为返回 `(bool, error)`，连接失败不再误报
  “服务不存在”；
- app 信号等待、服务运行入口、消息框调用注入化，补生命周期测试；
- logger 初始化默认值测试；错误码注册测试；
- 真实服务管理器适配层集成测试（无权限时自动跳过）。

### 质量

- lib/errors、lib/logger 语句覆盖率 100%；
- lib/service 控制逻辑分支全覆盖（系统调用适配层以集成测试覆盖）；
- race / vet / staticcheck 全绿。
