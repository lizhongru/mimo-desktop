import { useState, useCallback, useEffect } from "react";
import {
  ListTodo,
  Plus,
  Play,
  CheckCircle2,
  XCircle,
  Trash2,
  ChevronDown,
  ChevronRight,
  Clock,
  AlertCircle,
  Archive,
} from "lucide-react";

import { t } from "../../lib/i18n";

interface TaskInfo {
  id: string;
  session_id: string;
  parent_task_id?: string;
  status: string;
  summary: string;
  owner?: string;
  created_at: number;
  last_event_at: number;
  ended_at?: number;
}

interface TaskEventInfo {
  id: number;
  task_id: string;
  at: number;
  kind: string;
  summary?: string;
}

interface TaskResult {
  success: boolean;
  message: string;
  task?: TaskInfo;
}

const STATUS_COLORS: Record<string, string> = {
  open: "text-txt-m",
  in_progress: "text-blue-400",
  blocked: "text-yellow-400",
  done: "text-green-400",
  abandoned: "text-red-400",
  archived: "text-gray-500",
};

const STATUS_ICONS: Record<string, typeof ListTodo> = {
  open: ListTodo,
  in_progress: Play,
  blocked: AlertCircle,
  done: CheckCircle2,
  abandoned: XCircle,
  archived: Archive,
};

