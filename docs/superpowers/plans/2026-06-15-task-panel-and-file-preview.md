# Task Panel Enhancements + File Preview Enhancement

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add rename/archive/progress operations to TaskPanel, and enhance file preview with Markdown rendering, JSON folding, and image display.

**Architecture:** 
- TaskPanel gains inline rename editing, archive button for done/abandoned/blocked tasks, and a progress input for in_progress tasks. All use existing backend APIs (`TaskRename`, `TaskArchive`, `TaskProgress`).
- File preview in RightSidebar gains Markdown rendering via `react-markdown`, JSON collapsible tree, and base64 image display for image files. Backend returns base64 content for image files.

**Tech Stack:** React, TypeScript, Tailwind CSS, react-markdown, lucide-react

---

## File Structure

### Task Panel Enhancement
- Modify: `desktop/frontend/src/components/task/TaskPanel.tsx` — Add rename/archive/progress UI
- Modify: `desktop/frontend/src/lib/i18n.ts` — Add translation keys for new buttons

### File Preview Enhancement
- Modify: `desktop/app_files.go` — Return base64 for image files in `ReadFilePreview`
- Modify: `desktop/frontend/src/components/layout/RightSidebar.tsx` — Add Markdown renderer, JSON tree, image preview
- Create: `desktop/frontend/src/components/file/MarkdownPreview.tsx` — Markdown rendering component
- Create: `desktop/frontend/src/components/file/JsonTree.tsx` — Collapsible JSON tree component
- Modify: `desktop/frontend/package.json` — Add `react-markdown` dependency

---

### Task 1: Task Panel — Add Rename, Archive, Progress

**Files:**
- Modify: `desktop/frontend/src/components/task/TaskPanel.tsx`
- Modify: `desktop/frontend/src/lib/i18n.ts`

- [ ] **Step 1: Add i18n keys for new task operations**

In `desktop/frontend/src/lib/i18n.ts`, add to the `TranslationKey` union type:
```
| "task_rename"
| "task_archive"
| "task_progress"
| "task_progress_placeholder"
| "task_rename_placeholder"
| "task_unblock"
```

Add to both `zh` and `en` translation objects:
```typescript
// zh
task_rename: "重命名",
task_archive: "归档",
task_progress: "进度",
task_progress_placeholder: "输入进度说明...",
task_rename_placeholder: "输入新名称...",
task_unblock: "解除阻塞",

// en
task_rename: "Rename",
task_archive: "Archive",
task_progress: "Progress",
task_progress_placeholder: "Enter progress note...",
task_rename_placeholder: "Enter new name...",
task_unblock: "Unblock",
```

- [ ] **Step 2: Add rename state and handler to TaskPanel**

In `desktop/frontend/src/components/task/TaskPanel.tsx`, add state:
```typescript
const [renamingTaskId, setRenamingTaskId] = useState<string | null>(null);
const [renameValue, setRenameValue] = useState("");
const [progressTaskId, setProgressTaskId] = useState<string | null>(null);
const [progressValue, setProgressValue] = useState("");
```

Add handlers:
```typescript
const handleRenameTask = useCallback(
  async (id: string) => {
    if (!renameValue.trim()) return;
    setIsLoading(true);
    try {
      const result = await window.go?.desktop?.App?.TaskRename?.(id, renameValue);
      setLastResult(result);
      if (result?.success) {
        setRenamingTaskId(null);
        setRenameValue("");
        loadTasks();
      }
    } catch (error) {
      console.error("Failed to rename task:", error);
    } finally {
      setIsLoading(false);
    }
  },
  [renameValue, loadTasks]
);

const handleArchiveTask = useCallback(
  async (id: string) => {
    setIsLoading(true);
    try {
      const result = await window.go?.desktop?.App?.TaskArchive?.(id);
      setLastResult(result);
      if (result?.success) loadTasks();
    } catch (error) {
      console.error("Failed to archive task:", error);
    } finally {
      setIsLoading(false);
    }
  },
  [loadTasks]
);

const handleProgressTask = useCallback(
  async (id: string) => {
    if (!progressValue.trim()) return;
    setIsLoading(true);
    try {
      const result = await window.go?.desktop?.App?.TaskProgress?.(id, progressValue);
      setLastResult(result);
      if (result?.success) {
        setProgressTaskId(null);
        setProgressValue("");
      }
    } catch (error) {
      console.error("Failed to add progress:", error);
    } finally {
      setIsLoading(false);
    }
  },
  [progressValue]
);
```

