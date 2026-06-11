import { useState, useCallback } from "react";
import { X, Plus, MessageSquare } from "lucide-react";
import { useSessionStore } from "../../stores/sessionStore";

interface Props {
  onNewChat: () => void;
  onLoadSession: (id: string) => void;
  onCloseSession: (id: string) => void;
}

export function SessionTabs({ onNewChat, onLoadSession, onCloseSession }: Props) {
  const sessions = useSessionStore((s) => s.sessions);
  const currentSessionId = useSessionStore((s) => s.currentSessionId);
  const [hoveredId, setHoveredId] = useState<string | null>(null);

  // Show only recent 10 sessions
  const recentSessions = sessions.slice(0, 10);

  const handleClose = useCallback(
    (e: React.MouseEvent, id: string) => {
      e.stopPropagation();
      onCloseSession(id);
    },
    [onCloseSession]
  );

  const truncate = (str: string, max: number) => {
    if (!str) return "New Chat";
    return str.length > max ? str.slice(0, max) + "..." : str;
  };

  return (
    <div className="flex items-center border-b border-bdr bg-surface overflow-x-auto">
      {/* New chat button */}
      <button
        onClick={onNewChat}
        className="flex-shrink-0 p-2 hover:bg-elevated transition-colors border-r border-bdr"
        title="New Chat"
      >
        <Plus className="w-4 h-4 text-txt-m" />
      </button>

      {/* Session tabs */}
      <div className="flex overflow-x-auto">
        {recentSessions.map((session) => {
          const isActive = session.id === currentSessionId;
          const isHovered = hoveredId === session.id;

          return (
            <div
              key={session.id}
              onClick={() => onLoadSession(session.id)}
              onMouseEnter={() => setHoveredId(session.id)}
              onMouseLeave={() => setHoveredId(null)}
              className={`
                group flex items-center gap-2 px-3 py-2 cursor-pointer
                border-r border-bdr min-w-[120px] max-w-[180px]
                transition-colors
                ${isActive
                  ? "bg-bg text-txt-1 border-b-2 border-b-accent"
                  : "text-txt-2 hover:bg-elevated"
                }
              `}
            >
              <MessageSquare className="w-3.5 h-3.5 flex-shrink-0 text-txt-m" />
              <span className="text-sm truncate flex-1">
                {truncate(session.lastMessage || "New Chat", 20)}
              </span>

              {/* Close button */}
              {(isHovered || isActive) && sessions.length > 1 && (
                <button
                  onClick={(e) => handleClose(e, session.id)}
                  className="flex-shrink-0 p-0.5 rounded hover:bg-red-500/20 text-txt-m hover:text-red-400"
                >
                  <X className="w-3 h-3" />
                </button>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