export function TaskPanel() {
  const [tasks, setTasks] = useState<TaskInfo[]>([]);
  const [summary, setSummary] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [lastResult, setLastResult] = useState<TaskResult | null>(null);
  const [expandedTasks, setExpandedTasks] = useState<Set<string>>(new Set());
  const [taskEvents, setTaskEvents] = useState<Record<string, TaskEventInfo[]>>(
    {}
  );
  const [renamingTaskId, setRenamingTaskId] = useState<string | null>(null);
  const [renameValue, setRenameValue] = useState("");
  const [progressTaskId, setProgressTaskId] = useState<string | null>(null);
  const [progressValue, setProgressValue] = useState("");

  const loadTasks = useCallback(async () => {
    try {
      const list = await window.go?.desktop?.App?.TaskList?.("", true);
      setTasks(list || []);
    } catch (error) {
      console.error("Failed to load tasks:", error);
    }
  }, []);

  useEffect(() => {
    loadTasks();
  }, [loadTasks]);

  const handleCreateTask = useCallback(async () => {
    if (!summary.trim()) return;

    setIsLoading(true);
    try {
      const result = await window.go?.desktop?.App?.TaskCreate?.(summary, "");
      setLastResult(result);
      if (result?.success) {
        setSummary("");
        loadTasks();
      }
    } catch (error) {
      console.error("Failed to create task:", error);
      setLastResult({ success: false, message: "Failed to create task" });
    } finally {
      setIsLoading(false);
    }
  }, [summary, loadTasks]);

  const handleStartTask = useCallback(
    async (id: string) => {
      setIsLoading(true);
      try {
        const result = await window.go?.desktop?.App?.TaskStart?.(id, "user", "Task started");
        setLastResult(result);
        if (result?.success) loadTasks();
      } catch (error) {
        console.error("Failed to start task:", error);
      } finally {
        setIsLoading(false);
      }
    },
    [loadTasks]
  );

  const handleDoneTask = useCallback(
    async (id: string) => {
      setIsLoading(true);
      try {
        const result = await window.go?.desktop?.App?.TaskDone?.(id, "Task completed");
        setLastResult(result);
        if (result?.success) loadTasks();
      } catch (error) {
        console.error("Failed to complete task:", error);
      } finally {
        setIsLoading(false);
      }
    },
    [loadTasks]
  );

  const handleBlockTask = useCallback(
    async (id: string) => {
      setIsLoading(true);
      try {
        const result = await window.go?.desktop?.App?.TaskBlock?.(id, "Task blocked");
        setLastResult(result);
        if (result?.success) loadTasks();
      } catch (error) {
        console.error("Failed to block task:", error);
      } finally {
        setIsLoading(false);
      }
    },
    [loadTasks]
  );

  const handleDeleteTask = useCallback(
    async (id: string) => {
      if (!confirm("确定要删除此任务吗？")) return;

      setIsLoading(true);
      try {
        const result = await window.go?.desktop?.App?.TaskDelete?.(id);
        setLastResult(result);
        if (result?.success) loadTasks();
      } catch (error) {
        console.error("Failed to delete task:", error);
      } finally {
        setIsLoading(false);
      }
    },
    [loadTasks]
  );

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

  const toggleExpand = useCallback(async (taskId: string) => {
    setExpandedTasks((prev) => {
      const next = new Set(prev);
      if (next.has(taskId)) {
        next.delete(taskId);
      } else {
        next.add(taskId);
      }
      return next;
    });
  }, []);

  const formatTime = (ts: number) => {
    return new Date(ts * 1000).toLocaleString();
  };

  const rootTasks = tasks.filter((t) => !t.parent_task_id);
  const childTasks = tasks.filter((t) => t.parent_task_id);

  return (
    <div className="p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium flex items-center gap-2">
          <ListTodo className="w-4 h-4" />
          任务管理
        </h3>
        <button
          onClick={loadTasks}
          className="p-1.5 rounded hover:bg-elevated text-txt-m text-xs"
        >
          刷新
        </button>
      </div>

      {/* Create task */}
      <div className="space-y-2">
        <input
          value={summary}
          onChange={(e) => setSummary(e.target.value)}
          placeholder="输入任务描述..."
          className="w-full px-3 py-2 text-sm text-txt placeholder:text-txt-m bg-elevated border border-bdr rounded-md focus:outline-none focus:border-accent/50"
          onKeyDown={(e) => e.key === "Enter" && handleCreateTask()}
        />
        <button
          onClick={handleCreateTask}
          disabled={!summary.trim() || isLoading}
          className="w-full flex items-center justify-center gap-2 px-3 py-1.5 text-sm bg-accent text-white rounded disabled:opacity-50"
        >
          <Plus className="w-4 h-4" />
          创建任务
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

      {/* Task list */}
      <div className="space-y-2">
        <h4 className="text-xs font-medium text-txt-m">
          任务列表 ({tasks.length})
        </h4>
        {rootTasks.length === 0 ? (
          <p className="text-xs text-txt-m">暂无任务</p>
        ) : (
          rootTasks.map((task) => {
            const StatusIcon = STATUS_ICONS[task.status] || ListTodo;
            const isExpanded = expandedTasks.has(task.id);
            const children = childTasks.filter(
              (t) => t.parent_task_id === task.id
            );

            return (
              <div key={task.id} className="bg-elevated border border-bdr rounded">
                {/* Task header */}
                <div className="p-3 space-y-2">
                  <div className="flex items-start gap-2">
                    {children.length > 0 && (
                      <button
                        onClick={() => toggleExpand(task.id)}
                        className="mt-0.5 text-txt-m hover:text-txt-2"
                      >
                        {isExpanded ? (
                          <ChevronDown className="w-3 h-3" />
                        ) : (
                          <ChevronRight className="w-3 h-3" />
                        )}
                      </button>
                    )}
                    <StatusIcon
                      className={`w-4 h-4 mt-0.5 ${STATUS_COLORS[task.status]}`}
                    />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm">{task.summary}</p>
                      <div className="flex items-center gap-2 text-xs text-txt-m mt-1">
                        <Clock className="w-3 h-3" />
                        <span>{formatTime(task.created_at)}</span>
                        {task.owner && <span>· {task.owner}</span>}
                      </div>
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="flex gap-1 ml-6">
                    {task.status === "open" && (
                      <button
                        onClick={() => handleStartTask(task.id)}
                        disabled={isLoading}
                        className="px-2 py-0.5 text-xs bg-blue-500/20 text-blue-400 rounded hover:bg-blue-500/30"
                      >
                        开始
                      </button>
                    )}
                    {(task.status === "in_progress" ||
                      task.status === "blocked") && (
                      <button
                        onClick={() => handleDoneTask(task.id)}
                        disabled={isLoading}
                        className="px-2 py-0.5 text-xs bg-green-500/20 text-green-400 rounded hover:bg-green-500/30"
                      >
                        完成
                      </button>
                    )}
                    {task.status === "in_progress" && (
                      <button
                        onClick={() => handleBlockTask(task.id)}
                        disabled={isLoading}
                        className="px-2 py-0.5 text-xs bg-yellow-500/20 text-yellow-400 rounded hover:bg-yellow-500/30"
                      >
                        阻塞
                      </button>
                    )}
                    {/* Rename */}
                    {task.status !== "done" && task.status !== "abandoned" && task.status !== "archived" && (
                      <button
                        onClick={() => { setRenamingTaskId(task.id); setRenameValue(task.summary); setProgressTaskId(null); }}
                        disabled={isLoading}
                        className="px-2 py-0.5 text-xs bg-purple-500/20 text-purple-400 rounded hover:bg-purple-500/30"
                      >
                        {t("task_rename")}
                      </button>
                    )}
                    {/* Progress */}
                    {task.status === "in_progress" && (
                      <button
                        onClick={() => { setProgressTaskId(task.id); setProgressValue(""); setRenamingTaskId(null); }}
                        disabled={isLoading}
                        className="px-2 py-0.5 text-xs bg-cyan-500/20 text-cyan-400 rounded hover:bg-cyan-500/30"
                      >
                        {t("task_progress")}
                      </button>
                    )}
                    {/* Archive */}
                    {(task.status === "done" || task.status === "abandoned" || task.status === "blocked") && (
                      <button
                        onClick={() => handleArchiveTask(task.id)}
                        disabled={isLoading}
                        className="px-2 py-0.5 text-xs bg-gray-500/20 text-gray-400 rounded hover:bg-gray-500/30"
                      >
                        {t("task_archive")}
                      </button>
                    )}
                    <button
                      onClick={() => handleDeleteTask(task.id)}
                      disabled={isLoading}
                      className="px-2 py-0.5 text-xs bg-red-500/20 text-red-400 rounded hover:bg-red-500/30"
                    >
                      删除
                    </button>
                  </div>
                </div>

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

                {/* Children */}
                {isExpanded && children.length > 0 && (
                  <div className="border-t border-bdr px-4 py-2 space-y-2">
                    {children.map((child) => {
                      const ChildIcon = STATUS_ICONS[child.status] || ListTodo;
                      return (
                        <div
                          key={child.id}
                          className="flex items-center gap-2 text-sm"
                        >
                          <ChildIcon
                            className={`w-3 h-3 ${STATUS_COLORS[child.status]}`}
                          />
                          <span className="flex-1">{child.summary}</span>
                          <span className="text-xs text-txt-m">
                            {formatTime(child.created_at)}
                          </span>
                        </div>
                      );
                    })}
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
