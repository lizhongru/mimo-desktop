﻿import { useEffect, useState, useRef } from "react";
import {
  Plus,
  MessageSquare,
  Trash2,
  Loader2,
  Download,
  FolderOpen,
  Settings,
  ChevronDown,
  ChevronRight,
  SquareCheck,
  Square,
  Pencil,
  X,
  Pin,
  PinOff,
  FolderInput,
  Edit3,
  Keyboard,
  HelpCircle,
  Info,
  Languages,
  Moon,
  Sun,
  Database,
  History,
  ListTodo,
  Bot,
  Search,
} from "lucide-react";
import { useSessionStore, type SessionItem } from "../../stores/sessionStore";
import { useSettingsStore } from "../../stores/settingsStore";
import { t } from "../../lib/i18n";
import { ShortcutsPanel } from "../common/ShortcutsPanel";
import { HelpLogPanel } from "../common/HelpLogPanel";
import { AboutPanel } from "../common/AboutPanel";

interface Props {
  onNewChat: () => void;
  onLoadSession: (id: string) => void;
  onDeleteSession: (id: string) => void;
  onExportSession?: (id: string) => Promise<void>;
  onOpenSettings: () => void;
  onOpenMemory: () => void;
  onOpenCheckpoint: () => void;
  onOpenTask: () => void;
  onOpenActor: () => void;
}

interface ContextMenuState {
  x: number;
  y: number;
  sessionId?: string;
  workspaceDir?: string;
}

function formatDate(dateStr: string): string {
  try {
    const d = new Date(dateStr);
    const now = new Date();
    const isToday = d.toDateString() === now.toDateString();
    if (isToday) {
      return d.toLocaleTimeString(undefined, {
        hour: "2-digit",
        minute: "2-digit",
      });
    }
    return d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
  } catch {
    return "";
  }
}

function truncate(str: string, max: number): string {
  if (!str) return "";
  return str.length > max ? str.slice(0, max) + "..." : str;
}

function sessionTitle(session: SessionItem): string {
  const title = session.firstMessage?.trim() || session.lastMessage?.trim();
  return title ? truncate(title.replace(/\s+/g, " "), 72) : t("new_chat");
}

function ContextMenu({
  menu,
  pinned,
  onClose,
  onPin,
  onOpenExplorer,
  onRename,
  onExport,
  onRemove,
}: {
  menu: ContextMenuState;
  pinned: boolean;
  onClose: () => void;
  onPin: () => void;
  onOpenExplorer: () => void;
  onRename: () => void;
  onExport: () => void;
  onRemove: () => void;
}) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose();
    };
    const handleScroll = () => onClose();
    document.addEventListener("mousedown", handleClick);
    document.addEventListener("scroll", handleScroll, true);
    return () => {
      document.removeEventListener("mousedown", handleClick);
      document.removeEventListener("scroll", handleScroll, true);
    };
  }, [onClose]);

  const items = [
    {
      icon: pinned ? PinOff : Pin,
      label: pinned ? t("unpin") : t("pin"),
      onClick: onPin,
    },
    { icon: FolderInput, label: t("open_in_explorer"), onClick: onOpenExplorer },
    ...(menu.sessionId
      ? [{ icon: Edit3, label: t("rename"), onClick: onRename }]
      : []),
    ...(menu.sessionId
      ? [{ icon: Download, label: t("export_chat"), onClick: onExport }]
      : []),
    { icon: Trash2, label: t("remove_project"), onClick: onRemove, danger: true },
  ];

  return (
    <div
      ref={ref}
      className="fixed z-[100] bg-surface border border-bdr rounded-lg shadow-xl py-1 min-w-[180px]"
      style={{ left: menu.x, top: menu.y }}
    >
      {items.map((item, i) => (
        <button
          key={i}
          onClick={() => {
            item.onClick();
            onClose();
          }}
          className={`w-full flex items-center gap-2.5 px-3 py-1.5 text-sm transition-colors cursor-pointer ${item.danger
              ? "text-red-400 hover:bg-red-500/10"
              : "text-txt-2 hover:bg-elevated"
            }`}
        >
          <item.icon className="w-3.5 h-3.5" />
          {item.label}
        </button>
      ))}
    </div>
  );
}

