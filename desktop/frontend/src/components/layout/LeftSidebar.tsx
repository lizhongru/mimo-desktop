import { useEffect, useState, useRef } from "react";
import {
  Plus,
  MessageSquare,
  Trash2,
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
} from "lucide-react";
import { useSessionStore, type SessionItem, type WorkspaceItem } from "../../stores/sessionStore";
import { useSettingsStore } from "../../stores/settingsStore";
import { t } from "../../lib/i18n";
import { ShortcutsPanel } from "../common/ShortcutsPanel";
import { HelpLogPanel } from "../common/HelpLogPanel";
import { AboutPanel } from "../common/AboutPanel";
import { animateThemeSwitch } from "../../lib/theme-transition";

interface Props {
  onNewChat: () => void;
  onLoadSession: (id: string) => void;
  onDeleteSession: (id: string) => void;
  onExportSession?: (id: string) => Promise<void>;
  onOpenSettings: () => void;
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

function getDirName(path: string): string {
  if (!path) return "其他";
  const parts = path.replace(/\\/g, "/").split("/");
  return parts[parts.length - 1] || path;
}

function ContextMenu({
  menu,
  pinned,
  onClose,
  onPin,
  onOpenExplorer,
  onRename,
  onRemove,
}: {
  menu: ContextMenuState;
  pinned: boolean;
  onClose: () => void;
  onPin: () => void;
  onOpenExplorer: () => void;
  onRename: () => void;
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
          className={`w-full flex items-center gap-2.5 px-3 py-1.5 text-sm transition-colors cursor-pointer ${
            item.danger
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
  isManageMode,
  isSelected,
  onLoad,
  onDelete,
  onToggleSelect,
  onContextMenu,
}: {
  session: SessionItem;
  isActive: boolean;
  isManageMode: boolean;
  isSelected: boolean;
  onLoad: (id: string) => void;
  onDelete: (id: string) => void;
  onToggleSelect: (id: string) => void;
  onContextMenu: (e: React.MouseEvent, sessionId: string) => void;
}) {
  return (
    <div
      className={`group flex items-start gap-2 px-3 py-2 mx-2 rounded-md cursor-pointer transition-colors ${
        isActive && !isManageMode
          ? "bg-accent/10 border-l-2 border-accent"
          : "border-l-2 border-transparent hover:bg-elevated/40"
      } ${isSelected ? "bg-accent/10" : ""}`}
      onClick={(e) => {
        e.stopPropagation();
        if (isManageMode) {
          onToggleSelect(session.id);
        } else {
          onLoad(session.id);
        }
      }}
      onContextMenu={(e) => onContextMenu(e, session.id)}
    >
      {isManageMode ? (
        <div className="flex-shrink-0 mt-0.5">
          {isSelected ? (
            <SquareCheck className="w-4 h-4 text-accent" />
          ) : (
            <Square className="w-4 h-4 text-txt-g" />
          )}
        </div>
      ) : (
        <MessageSquare className={`w-3.5 h-3.5 mt-0.5 flex-shrink-0 ${isActive ? "text-accent" : "text-txt-g"}`} />
      )}
      <div className="flex-1 min-w-0">
        <div className={`text-sm truncate ${isActive ? "text-accent" : "text-txt"}`}>
          {truncate(session.lastMessage, 35)}
        </div>
        <div className="text-[10px] text-txt-m mt-0.5">
          {formatDate(session.updatedAt)}
        </div>
      </div>
      {!isManageMode && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete(session.id);
          }}
          className="opacity-0 group-hover:opacity-100 p-1 hover:text-red-400 text-txt-g transition-all cursor-pointer"
        >
          <Trash2 className="w-3 h-3" />
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
}: Props) {
  const sessions = useSessionStore((s) => s.sessions);
  const currentSessionId = useSessionStore((s) => s.currentSessionId);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const workspaces = useSessionStore((s) => s.workspaces);
  const [defaultExpanded, setDefaultExpanded] = useState(true);
  const [expandedWorkspaceIds, setExpandedWorkspaceIds] = useState<Set<string>>(new Set());
  const [manageMode, setManageMode] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [pinnedDirs, setPinnedDirs] = useState<Set<string>>(new Set());
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);
  const [menuOpen, setMenuOpen] = useState(false);
  const [shortcutsOpen, setShortcutsOpen] = useState(false);
  const [helpLogOpen, setHelpLogOpen] = useState(false);
  const [aboutOpen, setAboutOpen] = useState(false);
  const [langPanelOpen, setLangPanelOpen] = useState(false);
  const theme = useSettingsStore((s) => s.theme);
  const language = useSettingsStore((s) => s.language);
  const setTheme = useSettingsStore((s) => s.setTheme);
  const setLanguage = useSettingsStore((s) => s.setLanguage);
  const [renameTarget, setRenameTarget] = useState<string | null>(null);
  const [renameValue, setRenameValue] = useState("");

  useEffect(() => {
  // Workspaces are loaded in App.tsx on mount



  }, []);

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

  for (const session of sessions) {
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
    const wsManage = manageMode && selectedIds.size > 0;
    const GroupIcon = wsId === "default" ? MessageSquare : FolderOpen;

    return (
      <div key={wsId} className="mb-1">
        <button
          onClick={onToggle}
          onContextMenu={(e) => handleWorkspaceContextMenu(e, wsId)}
          className={`flex items-center gap-1.5 px-4 py-1.5 text-[10px] uppercase tracking-wider w-full transition-colors cursor-pointer ${
            isPinned ? "text-accent" : "text-txt-m hover:text-txt-g"
          }`}
        >
          <ChevronDown className={`w-3 h-3 transition-transform ${expanded ? "" : "-rotate-90"}`} />
          <GroupIcon className="w-3 h-3" />
          <span className="truncate">{getWorkspaceName(wsId)}</span>
          {isPinned && <Pin className="w-2.5 h-2.5 ml-0.5" />}
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
              className="ml-auto text-txt-g hover:text-accent cursor-pointer"
            >
              {dirSessions.every((s) => selectedIds.has(s.id)) ? (
                <SquareCheck className="w-3 h-3 text-accent" />
              ) : (
                <Square className="w-3 h-3" />
              )}
            </button>
          )}
          {!manageMode && (
            <span className="ml-auto text-txt-g">{dirSessions.length}</span>
          )}
        </button>
        {expanded && dirSessions.map(renderSessionRow)}
      </div>
    );
  };

  return (
    <div className="flex flex-col h-full w-[260px] relative" style={{ zIndex: 50 }}>
      {/* Header */}
      <div className="p-3 border-b border-bdr-sub space-y-2">
        <div className="flex items-center gap-2">
          <button
            onClick={onNewChat}
            className="flex-1 flex items-center gap-2 px-3 py-2 rounded-lg bg-accent/15 text-accent hover:bg-accent/25 transition-colors text-sm font-medium cursor-pointer"
          >
            <Plus className="w-4 h-4" />
            {t("new_chat")}
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
            className={`p-2 rounded-lg transition-colors cursor-pointer ${
              manageMode ? "bg-accent/20 text-accent" : "bg-elevated text-txt-g hover:text-txt"
            }`}
            title={t("manage")}
          >
            {manageMode ? <X className="w-4 h-4" /> : <Pencil className="w-4 h-4" />}
          </button>
        </div>

        {manageMode && selectedIds.size > 0 && (
          <div className="flex items-center gap-2 text-xs">
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
      <div className="flex-1 overflow-y-auto py-1">
        {sessions.length === 0 && (
          <div className="px-4 py-8 text-center text-txt-m text-xs">
            {t("no_sessions")}
          </div>
        )}

        {otherEntries.map(([wsId, dirSessions]) =>
          renderWorkspaceGroup(wsId, dirSessions, expandedWorkspaceIds.has(wsId), () => toggleWorkspaceExpanded(wsId))
        )}

        {defaultSessions.length > 0 &&
          renderWorkspaceGroup("default", defaultSessions, defaultExpanded, () => setDefaultExpanded(!defaultExpanded))}
      </div>

      {/* Footer ? User profile menu */}
      <UserProfileFooter onOpenSettings={onOpenSettings} />

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
}: {
  onOpenSettings: () => void;
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
    <div ref={footerRef} className="relative border-t border-bdr-sub">
      <button
        onClick={() => (rawMenuOpen ? close() : openMenu())}
        className="flex items-center gap-2.5 w-full px-3 py-2.5 rounded-md hover:bg-elevated transition-colors cursor-pointer"
      >
        <div className="w-7 h-7 rounded-full bg-accent/20 flex items-center justify-center flex-shrink-0">
          <span className="text-xs font-bold text-accent">M</span>
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
                    className={`w-full flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md transition-colors cursor-pointer ${
                      language === "zh" ? "text-accent bg-accent/10" : "text-txt-2 hover:bg-elevated"
                    }`}
                  >
                    {t("chinese")}
                    {language === "zh" && <span className="text-accent">&#10003;</span>}
                  </button>
                  <button
                    onClick={() => { setLanguage("en"); close(); }}
                    className={`w-full flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md transition-colors cursor-pointer ${
                      language === "en" ? "text-accent bg-accent/10" : "text-txt-2 hover:bg-elevated"
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
                className={`relative w-9 h-5 rounded-full transition-colors duration-200 cursor-pointer flex-shrink-0 ${
                  theme === "dark" ? "bg-accent" : "bg-elevated border border-bdr"
                }`}
              >
                <div
                  className={`absolute top-0.5 w-4 h-4 rounded-full bg-white shadow-md transition-transform duration-200 ${
                    theme === "dark" ? "translate-x-4" : "translate-x-0.5"
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
