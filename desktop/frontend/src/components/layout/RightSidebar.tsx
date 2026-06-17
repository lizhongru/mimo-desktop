import {
  Wrench,
  CheckCircle,
  Loader2,
  XCircle,
  FileText,
  ChevronDown,
  ChevronRight,
  ListChecks,
  X,
  FolderTree,
  Activity,
  AlertCircle,
} from "lucide-react";
import { useState, useCallback } from "react";
import { useActivityStore, type ActivityEntry } from "../../stores/activityStore";
import { FileTree, FileTreeRefreshButton } from "../file/FileTree";
import { FilePreviewModal, type FilePreviewData } from "../file/FilePreviewModal";
import { t } from "../../lib/i18n";

// ── Activity ───────────────────────────────────────────

function ActivityEntryItem({ entry }: { entry: ActivityEntry }) {
  const [expanded, setExpanded] = useState(false);

  const statusIcon =
    entry.status === "running" ? (
      <span className="relative flex h-4 w-4 flex-shrink-0 items-center justify-center">
        <span className="absolute inline-flex h-4 w-4 rounded-full bg-amber-400/30 animate-ping" />
        <Loader2 className="relative w-3.5 h-3.5 text-amber-400 animate-spin" />
      </span>
    ) : entry.status === "done" ? (
      <CheckCircle className="w-3.5 h-3.5 text-green-500" />
    ) : (
      <XCircle className="w-3.5 h-3.5 text-red-500" />
    );

  return (
    <div className="border-b border-bdr-sub/50 last:border-0">
      <button
        onClick={() => setExpanded(!expanded)}
        className={`flex items-center gap-2 w-full px-3 py-2 text-xs hover:bg-elevated/30 transition-colors cursor-pointer ${entry.status === "running" ? "bg-amber-400/5" : ""}`}
      >
        {expanded ? (
          <ChevronDown className="w-3 h-3 text-txt-g flex-shrink-0" />
        ) : (
          <ChevronRight className="w-3 h-3 text-txt-g flex-shrink-0" />
        )}
        {statusIcon}
        <span className="font-mono text-txt-2 truncate">{entry.name}</span>
        {entry.count > 1 && (
          <span className="px-1.5 py-0.5 rounded-full bg-elevated text-[10px] text-txt-m flex-shrink-0">
            ×{entry.count}
          </span>
        )}
        <span className="ml-auto text-txt-m text-[10px] flex-shrink-0">
          {new Date(entry.lastUpdated).toLocaleTimeString(undefined, {
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
          })}
        </span>
      </button>
      {expanded && entry.detail && (
        <pre className="px-5 pb-2 text-[11px] text-txt-g whitespace-pre-wrap break-all max-h-40 overflow-y-auto">
          {entry.detail.length > 500
            ? entry.detail.slice(0, 500) + "..."
            : entry.detail}
        </pre>
      )}
    </div>
  );
}

// ── Inline toast for file preview errors ───────────────

function PreviewToast({ message, onDismiss }: { message: string; onDismiss: () => void }) {
  return (
    <div className="absolute bottom-3 left-3 right-3 z-10 animate-fade-in">
      <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-red-500/10 border border-red-500/20 text-xs text-red-400">
        <AlertCircle className="w-3.5 h-3.5 flex-shrink-0" />
        <span className="flex-1 min-w-0 truncate">{message}</span>
        <button
          onClick={onDismiss}
          className="p-0.5 rounded hover:bg-red-500/20 cursor-pointer flex-shrink-0"
        >
          <X className="w-3 h-3" />
        </button>
      </div>
    </div>
  );
}

// ── Right Sidebar ──────────────────────────────────────

type RightTab = "activity" | "files";

const MIN_WIDTH = 320;
const MAX_WIDTH = 800;
const DEFAULT_WIDTH = 320;