- [ ] **Step 3: Add UI buttons and inline editors**

In the task actions `<div className="flex gap-1 ml-6">` section, add after existing buttons:

```tsx
{/* Rename button — available for all non-terminal tasks */}
{task.status !== "done" && task.status !== "abandoned" && task.status !== "archived" && (
  <button
    onClick={() => {
      setRenamingTaskId(task.id);
      setRenameValue(task.summary);
      setProgressTaskId(null);
    }}
    disabled={isLoading}
    className="px-2 py-0.5 text-xs bg-purple-500/20 text-purple-400 rounded hover:bg-purple-500/30"
  >
    {t("task_rename")}
  </button>
)}

{/* Progress button — only for in_progress tasks */}
{task.status === "in_progress" && (
  <button
    onClick={() => {
      setProgressTaskId(task.id);
      setProgressValue("");
      setRenamingTaskId(null);
    }}
    disabled={isLoading}
    className="px-2 py-0.5 text-xs bg-cyan-500/20 text-cyan-400 rounded hover:bg-cyan-500/30"
  >
    {t("task_progress")}
  </button>
)}

{/* Archive button — only for done/abandoned/blocked tasks */}
{(task.status === "done" || task.status === "abandoned" || task.status === "blocked") && (
  <button
    onClick={() => handleArchiveTask(task.id)}
    disabled={isLoading}
    className="px-2 py-0.5 text-xs bg-gray-500/20 text-gray-400 rounded hover:bg-gray-500/30"
  >
    {t("task_archive")}
  </button>
)}
```

Add inline rename editor below the actions div:
```tsx
{/* Inline rename editor */}
{renamingTaskId === task.id && (
  <div className="flex gap-2 ml-6 mt-2">
    <input
      value={renameValue}
      onChange={(e) => setRenameValue(e.target.value)}
      onKeyDown={(e) => e.key === "Enter" && handleRenameTask(task.id)}
      placeholder={t("task_rename_placeholder")}
      className="flex-1 px-2 py-1 text-xs bg-surface border border-bdr rounded text-txt focus:outline-none focus:border-accent"
      autoFocus
    />
    <button
      onClick={() => handleRenameTask(task.id)}
      disabled={isLoading || !renameValue.trim()}
      className="px-2 py-1 text-xs bg-accent/20 text-accent rounded hover:bg-accent/30"
    >
      {t("save")}
    </button>
    <button
      onClick={() => setRenamingTaskId(null)}
      className="px-2 py-1 text-xs text-txt-m hover:text-txt"
    >
      {t("cancel")}
    </button>
  </div>
)}

{/* Inline progress editor */}
{progressTaskId === task.id && (
  <div className="flex gap-2 ml-6 mt-2">
    <input
      value={progressValue}
      onChange={(e) => setProgressValue(e.target.value)}
      onKeyDown={(e) => e.key === "Enter" && handleProgressTask(task.id)}
      placeholder={t("task_progress_placeholder")}
      className="flex-1 px-2 py-1 text-xs bg-surface border border-bdr rounded text-txt focus:outline-none focus:border-accent"
      autoFocus
    />
    <button
      onClick={() => handleProgressTask(task.id)}
      disabled={isLoading || !progressValue.trim()}
      className="px-2 py-1 text-xs bg-accent/20 text-accent rounded hover:bg-accent/30"
    >
      {t("save")}
    </button>
    <button
      onClick={() => setProgressTaskId(null)}
      className="px-2 py-1 text-xs text-txt-m hover:text-txt"
    >
      {t("cancel")}
    </button>
  </div>
)}
```

- [ ] **Step 4: Add archived status to STATUS_COLORS and STATUS_ICONS**

