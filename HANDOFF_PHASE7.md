# MiMo CLI - 上下文交接文档 Phase 7

> 日期：2026-06-07
> 对话目的：Bug 修复 + 多行消息换行优化

---

## 一、本次完成的工作

### Bug 修复

| # | Bug | 修复 | 状态 |
|---|-----|------|------|
| 1 | Resize 后工具调用记录消失 | rebuildMessages() 追加 msg.toolLines 渲染 | done |
| 2 | 会话序列化丢失 thinking/toolLines | JSON序列化 + SQLite migration + loadResumeSession恢复 | done |
| 3 | 中文长文本不换行 | wrapText 仍用 len() 判断宽度，中文算错 | TODO |

### 多行输入改进

1. textarea 高度 1->3
2. inputAreaHeight 从 const 改 var，新增 calcInputAreaHeight() 动态计算
3. renderUserText 加 width 参数 + displayWidth 判断换行

### 文件修改清单

- cmd/tui_model.go: rebuildMessages追加toolLines; renderUserText加width+换行; displayWidth函数; calcInputAreaHeight动态高度; textarea高度1->3; loadResumeSession恢复thinking/toolLines; saveSessionToFile传入thinking/toolLines
- cmd/chat_message_json.go: chatMessageJSON加Thinking/ToolLines字段, MarshalJSON/UnmarshalJSON同步
- internal/session/store.go: Message struct加Thinking/ToolLines; migration加thinking/tool_lines列+ALTER TABLE兼容旧库; SaveSession/LoadSession支持新字段

---

## 二、最紧急问题：wrapText 不支持中文换行

### 根因

wrapText（tui_model.go 约1142行）两个致命问题：
1. strings.Fields() 按空格分词 - 中文无空格，整段当作一个word，永远不换行
2. len() 是字节数 - 中文字符3字节占2列，宽度完全算错

renderMarkdownSimple（约1123行）也有同样问题：if len(line) > width

displayWidth（约701行）新增的函数逻辑正确但没被 wrapText 使用。

### 修复方案

方案A（推荐）：import go-runewidth（已有间接依赖v0.0.19），逐rune按显示宽度换行
方案B（轻量）：用已有 displayWidth 函数重写 wrapText

同时需修复 renderMarkdownSimple 中 len(line) > width

---

## 三、潜在问题和风险

### 高风险
1. calcInputAreaHeight 可能导致视口抖动（textarea行数变化时viewport高度跳动），建议改固定最大高度或加debounce
2. displayWidth 对非CJK非ASCII字符处理简单（r>0x7F就算2），建议用go-runewidth替代
3. renderMarkdownSimple 换行逻辑也需同步修复

### 中风险
4. SQLite ALTER TABLE 容错 - 忽略列已存在错误（有意为之，兼容旧库）
5. toolLines 存储为JSON字符串 - ANSI颜色码兼容OK，但长内容占存储
6. textarea高度1->3后 inputAreaHeight 基准值变化，需验证不同终端高度布局

### 低风险
7. go-runewidth 是间接依赖，直接import即可，Go modules自动处理

---

## 四、下一步行动计划

### P0 立即
1. 修复 wrapText（go-runewidth或displayWidth重写，逐rune换行）
2. 同步修复 renderMarkdownSimple 中 len(line) > width
3. go build 编译验证

### P1 短期
4. calcInputAreaHeight 稳定性优化
5. 全面测试：中文换行、多行输入、resize toolLines、MCP工具、/resume恢复

### P2 后续
6. ReAct 规划引擎
7. 单元测试
8. 第四阶段生态功能

---

## 五、构建命令

`powershell
$env:GOPROXY="https://goproxy.cn,direct"
$env:GOCACHE="D:\works\study\mimo cli\.gocache"
$env:GOMODCACHE="D:\works\study\mimo cli\.gomod"
cd "D:\works\study\mimo cli"
go build -o bin\mimo.exe ./main.go
`

---

## 六、关键架构参考

消息渲染流程:
  用户输入 -> sendMessage() -> renderUserText(text,msgs,width) -> displayWidth判断 -> wrapText [BUG]
  Agent响应 -> finalizeResponse() -> streamLines存入chatMessage.toolLines -> renderMarkdownSafe [BUG]
  窗口变化 -> WindowSizeMsg -> rebuildMessages() -> renderUserText + toolLines + markdown

输入组件:
  cmdInput textinput.Model (向导/选择, 单行)
  ta textarea.Model (消息输入, 多行Ctrl+J, 高度3)
  calcInputAreaHeight(): lines+4, 范围5~12

会话存储:
  SQLite ~/.mimo/sessions.db
  messages表新增 thinking TEXT, tool_lines TEXT(JSON)
  Migration: CREATE TABLE + ALTER TABLE兼容旧库