function SessionItemRow({
  session,
  isActive,
  isStreaming,
  isManageMode,
  isSelected,
  onLoad,
  onDelete,
  onToggleSelect,
  onContextMenu,
}: {
  session: SessionItem;
  isActive: boolean;
  isStreaming: boolean;
  isManageMode: boolean;
  isSelected: boolean;
  onLoad: (id: string) => void;
  onDelete: (id: string) => void;
  onToggleSelect: (id: string) => void;
  onContextMenu: (e: React.MouseEvent, sessionId: string) => void;
}) {
  return (
    <div
      role="button"
      tabIndex={0}
      aria-current={isActive && !isManageMode ? "page" : undefined}
      aria-label={sessionTitle(session)}
      className={`group relative mx-1.5 flex min-h-[38px] cursor-pointer items-center gap-2 rounded-lg px-2.5 py-1.5 transition-colors duration-150 ${
        isActive && !isManageMode
          ? "bg-[var(--sidebar-item-active)] text-txt shadow-[inset_2px_0_0_rgba(196,136,112,0.95)]"
          : "text-txt-2 hover:bg-[var(--sidebar-item-hover)]"
      } ${isSelected ? "bg-[var(--sidebar-accent-soft)] shadow-[inset_2px_0_0_rgba(196,136,112,0.75)]" : ""}`}
      onClick={(e) => {
        e.stopPropagation();
        if (isManageMode) {
          onToggleSelect(session.id);
        } else {
          onLoad(session.id);
        }
      }}
      onKeyDown={(e) => {
        if (e.key !== "Enter" && e.key !== " ") return;
        e.preventDefault();
        if (isManageMode) {
          onToggleSelect(session.id);
        } else {
          onLoad(session.id);
        }
      }}
      onContextMenu={(e) => onContextMenu(e, session.id)}
    >
      {isManageMode ? (
        <div className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-md">
          {isSelected ? (
            <SquareCheck className="w-4 h-4 text-accent" />
          ) : (
            <Square className="w-4 h-4 text-txt-g" />
          )}
        </div>
      ) : (
        <div
          className={`flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-md transition-colors ${
            isActive
              ? "text-accent"
              : "text-txt-g group-hover:text-txt-2"
          }`}
        >
          {isStreaming ? (
            <Loader2 className="h-3.5 w-3.5 animate-spin text-[var(--color-accent)]" />
          ) : (
            <MessageSquare className="h-3.5 w-3.5" />
          )}
        </div>
      )}
      <div className="flex min-w-0 flex-1 items-center gap-2 pr-7">
        <div className="min-w-0 flex-1 leading-none">
          <div className={`truncate text-[13px] font-medium ${isActive && !isManageMode ? "text-txt" : "text-txt-2"}`}>
            {sessionTitle(session)}
          </div>
        </div>
        {!isManageMode && (
          <span className="flex-shrink-0 text-[10px] text-txt-g transition-opacity group-hover:opacity-0">
            {formatDate(session.updatedAt)}
          </span>
        )}
      </div>
      {!isManageMode && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete(session.id);
          }}
          className="absolute right-1.5 top-1/2 flex h-7 w-7 -translate-y-1/2 items-center justify-center rounded-md text-txt-g opacity-0 transition-colors duration-150 hover:bg-red-500/10 hover:text-red-400 group-hover:opacity-100 cursor-pointer"
          aria-label={t("delete")}
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      )}
    </div>
  );
}

