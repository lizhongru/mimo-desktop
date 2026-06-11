import { useState, useCallback } from "react";
import { History, Plus, Trash2, RotateCcw, Download } from "lucide-react";

interface CheckpointInfo {
  id: string;
  summary: string;
  token_count: number;
  message_offset: number;
  created_at: string;
}

interface CheckpointResult {
  success: boolean;
  message: string;
  id?: string;
}

export function CheckpointPanel() {
  const [checkpoints, setCheckpoints] = useState<CheckpointInfo[]>([]);
  const [summary, setSummary] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [lastResult, setLastResult] = useState<CheckpointResult | null>(null);

  const loadCheckpoints = useCallback(async () => {
    try {
      const list = await window.go?.desktop?.App?.ListCheckpoints?.();
      setCheckpoints(list || []);
    } catch (error) {
      console.error("Failed to load checkpoints:", error);
    }
  }, []);

  const handleCreateCheckpoint = useCallback(async () => {
    if (!summary.trim()) return;

    setIsLoading(true);
    try {
      const result = await window.go?.desktop?.App?.CreateCheckpoint?.(summary);
      setLastResult(result);
      if (result?.success) {
        setSummary("");
        loadCheckpoints();
      }
    } catch (error) {
      console.error("Failed to create checkpoint:", error);
      setLastResult({ success: false, message: "Failed to create checkpoint" });
    } finally {
      setIsLoading(false);
    }
  }, [summary, loadCheckpoints]);

  const handleRestoreCheckpoint = useCallback(async (id: string) => {
    setIsLoading(true);
    try {
      const result = await window.go?.desktop?.App?.RestoreCheckpoint?.(id);
      setLastResult(result);
    } catch (error) {
      console.error("Failed to restore checkpoint:", error);
      setLastResult({ success: false, message: "Failed to restore checkpoint" });
    } finally {
      setIsLoading(false);
    }
  }, []);

  const handleDeleteCheckpoint = useCallback(async (id: string) => {
    if (!confirm("确定要删除此检查点吗？")) return;

    setIsLoading(true);
    try {
      const result = await window.go?.desktop?.App?.DeleteCheckpoint?.(id);
      setLastResult(result);
      if (result?.success) {
        loadCheckpoints();
      }
    } catch (error) {
      console.error("Failed to delete checkpoint:", error);
      setLastResult({ success: false, message: "Failed to delete checkpoint" });
    } finally {
      setIsLoading(false);
    }
  }, [loadCheckpoints]);

  const handleExport = useCallback(async () => {
    try {
      const json = await window.go?.desktop?.App?.ExportCheckpoints?.();
      const blob = new Blob([json], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `checkpoints-${new Date().toISOString().split("T")[0]}.json`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error("Failed to export checkpoints:", error);
    }
  }, []);

  return (
    <div className="p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium flex items-center gap-2">
          <History className="w-4 h-4" />
          检查点管理
        </h3>
        <div className="flex gap-2">
          <button
            onClick={loadCheckpoints}
            className="p-1.5 rounded hover:bg-elevated text-txt-m"
            title="刷新"
          >
            <RotateCcw className="w-4 h-4" />
          </button>
          <button
            onClick={handleExport}
            className="p-1.5 rounded hover:bg-elevated text-txt-m"
            title="导出"
          >
            <Download className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Create checkpoint */}
      <div className="space-y-2">
        <textarea
          value={summary}
          onChange={(e) => setSummary(e.target.value)}
          placeholder="输入检查点摘要..."
          className="w-full px-3 py-2 text-sm text-txt placeholder:text-txt-m bg-elevated border border-bdr rounded-md resize-none focus:outline-none focus:border-accent/50"
          rows={3}
        />
        <button
          onClick={handleCreateCheckpoint}
          disabled={!summary.trim() || isLoading}
          className="w-full flex items-center justify-center gap-2 px-3 py-1.5 text-sm bg-accent text-white rounded disabled:opacity-50"
        >
          <Plus className="w-4 h-4" />
          创建检查点
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

      {/* Checkpoint list */}
      <div className="space-y-2">
        <h4 className="text-xs font-medium text-txt-m">检查点列表</h4>
        {checkpoints.length === 0 ? (
          <p className="text-xs text-txt-m">暂无检查点</p>
        ) : (
          checkpoints.map((cp) => (
            <div
              key={cp.id}
              className="p-3 bg-elevated border border-bdr rounded space-y-2"
            >
              <p className="text-sm line-clamp-2">{cp.summary}</p>
              <div className="flex items-center justify-between text-xs text-txt-m">
                <span>
                  {new Date(cp.created_at).toLocaleString()} · {cp.token_count} tokens
                </span>
                <div className="flex gap-1">
                  <button
                    onClick={() => handleRestoreCheckpoint(cp.id)}
                    disabled={isLoading}
                    className="p-1 rounded hover:bg-accent/20 text-accent"
                    title="恢复"
                  >
                    <RotateCcw className="w-3 h-3" />
                  </button>
                  <button
                    onClick={() => handleDeleteCheckpoint(cp.id)}
                    disabled={isLoading}
                    className="p-1 rounded hover:bg-red-500/20 text-red-400"
                    title="删除"
                  >
                    <Trash2 className="w-3 h-3" />
                  </button>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
