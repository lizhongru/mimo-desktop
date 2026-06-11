# 前端 Modal 与浮层组件规范

## 现状盘点

当前前端里有两类相似但不应混用的公共弹层模式。

第一类是全屏 Modal，使用 `modal-overlay`、`modal-dialog` 和 `is-open` 触发淡入与缩放动画。当前使用者包括：

- `desktop/frontend/src/components/common/AboutPanel.tsx`
- `desktop/frontend/src/components/common/ShortcutsPanel.tsx`
- `desktop/frontend/src/components/common/HelpLogPanel.tsx`
- `desktop/frontend/src/components/common/ToolsViewer.tsx`
- `desktop/frontend/src/components/settings/SettingsPage.tsx`
- `desktop/frontend/src/components/confirm/ConfirmDialog.tsx`
- `desktop/frontend/src/components/common/MemoryPanelModal.tsx`
- `desktop/frontend/src/components/common/CheckpointPanelModal.tsx`
- `desktop/frontend/src/components/common/TaskPanelModal.tsx`
- `desktop/frontend/src/components/common/ActorPanelModal.tsx`

第二类是局部 Popover 或 Dropdown，使用 `useAnimatedOpen` 控制卸载时机，并使用 `animate-pop-up`、`animate-pop-out` 触发进出动画。当前使用者包括：

- `desktop/frontend/src/components/chat/ChatInput.tsx`
- `desktop/frontend/src/components/chat/ModelReasoningPicker.tsx`
- `desktop/frontend/src/components/welcome/WelcomeView.tsx`
- `desktop/frontend/src/components/agent/AgentSwitcher.tsx`
- `desktop/frontend/src/components/layout/LeftSidebar.tsx` 里的用户菜单和语言子菜单

## 选择规则

需要遮罩整个应用、点击遮罩关闭、居中显示内容时，使用全屏 Modal 模式。

需要锚定在某个按钮、输入框、菜单项附近，并且需要退出动画后再卸载 DOM 时，使用 Popover 或 Dropdown 模式。

不要用 `setTimeout` 人为错开 Modal 的打开时机。Modal 的动画应只由 `open` 状态切换到 `is-open` 类名来触发。

不要在被 `overflow-hidden` 裁剪的布局容器里打开全屏 Modal。如果入口在 Sidebar 或菜单里，Modal 的 `open` 状态应提升到不会被裁剪的稳定父级，例如 `AppLayout`。

不要为同一个 Modal 同时保留两套 `open` 状态和两处渲染。入口组件只负责调用 `onOpenXxx`，Modal 实例只保留在统一的宿主组件里。

## 全屏 Modal 写法

推荐结构如下：

```tsx
interface Props {
  open: boolean;
  onClose: () => void;
}

export function ExamplePanel({ open, onClose }: Props) {
  return (
    <div className={`modal-overlay ${open ? "is-open" : ""}`} onClick={onClose}>
      <div className="modal-dialog w-[500px] mx-4" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between px-5 py-4 border-b border-bdr-div">
          <h3 className="text-sm font-medium text-txt">标题</h3>
          <button onClick={onClose} className="p-1 rounded hover:bg-elevated text-txt-g cursor-pointer">
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="overflow-y-auto max-h-[60vh]">
          {/* 内容组件 */}
        </div>
      </div>
    </div>
  );
}
```

全屏 Modal 的内容组件应尽量保持纯内容职责，例如 `MemorySearch`、`CheckpointPanel`、`TaskPanel`、`ActorPanel`。Modal 文件只负责标题、关闭按钮、宽度、滚动容器和遮罩行为。

如果入口在 `LeftSidebar` 这类嵌套组件里，应采用下面的数据流：

```tsx
// AppLayout.tsx
const [exampleOpen, setExampleOpen] = useState(false);

<LeftSidebar onOpenExample={() => setExampleOpen(true)} />
<ExamplePanel open={exampleOpen} onClose={() => setExampleOpen(false)} />

// LeftSidebar.tsx
<button onClick={() => { close(); onOpenExample(); }}>
  打开面板
</button>
```

这种写法可以避免 Sidebar 的 `overflow-hidden` 裁剪，也可以避免本地 Modal 和全局 Modal 同时存在。

## Popover 与 Dropdown 写法

局部浮层应使用 `useAnimatedOpen`，因为这类组件通常需要在关闭动画结束后再卸载 DOM。

```tsx
const [rawOpen, setRawOpen] = useState(false);
const { shouldRender, closing } = useAnimatedOpen(rawOpen, 120);

return (
  <div className="relative">
    <button onClick={() => setRawOpen(true)}>打开</button>
    {shouldRender && (
      <>
        <div className="fixed inset-0 z-40" onClick={() => setRawOpen(false)} />
        <div className={closing ? "animate-pop-out" : "animate-pop-up"}>
          {/* 浮层内容 */}
        </div>
      </>
    )}
  </div>
);
```

Popover 或 Dropdown 不应使用 `modal-overlay`。全屏 Modal 也不应使用 `useAnimatedOpen`，除非未来统一改造关闭动画和 DOM 生命周期。

## 何时抽公共 React 组件

目前项目已经通过 `globals.css` 统一了全屏 Modal 动画样式，但还没有统一 React 壳组件。

如果后续继续新增两个以上相似 Modal，建议新增 `desktop/frontend/src/components/common/ModalShell.tsx`，封装以下职责：

- 遮罩和 `is-open` 类名
- 点击遮罩关闭
- `modal-dialog` 基础类名
- 标题栏、图标、关闭按钮
- 默认宽度、滚动区域和可选 `bodyClassName`

建议接口如下：

```tsx
interface ModalShellProps {
  open: boolean;
  onClose: () => void;
  title: React.ReactNode;
  icon?: React.ReactNode;
  widthClassName?: string;
  bodyClassName?: string;
  dialogClassName?: string;
  children: React.ReactNode;
}
```

抽取前应优先迁移最简单且结构一致的面板，例如 `AboutPanel`、`ShortcutsPanel`、`HelpLogPanel`、`MemoryPanelModal`、`CheckpointPanelModal`、`TaskPanelModal`、`ActorPanelModal`。`ToolsViewer`、`SettingsPage` 和 `ConfirmDialog` 有更强的自定义布局，适合最后迁移或保留独立实现。

## 新增同类组件检查清单

- 全屏遮罩使用 `modal-overlay`，内容容器使用 `modal-dialog`。
- 打开动画只依赖 `open` 到 `is-open` 的类名变化。
- 入口在嵌套组件时，只调用父级传入的 `onOpenXxx`。
- Modal 只在一个稳定父级中渲染一次。
- 不使用 `setTimeout` 作为打开动画的触发手段。
- 点击遮罩关闭，点击 `modal-dialog` 内部使用 `stopPropagation`。
- 面板业务内容放在领域目录，Modal 壳放在 `common` 或页面级宿主中。
- 局部 Popover 使用 `useAnimatedOpen`，不要和全屏 Modal 模式混用。