export function RightSidebar({ width, onWidthChange, onDragStart, onDragEnd }: { width: number; onWidthChange: (w: number) => void; onDragStart?: () => void; onDragEnd?: () => void }) {
  const entries = useActivityStore((s) => s.entries);
  const plan = useActivityStore((s) => s.plan);
  const fileDiffs = useActivityStore((s) => s.fileDiffs);
  const setRightSidebarOpen = useActivityStore((s) => s.setRightSidebarOpen);

  const [tab, setTab] = useState<RightTab>("activity");
  const [treeKey, setTreeKey] = useState(0);
  const [preview, setPreview] = useState<FilePreviewData | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState<string | null>(null);

  const handleFileClick = useCallback(async (path: string, isDir: boolean) => {
    setPreviewLoading(true);
    setPreviewError(null);
    try {
      const data = await window.go?.desktop?.App?.ReadFilePreview?.(path);
      if (data) {
        setPreview(data as FilePreviewData);
      } else {
        setPreviewError("No preview data returned.");
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      setPreviewError(msg || "Failed to load preview");
    } finally {
      setPreviewLoading(false);
    }
  }, []);

  const handleRefreshTree = useCallback(() => {
    setPreview(null);
    setPreviewError(null);
    setPreviewLoading(false);
    setTreeKey((k) => k + 1);
  }, []);

  // Drag handle
  const handleDrag = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      onDragStart?.();
      const startX = e.clientX;
      const startWidth = width;

      const onMouseMove = (ev: MouseEvent) => {
        const delta = startX - ev.clientX;
        const newWidth = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, startWidth + delta));
        onWidthChange(newWidth);
      };
      const onMouseUp = () => {
        onDragEnd?.();
        document.removeEventListener("mousemove", onMouseMove);
        document.removeEventListener("mouseup", onMouseUp);
      };
      document.addEventListener("mousemove", onMouseMove);
      document.addEventListener("mouseup", onMouseUp);
    },
    [width, onWidthChange, onDragStart, onDragEnd]
  );

  return (
    <div
      className="flex-shrink-0 flex flex-col bg-bg border-l border-bdr relative"
      style={{ width }}
    >
      {/* Drag handle */}
      <div
        onMouseDown={handleDrag}
        className="absolute left-0 top-0 bottom-0 w-1 cursor-col-resize z-20 hover:bg-accent/30 transition-colors"
        title="Drag to resize"
      />

      {/* Main content */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Header: tabs + close */}
        <div className="px-2 py-1.5 border-b border-bdr-sub flex items-center gap-0.5">
          <button
            onClick={() => setTab("activity")}
            className={`flex items-center gap-1.5 px-2 py-1 rounded text-xs cursor-pointer transition-colors ${
              tab === "activity"
                ? "bg-elevated text-txt font-medium"
                : "text-txt-g hover:text-txt hover:bg-elevated/50"
            }`}
          >
            <Activity className="w-3.5 h-3.5" />
            <span>{t("activity")}</span>
            {entries.length > 0 && (
              <span className="text-[10px] text-txt-m ml-0.5">
                {entries.length}
              </span>
            )}
          </button>
          <button
            onClick={() => setTab("files")}
            className={`flex items-center gap-1.5 px-2 py-1 rounded text-xs cursor-pointer transition-colors ${
              tab === "files"
                ? "bg-elevated text-txt font-medium"
                : "text-txt-g hover:text-txt hover:bg-elevated/50"
            }`}
          >
            <FolderTree className="w-3.5 h-3.5" />
            <span>Files</span>
          </button>

          <div className="ml-auto flex items-center gap-1">
            {tab === "files" && (
              <FileTreeRefreshButton onClick={handleRefreshTree} />
            )}
            <button
              onClick={() => setRightSidebarOpen(false)}
              className="p-1 rounded hover:bg-elevated text-txt-g hover:text-txt transition-colors cursor-pointer"
              title={`${t("close")} (Ctrl+I)`}
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        {/* Tab content */}
        {tab === "files" ? (
          <div className="flex-1 overflow-y-auto relative">
            {/* File tree — 始终显示，不被 loading/error 遮挡 */}
            <FileTree key={treeKey} onFileClick={handleFileClick} />

            {/* Loading indicator — 浮在底部 */}
            {previewLoading && (
              <div className="absolute bottom-3 left-3 right-3 z-10">
                <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-surface border border-bdr-sub text-xs text-txt-m shadow-lg">
                  <Loader2 className="w-3.5 h-3.5 text-accent animate-spin flex-shrink-0" />
                  <span>Loading preview...</span>
                </div>
              </div>
            )}

            {/* Error toast */}
            {previewError && (
              <PreviewToast
                message={previewError}
                onDismiss={() => setPreviewError(null)}
              />
            )}
          </div>
        ) : (
          <div className="flex-1 overflow-y-auto">
            {/* Plan Panel */}
            {plan && (
            <div className="border-b border-bdr-sub p-3">
              <div className="flex items-center gap-2 mb-2">
                <ListChecks className="w-3.5 h-3.5 text-accent" />
                <span className="text-xs font-medium text-txt-2">
                  {plan.goal}
                </span>
              </div>
              <div className="w-full h-1.5 bg-elevated rounded-full mb-2 overflow-hidden">
                <div
                  className="h-full bg-accent rounded-full transition-all duration-300"
                  style={{
                    width: `${
                      plan.totalSteps > 0
                        ? (plan.steps.filter((s) => s.status === "completed")
                            .length /
                            plan.totalSteps) *
                          100
                        : 0
                    }%`,
                  }}
                />
              </div>
              <div className="space-y-1">
                {plan.steps.map((step) => (
                  <div
                    key={step.id}
                    className="flex items-center gap-2 text-xs py-0.5"
                  >
                    {step.status === "completed" ? (
                      <CheckCircle className="w-3 h-3 text-green-500 flex-shrink-0" />
                    ) : step.status === "in_progress" ? (
                      <Loader2 className="w-3 h-3 text-amber-400 animate-spin flex-shrink-0" />
                    ) : step.status === "failed" ? (
                      <XCircle className="w-3 h-3 text-red-500 flex-shrink-0" />
                    ) : (
                      <span className="w-3 h-3 rounded-full border border-txt-g flex-shrink-0" />
                    )}
                    <span
                      className={`${
                        step.status === "in_progress"
                          ? "text-txt"
                          : step.status === "completed"
                          ? "text-txt-m"
                          : "text-txt-g"
                      }`}
                    >
                      {step.description}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* File Changes */}
          {fileDiffs.length > 0 && (
            <div className="border-b border-bdr-sub p-3">
              <div className="flex items-center gap-2 mb-2">
                <FileText className="w-3.5 h-3.5 text-blue-400" />
                <span className="text-xs font-medium text-txt-2">
                  {t("file_changes")}
                </span>
              </div>
              <div className="space-y-1">
                {fileDiffs.map((diff) => (
                  <div
                    key={diff.path}
                    className="flex items-center gap-2 text-xs font-mono"
                  >
                    <span className="text-txt-2 truncate flex-1">
                      {diff.path}
                    </span>
                    <span className="text-green-500 text-[10px]">
                      +{diff.additions}
                    </span>
                    <span className="text-red-400 text-[10px]">
                      -{diff.deletions}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

            {/* Activity Log */}
            <div>
              {entries.length === 0 && (
                <div className="px-4 py-8 text-center text-txt-m text-xs">
                  {t("no_activity")}
                </div>
              )}
              {entries.map((entry) => (
                <ActivityEntryItem key={entry.id} entry={entry} />
              ))}
            </div>
          </div>
        )}
      </div>

      {/* File Preview Modal */}
      <FilePreviewModal
        preview={preview}
        onClose={() => setPreview(null)}
        onOpenInExplorer={(path) => window.go?.desktop?.App?.OpenInExplorer?.(path)}
      />
    </div>
  );
}
