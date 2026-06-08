# 第二阶段功能测试清单

> 日期：2026-06-06
> 构建命令：`go build -o bin\mimo.exe ./main.go`

---

## 1. mimo run 增强

### 1.1 基础运行
- [ ] `mimo run "hello"` 正常输出回答

### 1.2 管道输入
- [ ] `echo "what is 1+1" | mimo run` 能从 stdin 读取并回答
- [ ] `type README.md | mimo run "summarize this"` 管道 + 参数组合
- [ ] `mimo run`（无参数无管道）提示错误信息而非卡住

### 1.3 JSON 输出
- [ ] `mimo run "hello" --output json` 输出合法 JSON
- [ ] JSON 包含 `task`、`response`、`duration`、`duration_ms`、`tokens`、`exit_code` 字段
- [ ] JSON 模式下终端不输出流式文本（静默）

### 1.4 Markdown 输出
- [ ] `mimo run "hello" --output markdown` 输出带标题和元信息的 Markdown

### 1.5 退出码
- [ ] 正常完成：`echo %ERRORLEVEL%` 返回 `0`
- [ ] 故意触发错误（如不传参数）：返回 `1`
- [ ] 响应包含"需要人工审核"时：返回 `2`

---

## 2. AGENT.md 解析

### 2.1 结构化提取
- [ ] 在有 `AGENT.md` 的项目中启动 mimo，系统提示词中包含解析后的 sections
- [ ] "代理规则"section 以 `⚠ MUST-FOLLOW Agent Rules` 标题出现
- [ ] "禁止操作"section 以 `🚫 FORBIDDEN Operations` 标题出现
- [ ] "编码规范""常用命令""技术栈"等 section 独立显示

### 2.2 Git 信息
- [ ] 启动后系统提示词中包含 `Git Branch:` 行
- [ ] 有未提交修改时显示 `Git Status:` 及文件列表
- [ ] 工作区干净时显示 `Git Status: clean`

### 2.3 额外规则
- [ ] 创建 `.mimo/rules.md` 文件后重启，规则出现在系统提示词中

### 2.4 无 AGENT.md 降级
- [ ] 在没有 AGENT.md 的目录启动，不报错，正常运行

---

## 3. 多行输入

### 3.1 模式切换
- [ ] 单行模式下底部提示显示 `Shift+Enter: 多行模式`
- [ ] 按 Shift+Enter 后切换到多行模式
- [ ] 多行模式下底部提示显示 `Shift+Enter: 切换单行  Ctrl+Enter: 发送  Enter: 换行`
- [ ] 再按 Shift+Enter 切回单行模式

### 3.2 多行输入
- [ ] 多行模式下按 Enter 输入框内换行（不发送）
- [ ] 输入多行文本后按 Ctrl+Enter 发送
- [ ] 发送后自动切回单行模式

### 3.3 内容保持
- [ ] 单行模式输入文字后切到多行，文字保留
- [ ] 多行模式输入文字后切到单行，文字保留

### 3.4 忙碌时
- [ ] 多行模式下 Agent 处理中，输入框行为正常

---

## 4. Git 工具格式化

### 4.1 git_status
- [ ] 工作区干净时显示 `✓ Working tree clean`
- [ ] 有修改时分三组显示：`Staged (N)` / `Unstaged (N)` / `Untracked (N)`
- [ ] 每组显示对应文件列表

### 4.2 git_diff
- [ ] 无修改时显示 `No changes`（或 `No staged changes`）
- [ ] 有修改时先显示 diffstat 摘要，再显示 diff 内容
- [ ] 超长 diff 截断为 4000 字符 + `... (truncated)`

### 4.3 git_log
- [ ] 输出格式为 `hash | date | author | message`
- [ ] `count` 参数生效（默认 10 条）

### 4.4 git_commit
- [ ] 提交成功后显示 `Committed:` + 提交摘要（hash | date | message）

---

## 5. 主题切换

### 5.1 /theme 命令
- [ ] 输入 `/theme` 显示当前主题和可选项
- [ ] `/theme dark` 切换到暗色主题
- [ ] `/theme light` 切换到亮色主题
- [ ] `/theme xxx`（无效值）显示错误提示

### 5.2 主题持久化
- [ ] 切换主题后重启 mimo，主题保持
- [ ] `~/.mimo/config.yaml` 中 `theme` 字段已更新

### 5.3 视觉效果
- [ ] dark 主题：深色背景、暖铜色品牌色
- [ ] light 主题：浅色背景、文字对比度足够
- [ ] 切换后所有 UI 元素（状态栏、输入框、消息、分隔线、横幅）颜色同步变化

---

## 快速测试脚本

```powershell
# 编译
$env:GOPROXY="https://goproxy.cn,direct"
$env:GOCACHE="D:\works\study\mimo cli\.gocache"
$env:GOMODCACHE="D:\works\study\mimo cli\.gomod"
cd "D:\works\study\mimo cli"
go build -o bin\mimo.exe ./main.go

# 测试管道输入 + JSON 输出
echo "say hello" | .\bin\mimo.exe run --output json

# 测试退出码
.\bin\mimo.exe run "say hello"
echo $LASTEXITCODE

# 测试 Markdown 输出
.\bin\mimo.exe run "say hello" --output markdown
```