```typescript
const STATUS_COLORS: Record<string, string> = {
  open: "text-txt-m",
  in_progress: "text-blue-400",
  blocked: "text-yellow-400",
  done: "text-green-400",
  abandoned: "text-red-400",
  archived: "text-gray-500",  // NEW
};

const STATUS_ICONS: Record<string, typeof ListTodo> = {
  open: ListTodo,
  in_progress: Play,
  blocked: AlertCircle,
  done: CheckCircle2,
  abandoned: XCircle,
  archived: Archive,  // NEW — import Archive from lucide-react
};
```

Add `Archive` to the lucide-react import.

- [ ] **Step 5: Import t function**

Add at top:
```typescript
import { t } from "../../lib/i18n";
```

- [ ] **Step 6: Build and verify**

Run: `cd desktop/frontend; npm run build`
Expected: Build succeeds

---

### Task 2: File Preview — Backend Image Support

**Files:**
- Modify: `desktop/app_files.go`

- [ ] **Step 1: Add image extension detection and base64 encoding**

In `desktop/app_files.go`, add helper function:
```go
func isImageExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".ico", ".svg":
		return true
	}
	return false
}

func mimeByExtension(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".ico":
		return "image/x-icon"
	case ".svg":
		return "image/svg+xml"
	}
	return "application/octet-stream"
}
```

- [ ] **Step 2: Add Mime field to FilePreview struct**

```go
type FilePreview struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	IsDir     bool   `json:"isDir"`
	SizeBytes int64  `json:"sizeBytes"`
	IsText    bool   `json:"isText"`
	IsImage   bool   `json:"isImage"`
	Truncated bool   `json:"truncated"`
	Content   string `json:"content"`
	Language  string `json:"language"`
	Mime      string `json:"mime"`
}
```

- [ ] **Step 3: Handle image files in ReadFilePreview**

In `ReadFilePreview`, after the binary check, add image handling:
```go
if isImageExtension(path) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}
	preview.IsImage = true
	preview.IsText = false
	preview.Content = base64.StdEncoding.EncodeToString(raw)
	preview.Mime = mimeByExtension(path)
	preview.Language = "image"
	return preview, nil
}
```

Add `"encoding/base64"` to imports.

- [ ] **Step 4: Build and verify**

Run: `go build ./desktop/...`
Expected: Build succeeds

---

### Task 3: File Preview — Frontend Markdown Rendering

**Files:**
- Create: `desktop/frontend/src/components/file/MarkdownPreview.tsx`
- Modify: `desktop/frontend/src/components/layout/RightSidebar.tsx`
- Modify: `desktop/frontend/package.json`

- [ ] **Step 1: Install react-markdown**

Run: `cd desktop/frontend; npm install react-markdown`
Expected: `react-markdown` added to dependencies

- [ ] **Step 2: Create MarkdownPreview component**

Create `desktop/frontend/src/components/file/MarkdownPreview.tsx`:
```tsx
import ReactMarkdown from "react-markdown";

interface Props {
  content: string;
}

export function MarkdownPreview({ content }: Props) {
  return (
    <div className="prose prose-sm prose-invert max-w-none text-txt-2 
      [&_h1]:text-lg [&_h1]:font-bold [&_h1]:text-txt [&_h1]:mt-4 [&_h1]:mb-2
      [&_h2]:text-base [&_h2]:font-semibold [&_h2]:text-txt [&_h2]:mt-3 [&_h2]:mb-1.5
      [&_h3]:text-sm [&_h3]:font-medium [&_h3]:text-txt [&_h3]:mt-2 [&_h3]:mb-1
      [&_p]:text-xs [&_p]:leading-relaxed [&_p]:mb-2
      [&_code]:text-[11px] [&_code]:bg-elevated [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:font-mono
      [&_pre]:bg-elevated [&_pre]:p-3 [&_pre]:rounded [&_pre]:overflow-x-auto [&_pre]:mb-2
      [&_pre_code]:bg-transparent [&_pre_code]:p-0
      [&_ul]:list-disc [&_ul]:pl-4 [&_ul]:text-xs [&_ul]:mb-2
      [&_ol]:list-decimal [&_ol]:pl-4 [&_ol]:text-xs [&_ol]:mb-2
      [&_li]:mb-0.5
      [&_a]:text-accent [&_a]:underline [&_a]:underline-offset-2
      [&_blockquote]:border-l-2 [&_blockquote]:border-accent [&_blockquote]:pl-3 [&_blockquote]:italic [&_blockquote]:text-txt-m
      [&_table]:text-xs [&_table]:border-collapse
      [&_th]:border [&_th]:border-bdr [&_th]:px-2 [&_th]:py-1 [&_th]:bg-elevated [&_th]:text-left
      [&_td]:border [&_td]:border-bdr [&_td]:px-2 [&_td]:py-1
      [&_img]:max-w-full [&_img]:rounded
      [&_hr]:border-bdr [&_hr]:my-3
    ">
      <ReactMarkdown>{content}</ReactMarkdown>
    </div>
  );
}
```

