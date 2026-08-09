# 更新日志

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
