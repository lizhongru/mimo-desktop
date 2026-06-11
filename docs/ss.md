可以，先不要急着改代码。

我建议先写一个 **《Workspace 会话归属 Bug 排查文档》**，然后按照文档一步步验证。这样不会改着改着又引入新的问题。

# Workspace 会话归属 Bug 排查方案

## 问题现象

用户已经选择文件夹作为工作区：

```text
项目A
```

但发送第一条消息后：

```text
Session.workspace_id = default
```

而不是：

```text
Session.workspace_id = ws:项目A
```

导致：

- 左侧会话出现在「对话区」
- 工作区下没有会话
- 后续上下文隔离失效

------

# 一、预期流程

## Step1 选择工作区

```ts
WorkspacePicker
    ↓
handleWorkspaceChange()
    ↓
onSelectWorkspace()
    ↓
CreateWorkspace()
    ↓
setSelectedWorkspace(ws.id)
```

最终：

```ts
selectedWorkspace =
"ws:D:\\ProjectA"
```

------

## Step2 发送消息

```ts
handleSend()
```

读取：

```ts
useSessionStore.selectedWorkspace
```

获得：

```ts
"ws:D:\\ProjectA"
```

然后：

```ts
CreateNewSession(
    workspaceId
)
```

------

## Step3 创建 Session

```go
CreateNewSession(
    workspaceId
)
```

↓

```go
CreateSession(
    id,
    workspaceId,
    ...
)
```

↓

```sql
INSERT INTO sessions
(
    workspace_id
)
VALUES
(
    'ws:D:\\ProjectA'
)
```

------

# 二、排查重点

## 检查1

发送消息前：

```ts
console.log(
    "selectedWorkspace:",
    sessStore.selectedWorkspace
)
```

位置：

```ts
handleSend()
```

### 结果A

打印：

```text
selectedWorkspace:
""
```

说明：

工作区状态未成功写入 Store。

问题在前端。

------

### 结果B

打印：

```text
selectedWorkspace:
"ws:D:\\ProjectA"
```

说明：

前端状态正确。

继续检查后端。

------

# 检查2

Workspace 选择逻辑

当前代码：

```ts
const handleWorkspaceChange =
async (dir:string) => {

    setWorkspace(dir);

    onSelectWorkspace(dir);
}
```

问题：

```ts
onSelectWorkspace()
```

可能是异步。

用户可能在：

```text
CreateWorkspace()
```

未完成时就发送消息。

导致：

```ts
selectedWorkspace
```

还没更新。

建议：

```ts
const handleWorkspaceChange =
async (dir:string) => {

    setWorkspace(dir);

    await onSelectWorkspace(dir);
}
```

------

# 检查3

CreateNewSession

增加日志：

```go
fmt.Println(
    "[CreateNewSession]",
    workspaceID,
)
```

确认收到的 workspaceID。

------

# 检查4

CreateSession

增加日志：

```go
fmt.Println(
    "[CreateSession]",
    workspaceID,
)
```

确认最终写库前的 workspaceID。

------

# 三、重点怀疑对象

优先级排序：

## P0

WorkspacePicker 异步问题

表现：

```text
已显示项目A

但 selectedWorkspace 仍为空
```

概率最高。

------

## P1

selectedWorkspace 与本地 state 不一致

例如：

```ts
const [workspace,setWorkspace]
```

和：

```ts
useSessionStore.selectedWorkspace
```

维护两份状态。

最终：

UI显示项目A

Store实际为空

```
概率高。

---

## P2

CreateSession 中 workspaceID 被覆盖

例如：

```go
if workspaceID == "" {
    workspaceID = DefaultWorkspaceID
}
```

触发。

------

## P3

Session 创建时未使用 selectedWorkspace

例如：

```ts
CreateNewSession(DEFAULT_WS)
```

被硬编码。

------

# 四、验证步骤

1. 启动程序
2. 打开控制台
3. 选择项目A
4. 发送第一条消息
5. 查看日志：

```text
selectedWorkspace = ?
CreateNewSession workspace = ?
CreateSession workspace = ?
```

如果三处打印不一致，即可快速定位问题层级。

下一步我建议直接看你的：

- `handleSend()`
- `CreateNewSession()`
- `onSelectWorkspace()`
- `setSelectedWorkspace()`

这 4 个函数。

因为从你描述的现象来看，我已经有点怀疑：

```ts
selectedWorkspace
```

根本没有参与会话创建，或者会话创建时拿到的是旧状态（React 状态更新延迟导致的竞态）。这类问题我见过很多次，通常 5 分钟内能定位出来。把这几个函数贴出来，我帮你顺着调用链找。