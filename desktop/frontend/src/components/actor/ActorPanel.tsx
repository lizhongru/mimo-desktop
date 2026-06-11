import { useState, useCallback, useEffect } from "react";
import {
  Bot,
  Play,
  CheckCircle2,
  XCircle,
  Loader2,
  Trash2,
  RefreshCw,
} from "lucide-react";

interface ActorInfo {
  id: string;
  type: string;
  session_id: string;
  parent_id?: string;
  status: string;
  prompt: string;
  result?: string;
  error?: string;
  created_at: number;
  started_at?: number;
  completed_at?: number;
}

interface ActorResult {
  success: boolean;
  message: string;
  actor?: ActorInfo;
}

const STATUS_COLORS: Record<string, string> = {
  pending: "text-yellow-400",
  running: "text-blue-400",
  completed: "text-green-400",
  failed: "text-red-400",
  cancelled: "text-txt-m",
};

const STATUS_ICONS: Record<string, typeof Bot> = {
  pending: Loader2,
  running: Loader2,
  completed: CheckCircle2,
  failed: XCircle,
  cancelled: XCircle,
};

const ACTOR_TYPES = [
  { value: "explore", label: "探索" },
  { value: "general", label: "通用" },
  { value: "title", label: "标题生成" },
  { value: "summary", label: "摘要" },
];

export function ActorPanel() {
  const [actors, setActors] = useState<ActorInfo[]>([]);
  const [actorType, setActorType] = useState("explore");
  const [prompt, setPrompt] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [lastResult, setLastResult] = useState<ActorResult | null>(null);

  const loadActors = useCallback(async () => {
    try {
      const list = await window.go?.desktop?.App?.ActorList?.("");
      setActors(list || []);
    } catch (error) {
      console.error("Failed to load actors:", error);
    }
  }, []);

  useEffect(() => {
    loadActors();
    // Auto-refresh every 2 seconds
    const interval = setInterval(loadActors, 2000);
    return () => clearInterval(interval);
  }, [loadActors]);

  const handleSpawn = useCallback(async () => {
    if (!prompt.trim()) return;

    setIsLoading(true);
    try {
      const result = await window.go?.desktop?.App?.ActorSpawn?.(
        actorType,
        prompt,
        ""
      );
      setLastResult(result);
      if (result?.success) {
        setPrompt("");
        loadActors();
      }
    } catch (error) {
      console.error("Failed to spawn actor:", error);
      setLastResult({ success: false, message: "Failed to spawn actor" });
    } finally {
      setIsLoading(false);
    }
  }, [actorType, prompt, loadActors]);

  const handleCancel = useCallback(
    async (id: string) => {
      setIsLoading(true);
      try {
        const result = await window.go?.desktop?.App?.ActorCancel?.(id);
        setLastResult(result);
        if (result?.success) loadActors();
      } catch (error) {
        console.error("Failed to cancel actor:", error);
      } finally {
        setIsLoading(false);
      }
    },
    [loadActors]
  );

  const handleCleanup = useCallback(async () => {
    try {
      const removed = await window.go?.desktop?.App?.ActorCleanup?.(300);
      setLastResult({
        success: true,
        message: `Cleaned up ${removed} actors`,
      });
      loadActors();
    } catch (error) {
      console.error("Failed to cleanup actors:", error);
    }
  }, [loadActors]);

  const formatTime = (ts: number) => {
    return new Date(ts * 1000).toLocaleTimeString();
  };

  return (
    <div className="p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium flex items-center gap-2">
          <Bot className="w-4 h-4" />
          子智能体
        </h3>
        <div className="flex gap-2">
          <button
            onClick={loadActors}
            className="p-1.5 rounded hover:bg-elevated text-txt-m"
            title="刷新"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
          <button
            onClick={handleCleanup}
            className="p-1.5 rounded hover:bg-elevated text-txt-m"
            title="清理"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Spawn form */}
      <div className="space-y-2">
        <select
          value={actorType}
          onChange={(e) => setActorType(e.target.value)}
          className="w-full px-3 py-2 text-sm text-txt bg-elevated border border-bdr rounded-md focus:outline-none focus:border-accent/50"
        >
          {ACTOR_TYPES.map((type) => (
            <option key={type.value} value={type.value}>
              {type.label}
            </option>
          ))}
        </select>
        <textarea
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          placeholder="输入提示词..."
          className="w-full px-3 py-2 text-sm text-txt placeholder:text-txt-m bg-elevated border border-bdr rounded-md resize-none focus:outline-none focus:border-accent/50"
          rows={3}
        />
        <button
          onClick={handleSpawn}
          disabled={!prompt.trim() || isLoading}
          className="w-full flex items-center justify-center gap-2 px-3 py-1.5 text-sm bg-accent text-white rounded disabled:opacity-50"
        >
          <Play className="w-4 h-4" />
          启动子智能体
        </button>
      </div>

      {/* Last result */}
      {lastResult && (
        <div
          className={`p-2 text-sm rounded ${
            lastResult.success
              ? "bg-green-500/10 text-green-400"
              : "bg-red-500/10 text-red-400"
          }`}
        >
          {lastResult.message}
        </div>
      )}

      {/* Actor list */}
      <div className="space-y-2">
        <h4 className="text-xs font-medium text-txt-m">
          活跃智能体 ({actors.length})
        </h4>
        {actors.length === 0 ? (
          <p className="text-xs text-txt-m">暂无子智能体</p>
        ) : (
          actors.map((actor) => {
            const StatusIcon = STATUS_ICONS[actor.status] || Bot;
            const isSpinning = actor.status === "running" || actor.status === "pending";

            return (
              <div
                key={actor.id}
                className="p-3 bg-elevated border border-bdr rounded space-y-2"
              >
                <div className="flex items-start gap-2">
                  <StatusIcon
                    className={`w-4 h-4 mt-0.5 ${STATUS_COLORS[actor.status]} ${
                      isSpinning ? "animate-spin" : ""
                    }`}
                  />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-medium px-1.5 py-0.5 bg-accent/20 text-accent rounded">
                        {actor.type}
                      </span>
                      <span className="text-xs text-txt-m">{actor.id}</span>
                    </div>
                    <p className="text-sm mt-1 line-clamp-2">{actor.prompt}</p>
                  </div>
                </div>

                {/* Result or Error */}
                {actor.result && (
                  <div className="ml-6 p-2 bg-green-500/5 border border-green-500/20 rounded text-xs text-green-400">
                    {actor.result}
                  </div>
                )}
                {actor.error && (
                  <div className="ml-6 p-2 bg-red-500/5 border border-red-500/20 rounded text-xs text-red-400">
                    {actor.error}
                  </div>
                )}

                {/* Timing */}
                <div className="ml-6 flex items-center gap-2 text-xs text-txt-m">
                  <span>创建: {formatTime(actor.created_at)}</span>
                  {actor.completed_at && (
                    <span>· 完成: {formatTime(actor.completed_at)}</span>
                  )}
                </div>

                {/* Actions */}
                {actor.status === "running" && (
                  <div className="ml-6">
                    <button
                      onClick={() => handleCancel(actor.id)}
                      disabled={isLoading}
                      className="px-2 py-0.5 text-xs bg-red-500/20 text-red-400 rounded hover:bg-red-500/30"
                    >
                      取消
                    </button>
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