- [ ] **Step 3: Integrate MarkdownPreview in RightSidebar**

In `RightSidebar.tsx`, add import:
```typescript
import { MarkdownPreview } from "../file/MarkdownPreview";
```

In the `PreviewView` component, add Markdown rendering before the generic text block:
```tsx
{preview.isText && preview.content && preview.language === "markdown" && (
  <MarkdownPreview content={preview.content} />
)}

{preview.isText && preview.content && preview.language !== "markdown" && (
  <CodeBlock content={preview.content} language={preview.language} />
)}
```

- [ ] **Step 4: Build and verify**

Run: `cd desktop/frontend; npm run build`
Expected: Build succeeds

---

### Task 4: File Preview — Frontend Image Display

**Files:**
- Modify: `desktop/frontend/src/components/layout/RightSidebar.tsx`

- [ ] **Step 1: Update FilePreviewState interface**

```typescript
interface FilePreviewState {
  name: string;
  path: string;
  isDir: boolean;
  sizeBytes: number;
  isText: boolean;
  isImage: boolean;
  truncated: boolean;
  content: string;
  language: string;
  mime: string;
}
```

- [ ] **Step 2: Add image preview section**

In `PreviewView`, add image rendering:
```tsx
{preview.isImage && (
  <div className="flex items-center justify-center p-4 min-h-[200px]">
    <img
      src={`data:${preview.mime};base64,${preview.content}`}
      alt={preview.name}
      className="max-w-full max-h-[60vh] object-contain rounded border border-bdr"
      onError={(e) => {
        (e.target as HTMLImageElement).style.display = "none";
      }}
    />
  </div>
)}
```

- [ ] **Step 3: Update metadata section for images**

```tsx
<span>{preview.isDir ? "Directory" : preview.isImage ? "Image" : preview.isText ? "Text" : "Binary"}</span>
```

- [ ] **Step 4: Build and verify**

Run: `cd desktop/frontend; npm run build`
Expected: Build succeeds

---

### Task 5: File Preview — JSON Collapsible Tree

**Files:**
- Create: `desktop/frontend/src/components/file/JsonTree.tsx`
- Modify: `desktop/frontend/src/components/layout/RightSidebar.tsx`

- [ ] **Step 1: Create JsonTree component**

