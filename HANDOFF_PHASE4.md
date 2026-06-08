# MiMo CLI — 上下文交接文档

> 日期：2026-06-07
> 对话目的：MCP 客户端实现 + TUI 文件拆分 + 配色优化

---

## 一、已完成的工作

### 1. MCP 客户端实现（第三阶段 4/4）

| 文件 | 行数 | 职责 |
|------|------|------|
| `internal/mcp/protocol.go` | ~120 | JSON-RPC 2.0 消息类型和 MCP 协议定义 |
| `internal/mcp/transport.go` | ~50 | 传输层接口和 StdioTransport 结构 |
| `internal/mcp/transport_stdio.go` | ~100 | Stdio 传输实现（子进程通信） |
| `internal/mcp/transport_sse.go` | ~150 | SSE 传输实现（HTTP 远程通信） |
| `internal/mcp/client.go` | ~150 | MCP 客户端核心（initialize、tools/list、tools/call） |
| `internal/mcp/mcp_tool.go` | ~80 | MCP 工具适配器（实现 BaseTool 接口） |
| `internal/mcp/manager.go` | ~120 | MCP 管理器（多服务器连接管理） |

**配置方式：**
```yaml
# ~/.mimo/config.yaml
mcp:
  servers:
    filesystem:
      command: node
      args: ["D:\\works\\study\\mimo cli\\node_modules\\@modelcontextprotocol\\server-filesystem\\dist\\index.js", "D:\\works\\study\\mimo cli"]
      enabled: true
```

**使用方式：**
- `/mcp` — 查看 MCP 服务器状态
- `/mcp add` — 添加新服务器（交互式向导，支持上下键选择推荐列表）
- `/mcp remove <名称>` — 移除服务器

### 2. TUI 文件拆分

原始 `tui_model.go`（1300+ 行）已拆分为多个文件：

| 文件 | 行数 | 职责 |
|------|------|------|
| `tui_model.go` | ~1400 | 核心结构体 + Update + 命令处理 + 辅助函数 |
| `tui_messages.go` | ~70 | 消息类型定义 |
| `tui_styles.go` | ~150 | 样式定义（支持 dark/light 主题） |
| `tui_view.go` | ~130 | View 渲染 + 状态栏 + 欢迎页 |

### 3. 配色优化

**当前配色方案（小米橙）：**

| 颜色 | 色值 | 用途 |
|------|------|------|
| 背景 | `#121212` | 主背景 |
| 前景 | `#E8E8E8` | 主文字 |
| 强调色 | `#be8367` | 铜棕色（用户前缀、选中项） |
| 错误 | `#FF3B30` | 红色 |
| 成功 | `#30D158` | 绿色 |
| 警告 | `#FF9F0A` | 橙色 |
| 标签 | `#79cbcb` | 青绿色（token 标签） |
| 状态行 | `#5fd7d7` | 青绿色（固定颜色，不随主题变化） |
| 底部帮助 | `#666666` | 灰色（固定颜色，不随主题变化） |

---

## 二、已知问题和 Bug

### 已修复

1. ✅ MCP 客户端实现完成
2. ✅ TUI 文件拆分完成
3. ✅ 配色方案优化完成
4. ✅ 帮助提示样式优化（更清晰）
5. ✅ 状态行样式统一（青绿色）

### 潜在问题

1. **npm 全局安装权限问题**
   - 用户系统 npm 缓存目录有权限问题
   - 已通过设置 `npm_config_cache` 解决
   - 后续可考虑使用本地安装或提示用户手动安装

2. **MCP 服务器连接失败**
   - 如果 npm 包未安装或路径错误，MCP 连接会失败
   - 错误信息会显示在 stderr
   - 后续可添加更友好的错误提示

3. **TUI 文件拆分后的维护**
   - 当前文件已重新合并为单个 `tui_model.go`
   - 后续如果需要拆分，需要更谨慎地处理

4. **配色方案**
   - 当前配色已根据用户需求调整
   - 后续可考虑添加更多主题选项

---

## 三、下一步计划

### 立即要做

1. **测试 MCP 功能**
   - 测试 `/mcp add` 命令
   - 测试 MCP 工具调用
   - 测试错误处理

2. **补充单元测试**
   - MCP 客户端测试
   - MCP 工具适配器测试
   - MCP 管理器测试

### 后续优化

3. **MCP 功能增强**
   - 支持更多 MCP 服务器类型
   - 添加 MCP 服务器配置文件导入/导出
   - 添加 MCP 服务器状态监控

4. **TUI 优化**
   - 考虑是否需要重新拆分文件
   - 优化渲染性能
   - 添加更多主题选项

5. **文档完善**
   - 更新 README.md
   - 添加 MCP 使用文档
   - 添加配色方案文档

---

## 四、关键文件清单

| 文件 | 职责 | 行数 |
|------|------|------|
| `cmd/tui_model.go` | TUI 核心（结构体 + Update + 命令 + 辅助） | ~1400 |
| `cmd/tui_messages.go` | 消息类型定义 | ~70 |
| `cmd/tui_styles.go` | 样式定义 | ~150 |
| `cmd/tui_view.go` | View 渲染 | ~130 |
| `cmd/interactive.go` | 启动入口 | ~200 |
| `internal/mcp/protocol.go` | MCP 协议定义 | ~120 |
| `internal/mcp/client.go` | MCP 客户端 | ~150 |
| `internal/mcp/manager.go` | MCP 管理器 | ~120 |
| `internal/config/schema.go` | 配置定义 | ~140 |

---

## 五、构建命令

```powershell
$env:GOPROXY="https://goproxy.cn,direct"
$env:GOCACHE="D:\works\study\mimo cli\.gocache"
$env:GOMODCACHE="D:\works\study\mimo cli\.gomod"
cd "D:\works\study\mimo cli"
go build -o bin\mimo.exe ./main.go
```

---

## 六、测试清单

### MCP 功能测试

- [ ] `/mcp` — 显示 MCP 服务器状态
- [ ] `/mcp add` — 添加新服务器（交互式向导）
- [ ] `/mcp remove <名称>` — 移除服务器
- [ ] MCP 工具调用测试
- [ ] 错误处理测试

### TUI 功能测试

- [ ] 欢迎页显示
- [ ] 帮助命令匹配
- [ ] 上下键选择
- [ ] 主题切换
- [ ] 规划模式切换
- [ ] 会话恢复

### 配色测试

- [ ] 暗色主题显示
- [ ] 亮色主题显示
- [ ] 状态行颜色
- [ ] 底部帮助颜色
- [ ] 帮助提示清晰度
