import { useState, useCallback, useEffect, useRef } from "react";
import {
  X,
  FileText,
  FileCode,
  FileImage,
  FileType,
  Binary,
  FolderTree,
  ZoomIn,
  ZoomOut,
  RotateCcw,
  WrapText,
  ExternalLink,
  Maximize2,
  Minimize2,
} from "lucide-react";
import { MarkdownPreview } from "./MarkdownPreview";
import { JsonTree } from "./JsonTree";
import { CodeBlock } from "./CodeBlock";

// ── Types ──────────────────────────────────────────────

export interface FilePreviewData {
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

// ── Helpers ─────────────────────────────────────────────

function formatSize(bytes: number): string {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  const i = Math.min(
    Math.floor(Math.log(bytes) / Math.log(1024)),
    units.length - 1
  );
  const value = bytes / Math.pow(1024, i);
  return `${value.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

function lineCount(text: string): number {
  if (!text) return 0;
  return text.split("\n").length;
}

// ── Small components ───────────────────────────────────

function MetaChip({
  label,
  accent,
  warning,
}: {
  label: string;
  accent?: boolean;
  warning?: boolean;
}) {
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium
        ${
          warning
            ? "bg-amber-500/10 text-amber-400 border border-amber-500/20"
            : accent
            ? "bg-accent/10 text-accent border border-accent/20"
            : "bg-elevated text-txt-m border border-bdr-sub"
        }`}
    >
      {label}
    </span>
  );
}

function FileTypeIcon({ preview }: { preview: FilePreviewData }) {
  if (preview.isImage)
    return <FileImage className="w-4 h-4 text-pink-400" />;
  if (preview.language === "markdown")
    return <FileType className="w-4 h-4 text-blue-400" />;
  if (preview.language === "json")
    return <FileCode className="w-4 h-4 text-amber-400" />;
  if (preview.isText)
    return <FileCode className="w-4 h-4 text-green-400" />;
  return <Binary className="w-4 h-4 text-txt-m" />;
}

// ── Image Viewer ───────────────────────────────────────

function ImageViewer({
  mime,
  content,
  name,
}: {
  mime: string;
  content: string;
  name: string;
}) {
  const [scale, setScale] = useState(1);
  const handleZoomIn = useCallback(
    () => setScale((s) => Math.min(s + 0.25, 5)),
    []
  );
  const handleZoomOut = useCallback(
    () => setScale((s) => Math.max(s - 0.25, 0.25)),
    []
  );
  const handleReset = useCallback(() => setScale(1), []);

  return (
    <div className="flex flex-col h-full">
      <div className="flex-shrink-0 flex items-center gap-1 px-3 py-1.5 border-b border-bdr-sub bg-surface/50">
        <button
          onClick={handleZoomOut}
          className="p-1 rounded hover:bg-elevated text-txt-g hover:text-txt cursor-pointer"
          title="Zoom out"
        >
          <ZoomOut className="w-3.5 h-3.5" />
        </button>
        <span className="text-[10px] text-txt-m font-mono min-w-[36px] text-center">
          {Math.round(scale * 100)}%
        </span>
        <button
          onClick={handleZoomIn}
          className="p-1 rounded hover:bg-elevated text-txt-g hover:text-txt cursor-pointer"
          title="Zoom in"
        >
          <ZoomIn className="w-3.5 h-3.5" />
        </button>
        <button
          onClick={handleReset}
          className="p-1 rounded hover:bg-elevated text-txt-g hover:text-txt cursor-pointer"
          title="Reset zoom"
        >
          <RotateCcw className="w-3 h-3" />
        </button>
      </div>
      <div className="flex-1 overflow-auto flex items-center justify-center p-6 bg-[#1a1a1a]">
        <img
          src={`data:${mime};base64,${content}`}
          alt={name}
          style={{
            transform: `scale(${scale})`,
            transformOrigin: "center center",
          }}
          className="max-w-none object-contain rounded-lg shadow-lg transition-transform duration-150"
          draggable={false}
        />
      </div>
    </div>
  );
}