export function LeftSidebar({
  onNewChat,
  onLoadSession,
  onDeleteSession,
  onOpenSettings,
  onExportSession,
  onOpenMemory,
  onOpenCheckpoint,
  onOpenTask,
  onOpenActor,
}: Props) {
  const sessions = useSessionStore((s) => s.sessions);
  const currentSessionId = useSessionStore((s) => s.currentSessionId);
  const streamingSessionId = useSessionStore((s) => s.streamingSessionId);
  const selectedWorkspace = useSessionStore((s) => s.selectedWorkspace);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const workspaces = useSessionStore((s) => s.workspaces);
  const [defaultExpanded, setDefaultExpanded] = useState(true);
  const [expandedWorkspaceIds, setExpandedWorkspaceIds] = useState<Set<string>>(new Set());
  const [manageMode, setManageMode] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [pinnedDirs, setPinnedDirs] = useState<Set<string>>(new Set());
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);
  const [sessionQuery, setSessionQuery] = useState("");
  const [renameTarget, setRenameTarget] = useState<string | null>(null);
  const [renameValue, setRenameValue] = useState("");

  useEffect(() => {
    const activeWorkspace =
      sessions.find((session) => session.id === currentSessionId)?.workspaceId ||
      selectedWorkspace;

    if (!activeWorkspace || activeWorkspace === "default") {
      if (activeWorkspace === "default") setDefaultExpanded(true);
      return;
    }

    setExpandedWorkspaceIds((prev) => {
      if (prev.has(activeWorkspace)) return prev;
      const next = new Set(prev);
      next.add(activeWorkspace);
      return next;
    });
  }, [currentSessionId, selectedWorkspace, sessions]);

  // Close context menu on escape
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setContextMenu(null);
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, []);

  const handleDelete = (id: string) => setDeleteTarget(id);

  const confirmDelete = () => {
    if (deleteTarget) {
      onDeleteSession(deleteTarget);
      setDeleteTarget(null);
    }
  };

  const toggleSelect = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const handleContextMenu = (e: React.MouseEvent, sessionId: string) => {
    e.preventDefault();
    e.stopPropagation();
    const session = sessions.find((s) => s.id === sessionId);
    setContextMenu({
      x: e.clientX,
      y: e.clientY,
      sessionId,
      workspaceDir: session?.workspaceId || "",
    });
  };

  const handleWorkspaceContextMenu = (e: React.MouseEvent, dir: string) => {
    e.preventDefault();
    e.stopPropagation();
    setContextMenu({
      x: e.clientX,
      y: e.clientY,
      workspaceDir: dir,
    });
  };

  const togglePin = (dir: string) => {
    setPinnedDirs((prev) => {
      const next = new Set(prev);
      if (next.has(dir)) next.delete(dir);
      else next.add(dir);
      return next;
    });
  };

  const deleteSelected = () => {
    for (const id of selectedIds) onDeleteSession(id);
    setSelectedIds(new Set());
    setManageMode(false);
  };

  const deleteWorkspace = (dir: string) => {
    const dirSessions = sessions.filter((s) => s.workspaceId === dir);
    for (const s of dirSessions) onDeleteSession(s.id);
  };

  const confirmRename = () => {
    if (renameTarget && renameValue.trim()) {
      window.go?.desktop?.App?.RenameSession?.(renameTarget, renameValue.trim())
        .then(() => {
          useSessionStore.getState().updateSession(renameTarget, renameValue.trim());
        })
        .catch(console.error);
    }
    setRenameTarget(null);
    setRenameValue("");
  };

  // Group sessions by workspace
  const workspaceGroups = new Map<string, SessionItem[]>();

  const normalizedQuery = sessionQuery.trim().toLowerCase();
  const visibleSessions = normalizedQuery
    ? sessions.filter((session) =>
        `${session.lastMessage} ${session.modelName} ${session.userName}`
          .toLowerCase()
          .includes(normalizedQuery)
      )
    : sessions;

  for (const session of visibleSessions) {
    const wsId = session.workspaceId || "default";
    if (!workspaceGroups.has(wsId)) workspaceGroups.set(wsId, []);
    workspaceGroups.get(wsId)!.push(session);
  }

  // Build display list: folder workspaces first, then the default conversation group.
  const defaultSessions = workspaceGroups.get("default") || [];
  workspaceGroups.delete("default");
  const otherEntries = Array.from(workspaceGroups.entries()).sort(([a], [b]) => {
    const aPinned = pinnedDirs.has(a) ? 0 : 1;
    const bPinned = pinnedDirs.has(b) ? 0 : 1;
    return aPinned - bPinned;
  });

  const getWorkspaceName = (wsId: string): string => {
    if (wsId === "default") return t("conversations");
    const ws = workspaces.find((w) => w.id === wsId);
    if (ws) return ws.name;
    // Fallback: extract from wsId (e.g. "ws:D:\path" -> "path")
    if (wsId.startsWith("ws:")) {
      const parts = wsId.slice(3).replace(/\\/g, "/").split("/").filter(Boolean);
      return parts[parts.length - 1] || wsId;
    }
    return wsId;
  };

  const toggleWorkspaceExpanded = (wsId: string) => {
    setExpandedWorkspaceIds((prev) => {
      const next = new Set(prev);
      if (next.has(wsId)) {
        next.delete(wsId);
      } else {
        next.add(wsId);
      }
      return next;
    });
  };

  const renderSessionRow = (session: SessionItem) => (
    <SessionItemRow
      key={session.id}
      session={session}
      isActive={currentSessionId === session.id}
      isStreaming={streamingSessionId === session.id}
      isManageMode={manageMode}
      isSelected={selectedIds.has(session.id)}
      onLoad={onLoadSession}
      onDelete={handleDelete}
      onToggleSelect={toggleSelect}
      onContextMenu={handleContextMenu}
    />
  );

  const renderWorkspaceGroup = (wsId: string, dirSessions: SessionItem[], expanded: boolean, onToggle: () => void) => {
    const isPinned = pinnedDirs.has(wsId);
    const GroupIcon = wsId === "default" ? MessageSquare : FolderOpen;

    return (
      <div key={wsId} className="mb-1.5">
        <button
          onClick={onToggle}
          onContextMenu={(e) => handleWorkspaceContextMenu(e, wsId)}
          className={`flex min-h-[28px] w-full items-center gap-1.5 rounded-lg px-2 py-1 text-[11px] font-medium tracking-normal transition-colors cursor-pointer ${isPinned ? "text-accent" : "text-txt-m hover:bg-[var(--sidebar-item-hover)] hover:text-txt-2"
            }`}
        >
          <ChevronDown className={`h-3 w-3 flex-shrink-0 transition-transform ${expanded ? "" : "-rotate-90"}`} />
          <GroupIcon className="h-3 w-3 flex-shrink-0" />
          <span className="truncate">{getWorkspaceName(wsId)}</span>
          {isPinned && <Pin className="w-2.5 h-2.5 flex-shrink-0" />}
          {manageMode && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                const ids = dirSessions.map((s) => s.id);
                const allSelected = ids.every((id) => selectedIds.has(id));
                setSelectedIds((prev) => {
                  const next = new Set(prev);
                  if (allSelected) ids.forEach((id) => next.delete(id));
                  else ids.forEach((id) => next.add(id));
                  return next;
                });
              }}
              className="ml-auto flex h-6 w-6 items-center justify-center rounded-md text-txt-g hover:bg-accent/10 hover:text-accent cursor-pointer"
              aria-label={t("manage")}
            >
              {dirSessions.every((s) => selectedIds.has(s.id)) ? (
                <SquareCheck className="w-3 h-3 text-accent" />
              ) : (
                <Square className="w-3 h-3" />
              )}
            </button>
          )}
          {!manageMode && (
            <span className="ml-auto text-[10px] text-txt-m">{dirSessions.length}</span>
          )}
        </button>
        {expanded && <div className="mt-px space-y-px">{dirSessions.map(renderSessionRow)}</div>}
      </div>
    );
  };

  return (
    <div className="flex flex-col h-full w-[284px] relative" style={{ zIndex: 50 }}>
      {/* Header */}
      <div className="border-b border-bdr-sub bg-sidebar px-3 pb-3 pt-2.5">
        <div className="mb-2.5 flex items-center gap-1.5">
          <div className="flex-1 min-w-0">
            <div className="text-[13px] font-semibold leading-tight text-txt">{t("conversations")}</div>
            <div className="mt-0.5 text-[10px] leading-tight text-txt-g">
              {sessions.length === 0 ? "暂无会话" : `${sessions.length}`}
            </div>
          </div>
          <button
            onClick={onNewChat}
            className="flex h-8 w-8 items-center justify-center rounded-lg border border-bdr-sub bg-[var(--sidebar-control)] text-txt-2 hover:border-accent/40 hover:text-accent transition-colors cursor-pointer"
            aria-label={t("new_chat")}
            title={t("new_chat")}
          >
            <Plus className="w-4 h-4" />
          </button>
          <button
            onClick={() => {
              if (manageMode) {
                setManageMode(false);
                setSelectedIds(new Set());
              } else {
                setManageMode(true);
              }
            }}
            className={`flex h-8 w-8 items-center justify-center rounded-lg transition-colors cursor-pointer ${manageMode ? "bg-[var(--sidebar-accent-soft)] text-accent" : "text-txt-g hover:bg-[var(--sidebar-control-hover)] hover:text-txt"
              }`}
            title={t("manage")}
            aria-label={t("manage")}
          >
            {manageMode ? <X className="w-4 h-4" /> : <Pencil className="w-4 h-4" />}
          </button>
        </div>

        <div className="relative">
          <Search className="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-txt-g" />
          <input
            value={sessionQuery}
            onChange={(e) => setSessionQuery(e.target.value)}
            placeholder="搜索对话"
            className="h-8 w-full rounded-lg border border-bdr-sub bg-[var(--sidebar-control)] pl-8 pr-3 text-[12px] text-txt outline-none transition-colors placeholder:text-txt-g hover:bg-[var(--sidebar-control-hover)] focus:border-accent/40"
          />
        </div>

        {manageMode && selectedIds.size > 0 && (
          <div className="mt-2.5 flex items-center gap-2 rounded-lg bg-[var(--sidebar-control)] px-2.5 py-2 text-xs">
            <span className="text-txt-m">
              {t("selected_count")} {selectedIds.size}
            </span>
            <button
              onClick={deleteSelected}
              className="ml-auto flex items-center gap-1 px-2 py-1 rounded bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors cursor-pointer"
            >
              <Trash2 className="w-3 h-3" />
              {t("delete_selected")}
            </button>
          </div>
        )}
      </div>

      {/* Session List */}
      <div className="flex-1 overflow-y-auto px-1.5 py-2">
        {sessions.length === 0 && (
          <div className="mx-4 mt-14 text-center">
            <div className="mx-auto mb-3 flex h-8 w-8 items-center justify-center rounded-lg border border-bdr-sub text-txt-g">
              <MessageSquare className="h-4 w-4" />
            </div>
            <div className="text-xs font-medium text-txt-2">{t("no_sessions")}</div>
            <button
              onClick={onNewChat}
              className="mx-auto mt-4 flex h-8 items-center gap-1.5 rounded-lg bg-[var(--sidebar-accent-soft)] px-3 text-xs text-accent transition-colors hover:bg-accent/20 cursor-pointer"
            >
              <Plus className="h-3.5 w-3.5" />
              {t("new_chat")}
            </button>
          </div>
        )}

        {sessions.length > 0 && visibleSessions.length === 0 && (
          <div className="mx-4 mt-14 text-center">
            <div className="mx-auto mb-3 flex h-8 w-8 items-center justify-center rounded-lg border border-bdr-sub text-txt-g">
              <Search className="h-4 w-4" />
            </div>
            <div className="text-xs font-medium text-txt-2">没有匹配的对话</div>
            <button
              onClick={() => setSessionQuery("")}
              className="mx-auto mt-4 flex h-8 items-center rounded-lg bg-[var(--sidebar-control)] px-3 text-xs text-txt-2 transition-colors hover:text-accent cursor-pointer"
            >
              清除搜索
            </button>
          </div>
        )}

        {otherEntries.map(([wsId, dirSessions]) =>
          renderWorkspaceGroup(wsId, dirSessions, expandedWorkspaceIds.has(wsId), () => toggleWorkspaceExpanded(wsId))
        )}

        {defaultSessions.length > 0 && otherEntries.length === 0 && (
          <div className="space-y-px">{defaultSessions.map(renderSessionRow)}</div>
        )}
        {defaultSessions.length > 0 && otherEntries.length > 0 &&
          renderWorkspaceGroup("default", defaultSessions, defaultExpanded, () => setDefaultExpanded(!defaultExpanded))}
      </div>

      {/* Footer ? User profile menu */}
      <UserProfileFooter
        onOpenSettings={onOpenSettings}
        onOpenMemory={onOpenMemory}
        onOpenCheckpoint={onOpenCheckpoint}
        onOpenTask={onOpenTask}
        onOpenActor={onOpenActor}
      />

      {/* Context Menu */}
      {contextMenu && (
        <ContextMenu
          menu={contextMenu}
          pinned={pinnedDirs.has(contextMenu.workspaceDir || "")}
          onClose={() => setContextMenu(null)}
          onPin={() => {
            if (contextMenu.workspaceDir) togglePin(contextMenu.workspaceDir);
          }}
          onOpenExplorer={() => {
            const dir = contextMenu.workspaceDir || "";
            window.go?.desktop?.App?.OpenInExplorer?.(dir).catch(console.error);
          }}
          onRename={() => {
            if (contextMenu.sessionId) {
              setRenameTarget(contextMenu.sessionId);
              const s = sessions.find((s) => s.id === contextMenu.sessionId);
              setRenameValue(s?.lastMessage || "");
            }
          }}
          onExport={() => {
            if (contextMenu.sessionId && onExportSession) {
              onExportSession(contextMenu.sessionId);
            }
          }}
          onRemove={() => {
            if (contextMenu.sessionId) {
              onDeleteSession(contextMenu.sessionId);
            } else if (contextMenu.workspaceDir) {
              deleteWorkspace(contextMenu.workspaceDir);
            }
          }}
        />
      )}

      {/* Rename Dialog */}
      {renameTarget && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 backdrop-blur-sm">
          <div className="bg-surface border border-bdr rounded-xl w-[320px] mx-4 shadow-2xl">
            <div className="px-5 py-4">
              <h3 className="text-sm font-medium text-txt mb-3">{t("rename")}</h3>
              <input
                value={renameValue}
                onChange={(e) => setRenameValue(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && confirmRename()}
                className="w-full bg-elevated border border-bdr rounded-md px-3 py-1.5 text-sm text-txt focus:outline-none focus:border-accent/50"
                autoFocus
              />
            </div>
            <div className="flex items-center gap-2 px-5 py-3 border-t border-bdr-sub justify-end">
              <button
                onClick={() => { setRenameTarget(null); setRenameValue(""); }}
                className="px-3 py-1.5 rounded-md text-sm bg-elevated text-txt-2 hover:bg-elevated/80 transition-colors cursor-pointer"
              >
                {t("cancel")}
              </button>
              <button
                onClick={confirmRename}
                className="px-3 py-1.5 rounded-md text-sm bg-accent/20 text-accent hover:bg-accent/30 transition-colors cursor-pointer"
              >
                {t("save")}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {deleteTarget && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 backdrop-blur-sm">
          <div className="bg-surface border border-bdr rounded-xl w-[320px] mx-4 shadow-2xl">
            <div className="px-5 py-4">
              <h3 className="text-sm font-medium text-txt mb-2">
                {t("delete_session")}
              </h3>
              <p className="text-xs text-txt-g">{t("delete_confirm")}</p>
            </div>
            <div className="flex items-center gap-2 px-5 py-3 border-t border-bdr-sub justify-end">
              <button
                onClick={() => setDeleteTarget(null)}
                className="px-3 py-1.5 rounded-md text-sm bg-elevated text-txt-2 hover:bg-elevated/80 transition-colors cursor-pointer"
              >
                {t("cancel")}
              </button>
              <button
                onClick={confirmDelete}
                className="px-3 py-1.5 rounded-md text-sm bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors cursor-pointer"
              >
                {t("delete")}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );


  function UserProfileFooter({
    onOpenSettings,
    onOpenMemory,
    onOpenCheckpoint,
    onOpenTask,
    onOpenActor,
  }: {
    onOpenSettings: () => void;
    onOpenMemory: () => void;
    onOpenCheckpoint: () => void;
    onOpenTask: () => void;
    onOpenActor: () => void;
  }) {
    const language = useSettingsStore((s) => s.language);
    const theme = useSettingsStore((s) => s.theme);
    const setLanguage = useSettingsStore((s) => s.setLanguage);
    const setTheme = useSettingsStore((s) => s.setTheme);

    const footerRef = useRef<HTMLDivElement>(null);
    const [rawMenuOpen, setRawMenuOpen] = useState(false);
    const [shortcutsOpen, setShortcutsOpen] = useState(false);
    const [helpLogOpen, setHelpLogOpen] = useState(false);
    const [aboutOpen, setAboutOpen] = useState(false);
    const [langPanelOpen, setLangPanelOpen] = useState(false);
    const [menuPos, setMenuPos] = useState({ left: 0, bottom: 0 });

    const close = () => {
      setRawMenuOpen(false);
      setLangPanelOpen(false);
    };

    const openMenu = () => {
      if (footerRef.current) {
        const rect = footerRef.current.getBoundingClientRect();
        setMenuPos({ left: rect.left + 8, bottom: window.innerHeight - rect.top + 8 });
      }
      setRawMenuOpen(true);
    };

    // Click outside to close
    useEffect(() => {
      if (!rawMenuOpen) return;
      const handler = (e: MouseEvent) => {
        if (footerRef.current && !footerRef.current.contains(e.target as Node)) {
          close();
        }
      };
      document.addEventListener("mousedown", handler);
      return () => document.removeEventListener("mousedown", handler);
    }, [rawMenuOpen]);

    return (
      <div ref={footerRef} className="relative border-t border-bdr-sub px-2 py-2">
        <button
          onClick={() => (rawMenuOpen ? close() : openMenu())}
          className="flex items-center gap-2.5 w-full rounded-lg px-2 py-2 hover:bg-[var(--sidebar-item-hover)] transition-colors cursor-pointer"
        >
          <div className="w-7 h-7 rounded-lg bg-accent/15 flex items-center justify-center flex-shrink-0">
            <span className="text-xs font-semibold text-accent">M</span>
          </div>
          <div className="flex-1 min-w-0 text-left">
            <div className="text-xs font-medium text-txt truncate">MiMo User</div>
            <div className="text-[10px] text-txt-g truncate">{t("click_to_settings")}</div>
          </div>
          <ChevronDown
            className={`w-3.5 h-3.5 text-txt-g transition-transform ${rawMenuOpen ? "rotate-180" : ""}`}
          />
        </button>

        {rawMenuOpen && (
          <div
            className="fixed z-[100]"
            style={{ left: menuPos.left, bottom: menuPos.bottom }}
          >
            {/* Main menu 170px */}
            <div className="w-[170px] bg-surface border border-bdr rounded-lg shadow-xl px-2 py-1.5 animate-pop-up">
              {/* Memory */}
              <button
                onClick={() => { close(); onOpenMemory(); }}
                className="w-full flex items-center gap-2.5 px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md"
              >
                <Database className="w-3.5 h-3.5 text-txt-g" />
                {t("memory") || "记忆"}
              </button>

              {/* Checkpoint */}
              <button
                onClick={() => { close(); onOpenCheckpoint(); }}
                className="w-full flex items-center gap-2.5 px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md"
              >
                <History className="w-3.5 h-3.5 text-txt-g" />
                {t("checkpoint") || "检查点"}
              </button>

              {/* Task */}
              <button
                onClick={() => { close(); onOpenTask(); }}
                className="w-full flex items-center gap-2.5 px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md"
              >
                <ListTodo className="w-3.5 h-3.5 text-txt-g" />
                {t("task") || "任务"}
              </button>

              {/* Actor */}
              <button
                onClick={() => { close(); onOpenActor(); }}
                className="w-full flex items-center gap-2.5 px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md"
              >
                <Bot className="w-3.5 h-3.5 text-txt-g" />
                {t("actor") || "子智能体"}
              </button>

              {/* Divider */}
              <div className="border-t border-bdr-sub my-1 mx-1" />

              {/* Settings */}
              <button
                onClick={() => { onOpenSettings(); close(); }}
                className="w-full flex items-center gap-2.5 px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md"
              >
                <Settings className="w-3.5 h-3.5 text-txt-g" />
                {t("settings")}
              </button>

              {/* Shortcuts */}
              <button
                onClick={() => { close(); setShortcutsOpen(true); }}
                className="w-full flex items-center gap-2.5 px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md"
              >
                <Keyboard className="w-3.5 h-3.5 text-txt-g" />
                {t("shortcuts")}
              </button>

              {/* Help Log */}
              <button
                onClick={() => { close(); setHelpLogOpen(true); }}
                className="w-full flex items-center gap-2.5 px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md"
              >
                <HelpCircle className="w-3.5 h-3.5 text-txt-g" />
                {t("help_log")}
              </button>

              {/* About */}
              <button
                onClick={() => { close(); setAboutOpen(true); }}
                className="w-full flex items-center gap-2.5 px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md"
              >
                <Info className="w-3.5 h-3.5 text-txt-g" />
                {t("about")}
              </button>

              {/* Divider */}
              <div className="border-t border-bdr-sub my-1 mx-1" />

              {/* Language */}
              <div
                className="relative"
                onMouseEnter={() => setLangPanelOpen(true)}
                onMouseLeave={() => setLangPanelOpen(false)}
              >
                <div className="w-full flex items-center justify-between px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md">
                  <span className="flex items-center gap-2.5">
                    <Languages className="w-3.5 h-3.5 text-txt-g" />
                    {t("language")}
                  </span>
                  <span className="flex items-center gap-1 text-txt-m">
                    {language === "zh" ? t("chinese") : t("english")}
                    <ChevronRight className="w-3 h-3" />
                  </span>
                </div>

                {langPanelOpen && (
                  <div className="absolute left-full bottom-0 w-[120px] bg-surface border border-bdr rounded-lg shadow-xl px-2 py-1.5 animate-pop-up">
                    <div className="px-2.5 py-1 text-[10px] text-txt-g uppercase tracking-wider">
                      {t("language")}
                    </div>
                    <button
                      onClick={() => { setLanguage("zh"); close(); }}
                      className={`w-full flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md transition-colors cursor-pointer ${language === "zh" ? "text-accent bg-accent/10" : "text-txt-2 hover:bg-elevated"
                        }`}
                    >
                      {t("chinese")}
                      {language === "zh" && <span className="text-accent">&#10003;</span>}
                    </button>
                    <button
                      onClick={() => { setLanguage("en"); close(); }}
                      className={`w-full flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md transition-colors cursor-pointer ${language === "en" ? "text-accent bg-accent/10" : "text-txt-2 hover:bg-elevated"
                        }`}
                    >
                      {t("english")}
                      {language === "en" && <span className="text-accent">&#10003;</span>}
                    </button>
                  </div>
                )}
              </div>

              {/* Theme toggle */}
              <div className="w-full flex items-center justify-between px-2.5 py-1.5 text-xs text-txt-2 rounded-md">
                <span className="flex items-center gap-2.5">
                  {theme === "dark" ? (
                    <Moon className="w-3.5 h-3.5 text-txt-g" />
                  ) : (
                    <Sun className="w-3.5 h-3.5 text-txt-g" />
                  )}
                  {t("theme")}
                </span>
                <button
                  onClick={(e) => setTheme(theme === "dark" ? "light" : "dark", e.clientX, e.clientY)}
                  className={`relative w-9 h-5 rounded-full transition-colors duration-200 cursor-pointer flex-shrink-0 ${theme === "dark" ? "bg-accent" : "bg-elevated border border-bdr"
                    }`}
                >
                  <div
                    className={`absolute top-0.5 w-4 h-4 rounded-full bg-white shadow-md transition-transform duration-200 ${theme === "dark" ? "translate-x-4" : "translate-x-0.5"
                      }`}
                  />
                </button>
              </div>
            </div>
          </div>
        )}

        <ShortcutsPanel open={shortcutsOpen} onClose={() => setShortcutsOpen(false)} />
        <HelpLogPanel open={helpLogOpen} onClose={() => setHelpLogOpen(false)} />
        <AboutPanel open={aboutOpen} onClose={() => setAboutOpen(false)} />
      </div>
    );
  }
}