Create `desktop/frontend/src/components/file/JsonTree.tsx`:
```tsx
import { useState } from "react";
import { ChevronRight, ChevronDown } from "lucide-react";

interface Props {
  data: unknown;
  level?: number;
}

export function JsonTree({ data, level = 0 }: Props) {
  if (data === null) {
    return <span className="text-txt-m italic">null</span>;
  }
  if (data === undefined) {
    return <span className="text-txt-m italic">undefined</span>;
  }
  if (typeof data === "string") {
    return <span className="text-green-400">"{data}"</span>;
  }
  if (typeof data === "number") {
    return <span className="text-amber-400">{data}</span>;
  }
  if (typeof data === "boolean") {
    return <span className="text-purple-400">{String(data)}</span>;
  }
  if (Array.isArray(data)) {
    return <JsonArray data={data} level={level} />;
  }
  return <JsonObject data={data as Record<string, unknown>} level={level} />;
}

function JsonObject({ data, level }: { data: Record<string, unknown>; level: number }) {
  const [expanded, setExpanded] = useState(level < 2);
  const keys = Object.keys(data);

  if (keys.length === 0) {
    return <span className="text-txt-m">{"{}"}</span>;
  }

  return (
    <div className="font-mono text-[11px]">
      <button
        onClick={() => setExpanded(!expanded)}
        className="inline-flex items-center gap-0.5 hover:bg-elevated rounded px-0.5 cursor-pointer"
      >
        {expanded ? (
          <ChevronDown className="w-3 h-3 text-txt-m" />
        ) : (
          <ChevronRight className="w-3 h-3 text-txt-m" />
        )}
        <span className="text-txt-m">{"{"}</span>
        {!expanded && <span className="text-txt-m">...{keys.length} keys{"}"}</span>}
      </button>
      {expanded && (
        <div className="ml-4 border-l border-bdr-sub pl-2">
          {keys.map((key) => (
            <div key={key} className="py-0.5">
              <span className="text-blue-400">"{key}"</span>
              <span className="text-txt-m">: </span>
              <JsonTree data={data[key]} level={level + 1} />
            </div>
          ))}
        </div>
      )}
      {expanded && <span className="text-txt-m">{"}"}</span>}
    </div>
  );
}

function JsonArray({ data, level }: { data: unknown[]; level: number }) {
  const [expanded, setExpanded] = useState(level < 2);

  if (data.length === 0) {
    return <span className="text-txt-m">[]</span>;
  }

  return (
    <div className="font-mono text-[11px]">
      <button
        onClick={() => setExpanded(!expanded)}
        className="inline-flex items-center gap-0.5 hover:bg-elevated rounded px-0.5 cursor-pointer"
      >
        {expanded ? (
          <ChevronDown className="w-3 h-3 text-txt-m" />
        ) : (
          <ChevronRight className="w-3 h-3 text-txt-m" />
        )}
        <span className="text-txt-m">[</span>
        {!expanded && <span className="text-txt-m">...{data.length} items]</span>}
      </button>
      {expanded && (
        <div className="ml-4 border-l border-bdr-sub pl-2">
          {data.map((item, idx) => (
            <div key={idx} className="py-0.5">
              <span className="text-txt-m">{idx}: </span>
              <JsonTree data={item} level={level + 1} />
            </div>
          ))}
        </div>
      )}
      {expanded && <span className="text-txt-m">]</span>}
    </div>
  );
}
```

- [ ] **Step 2: Integrate JsonTree in RightSidebar**

In `RightSidebar.tsx`, add import:
```typescript
import { JsonTree } from "../file/JsonTree";
```

Add JSON preview section:
```tsx
{preview.isText && preview.content && (preview.language === "json") && (() => {
  try {
    const parsed = JSON.parse(preview.content);
    return <JsonTree data={parsed} />;
  } catch {
    return <CodeBlock content={preview.content} language="json" />;
  }
})()}
```

Update the rendering logic to separate JSON from other text:
```tsx
{preview.isText && preview.content && preview.language === "markdown" && (
  <MarkdownPreview content={preview.content} />
)}

{preview.isText && preview.content && preview.language === "json" && (() => {
  try {
    const parsed = JSON.parse(preview.content);
    return <div className="p-1"><JsonTree data={parsed} /></div>;
  } catch {
    return <CodeBlock content={preview.content} language="json" />;
  }
})()}

{preview.isText && preview.content && preview.language !== "markdown" && preview.language !== "json" && (
  <CodeBlock content={preview.content} language={preview.language} />
)}
```

- [ ] **Step 3: Build and verify**

Run: `cd desktop/frontend; npm run build`
Expected: Build succeeds

---

### Task 6: Final Integration Test

- [ ] **Step 1: Run full build chain**

```powershell
cd desktop/frontend; npm run build
go build ./desktop/... ./internal/...
go vet ./desktop/... ./internal/...
go test ./desktop/... ./internal/...
```

Expected: All pass

- [ ] **Step 2: Update HANDOFF_CURRENT.md**

Document new features:
- TaskPanel: rename, archive, progress with inline editors
- File preview: Markdown rendering, JSON collapsible tree, image display

- [ ] **Step 3: Commit changes**

```bash
git add -A
git commit -m "feat: task panel rename/archive/progress + file preview enhancements"
```
