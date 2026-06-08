# 第三阶段已完成功能测试清单

> 日期：2026-06-06
> 构建：`go build -o bin\mimo.exe ./main.go`

---

## 一、新增工具（10 个）

### 1.1 file_diff（文件对比）
- [x] 问 mimo："对比 main.go 和 cmd/root.go 的区别"
- [x] AI 调用 `file_diff` 工具，传入两个文件路径
- [x] 输出显示逐行差异（`L行号: - 旧内容 / + 新内容`）
- [x] 对比两个相同文件时输出 "Files are identical"

### 1.2 clipboard（剪贴板）
- [x] 问 mimo："把最后一条回复复制到剪贴板"
- [x] AI 调用 `clipboard` 工具，action=write
- [x] 输出显示 "Copied N bytes to clipboard"
- [x] 问 mimo："读取剪贴板内容"
- [x] AI 调用 `clipboard` 工具，action=read
- [x] 输出显示当前剪贴板内容

### 1.3 process（进程管理）
- [x] 问 mimo："看看当前运行的进程"
- [x] AI 调用 `process` 工具，action=list
- [x] 输出显示进程列表（PID、名称、CPU、内存）
- [x] 问 mimo："有没有叫 xxx 的进程"
- [x] AI 传入 name_filter 参数，输出过滤后的进程
- [x] 尝试 kill 操作时应弹出安全确认

### 1.4 env（环境变量）
- [x] 问 mimo："GOPATH 环境变量是什么"
- [x] AI 调用 `env` 工具，action=get, name=GOPATH
- [x] 输出显示 `GOPATH=...`
- [x] 问 mimo："列出所有 GOPROXY 相关的环境变量"
- [x] AI 传入 filter=GOPROXY，输出过滤后的变量

### 1.5 dependency（包管理器）
- [x] 在 Go 项目中问 mimo："看看当前项目的依赖"
- [x] AI 调用 `dependency` 工具，action=list
- [x] 输出显示 `go list -m all` 的结果
- [x] 问 mimo："安装一个新的依赖"（应弹出确认）
- [x] AI 调用 action=add，自动检测 go.mod

### 1.6 http_request（HTTP 请求）
- [x] 问 mimo："访问 https://httpbin.org/get 看看"
- [x] AI 调用 `http_request` 工具，method=GET
- [x] 输出显示 HTTP 状态码、响应头、响应体
- [x] 问 mimo："POST 一个 JSON 到 httpbin.org/post"
- [x] AI 传入 method=POST, body=JSON, Content-Type 自动设置

### 1.7 web_search（网络搜索）
- [x] 问 mimo："搜索一下 Go 语言最新版本"
- [x] AI 调用 `web_search` 工具
- [x] 输出显示搜索结果列表（标题 + 链接 + 摘要）
- [x] 结果数量默认 5 条

### 1.8 file_delete（文件删除）
- [ ] 问 mimo："删除 temp.txt 文件"
- [ ] 弹出安全确认（RequiresConfirmation=true）
- [ ] 确认后文件被删除，输出 "Deleted file: temp.txt"
- [ ] 尝试删除不存在的文件时输出错误信息
- [ ] 尝试删除 .git / .mimo / node_modules 时被拒绝

### 1.9 dir_create（目录创建）
- [ ] 问 mimo："创建一个 test_dir 目录"
- [ ] AI 调用 `dir_create` 工具
- [ ] 输出 "Created directory: test_dir"
- [ ] 目录已存在时输出 "Directory already exists"

### 1.10 json_query（JSON/YAML 查询）
- [x] 问 mimo："看看 go.mod 里的 module 名称"
- [x] AI 调用 `json_query` 工具（或 file_read）
- [x] 问 mimo："读取 package.json 的 dependencies.react 版本"
- [x] AI 使用 dot-notation 路径查询，输出具体值
- [x] 查询不存在的路径时输出 "Path not found"
- [x] YAML 文件也能正确解析

---

## 二、SQLite 会话持久化

### 2.1 数据库初始化
- [x] 首次启动 mimo 后，`~/.mimo/sessions.db` 文件被创建
- [x] 文件大小不为 0

### 2.2 会话保存
- [x] 正常对话几轮后退出 mimo
- [x] 用 SQLite 工具查看 `~/.mimo/sessions.db`：
  - `sessions` 表有记录
  - `messages` 表有对应的消息
  - `last_message` 字段是最后一条用户消息
  - `updated_at` 是最新时间

### 2.3 会话恢复（/resume）
- [x] 输入 `/resume`，显示历史会话列表
- [ ] 列表显示每条会话的最后消息 + 时间（如 `帮我写个函数  (06-06 15:30)`）
- [x] 上下键选择，Enter 确认
- [x] 恢复后对话历史完整显示
- [x] 恢复后继续对话，新消息正确追加
- [ ] 退出后会话更新到同一条记录（不创建新记录）

### 2.4 Markdown 导出兼容
- [x] 退出 mimo 后，`~/.mimo/sessions/session_*.md` 文件仍然生成
- [x] Markdown 文件内容与之前格式一致

### 2.5 边界情况
- [x] 没有任何历史会话时 `/resume` 显示 "没有可恢复的会话"
- [x] sessionStore 为 nil 时（如数据库损坏），不崩溃，显示错误提示

---

## 三、工具注册与系统提示词

### 3.1 工具注册
- [ ] 启动 mimo 后，AI 能识别并调用所有新工具
- [ ] 问 mimo "你有哪些工具" — 回答应包含全部 25 个工具

### 3.2 系统提示词
- [ ] 系统提示词中 Available Tools 列表包含所有新工具
- [ ] AI 根据用户意图自动选择正确的工具

---

## 四、快速验证脚本

```powershell
# 编译
$env:GOPROXY="https://goproxy.cn,direct"
$env:GOCACHE="D:\works\study\mimo cli\.gocache"
$env:GOMODCACHE="D:\works\study\mimo cli\.gomod"
cd "D:\works\study\mimo cli"
go build -o bin\mimo.exe ./main.go

# 测试工具（非交互模式）
.\bin\mimo.exe run "列出当前目录的环境变量中包含 PATH 的" --verbose
.\bin\mimo.exe run "查看当前运行的进程" --verbose

# 验证 SQLite
# 启动 mimo，对话几轮，退出
# 然后检查：
Test-Path "$env:USERPROFILE\.mimo\sessions.db"
```