// ── Modal Component ────────────────────────────────────

interface Props {
  preview: FilePreviewData | null;
  onClose: () => void;
  onOpenInExplorer?: (path: string) => void;
}

export function FilePreviewModal({
  preview,
  onClose,
  onOpenInExplorer,
}: Props) {
  const [wrapLines, setWrapLines] = useState(false);
  const [maximized, setMaximized] = useState(false);
  const [visible, setVisible] = useState(false);
  const [animating, setAnimating] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const [scrollInfo, setScrollInfo] = useState({
    top: 0,
    height: 0,
    scrollHeight: 0,
  });

  // Enter animation
  useEffect(() => {
    if (preview) {
      setAnimating(true);
      requestAnimationFrame(() => {
        setVisible(true);
      });
    } else {
      setVisible(false);
    }
  }, [preview]);

  // Close on Escape
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") handleClose();
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, []);

  const handleClose = useCallback(() => {
    setVisible(false);
    setTimeout(() => {
      setAnimating(false);
      onClose();
    }, 200);
  }, [onClose]);

  const handleScroll = useCallback(() => {
    const el = scrollRef.current;
    if (!el) return;
    setScrollInfo({
      top: el.scrollTop,
      height: el.clientHeight,
      scrollHeight: el.scrollHeight,
    });
  }, []);

  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    handleScroll();
    el.addEventListener("scroll", handleScroll, { passive: true });
    return () => el.removeEventListener("scroll", handleScroll);
  }, [handleScroll]);

  if (!preview && !animating) return null;

  const p = preview!;

  const scrollPct =
    scrollInfo.scrollHeight > scrollInfo.height
      ? Math.round(
          (scrollInfo.top / (scrollInfo.scrollHeight - scrollInfo.height)) * 100
        )
      : 0;
  const showStatus = p.isText && p.content && !p.isImage;

  const modalClasses = maximized
    ? "fixed inset-0"
    : "fixed inset-4 sm:inset-8 md:inset-12 lg:inset-16 rounded-xl";

  return (
    <div
      className={`fixed inset-0 z-[1000] flex items-center justify-center file-preview-modal
        transition-opacity duration-200 ease-out
        ${visible ? "opacity-100 pointer-events-auto" : "opacity-0 pointer-events-none"}`}
    >
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/70 dark:bg-black/80 backdrop-blur-md"
        onClick={handleClose}
      />

      {/* Modal dialog */}
      <div
        className={`${modalClasses} flex flex-col overflow-hidden
          bg-surface border border-bdr
          transition-all duration-[250ms] ease-[cubic-bezier(0.16,1,0.3,1)]
          ${
            visible
              ? "scale-100 translate-y-0 opacity-100"
              : "scale-[0.97] translate-y-3 opacity-0"
          }`}
        style={{
          boxShadow:
            "0 24px 64px rgba(0,0,0,0.5), 0 0 0 1px var(--border-default)",
        }}
      >
        {/* Header */}
        <div className="flex-shrink-0 px-4 py-2.5 bg-surface border-b border-bdr-sub flex items-center gap-2">
          <FileTypeIcon preview={p} />
          <div className="flex-1 min-w-0">
            <div className="text-sm font-medium truncate text-txt">
              {p.name}
            </div>
          </div>

          <div className="flex items-center gap-1.5 flex-shrink-0">
            {p.truncated && <MetaChip label="Truncated" warning />}
            {p.sizeBytes > 0 && <MetaChip label={formatSize(p.sizeBytes)} />}
            {p.language && p.isText && (
              <MetaChip label={p.language} accent />
            )}
          </div>

          <div className="w-px h-4 bg-bdr mx-1.5" />

          {p.isText && p.content && !p.isImage && (
            <button
              onClick={() => setWrapLines((w) => !w)}
              className={`p-1.5 rounded transition-colors cursor-pointer ${
                wrapLines
                  ? "bg-accent/20 text-accent"
                  : "hover:bg-elevated text-txt-g hover:text-txt"
              }`}
              title="Toggle line wrap"
            >
              <WrapText className="w-3.5 h-3.5" />
            </button>
          )}

          {p.path && onOpenInExplorer && (
            <button
              onClick={() => onOpenInExplorer(p.path)}
              className="p-1.5 rounded hover:bg-elevated text-txt-g hover:text-txt transition-colors cursor-pointer"
              title="Open in explorer"
            >
              <ExternalLink className="w-3.5 h-3.5" />
            </button>
          )}

          <button
            onClick={() => setMaximized((m) => !m)}
            className="p-1.5 rounded hover:bg-elevated text-txt-g hover:text-txt transition-colors cursor-pointer"
            title={maximized ? "Restore" : "Maximize"}
          >
            {maximized ? (
              <Minimize2 className="w-3.5 h-3.5" />
            ) : (
              <Maximize2 className="w-3.5 h-3.5" />
            )}
          </button>

          <button
            onClick={handleClose}
            className="p-1.5 rounded hover:bg-red-500/20 text-txt-g hover:text-red-400 transition-colors cursor-pointer"
            title="Close (Esc)"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Content */}
        <div
          ref={scrollRef}
          className="flex-1 overflow-y-auto bg-root select-text cursor-auto"
          onScroll={handleScroll}
        >
          {/* Directory */}
          {p.isDir && (
            <div className="flex flex-col items-center justify-center h-full text-xs text-txt-m space-y-3 p-6">
              <div className="w-12 h-12 rounded-full bg-elevated flex items-center justify-center">
                <FolderTree className="w-6 h-6 text-txt-g" />
              </div>
              <div className="text-center">
                <div className="font-medium text-txt-2 mb-1">Directory</div>
                <div className="text-txt-m text-[11px]">
                  Select a file from the tree to preview it.
                </div>
              </div>
            </div>
          )}

          {/* Empty file */}
          {p.isText && !p.content && (
            <div className="flex flex-col items-center justify-center h-full text-xs text-txt-m space-y-3 p-6">
              <div className="w-12 h-12 rounded-full bg-elevated flex items-center justify-center">
                <FileText className="w-6 h-6 text-txt-g/50" />
              </div>
              <div className="text-center">
                <div className="font-medium text-txt-2 mb-1">Empty File</div>
                <div className="text-txt-m text-[11px]">
                  This file has no content.
                </div>
              </div>
            </div>
          )}

          {/* Image preview */}
          {p.isImage && (
            <ImageViewer mime={p.mime} content={p.content} name={p.name} />
          )}

          {/* Markdown */}
          {p.isText && p.content && p.language === "markdown" && (
            <div className="p-6">
              <MarkdownPreview content={p.content} />
            </div>
          )}

          {/* JSON */}
          {p.isText &&
            p.content &&
            p.language === "json" &&
            (() => {
              try {
                const parsed = JSON.parse(p.content);
                return (
                  <div className="p-4">
                    <JsonTree data={parsed} />
                  </div>
                );
              } catch {
                return (
                  <div className="p-4">
                    <CodeBlock
                      content={p.content}
                      language="json"
                      wrapLines={wrapLines}
                    />
                  </div>
                );
              }
            })()}

          {/* Code / other text */}
          {p.isText &&
            p.content &&
            p.language !== "markdown" &&
            p.language !== "json" && (
              <div className="p-4">
                <CodeBlock
                  content={p.content}
                  language={p.language}
                  wrapLines={wrapLines}
                />
              </div>
            )}
        </div>

        {/* Bottom status bar */}
        {showStatus && (
          <div className="flex-shrink-0 px-4 py-1.5 border-t border-bdr-sub bg-surface flex items-center justify-between text-[11px] text-txt-m">
            <span>{scrollPct}%</span>
            <span>
              {lineCount(p.content)} lines · {formatSize(p.sizeBytes)}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
