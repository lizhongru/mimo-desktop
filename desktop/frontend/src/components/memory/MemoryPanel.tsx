import { useState, useEffect, useCallback } from "react";
import { Database, FileText, RefreshCw, Plus, Trash2, Edit } from "lucide-react";
import { t } from "../../lib/i18n";

interface MemoryFileInfo {
  path: string;
  name: string;
  size: number;
  updatedAt: string;
  scope: string;
}

interface MemoryPanelProps {
  onClose?: () => void;
}

export function MemoryPanel({ onClose }: MemoryPanelProps) {
  const [files, setFiles] = useState<MemoryFileInfo[]>([]);
  const [selectedFile, setSelectedFile] = useState<MemoryFileInfo | null>(null);
  const [content, setContent] = useState("");
  const [isEditing, setIsEditing] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const loadFiles = useCallback(async () => {
    try {
      const fileList = await window.go?.desktop?.App?.ListMemoryFiles?.();
      setFiles(fileList || []);
    } catch (error) {
      console.error("Failed to load memory files:", error);
    }
  }, []);

  const loadFileContent = useCallback(async (file: MemoryFileInfo) => {
    setSelectedFile(file);
    setIsLoading(true);
    try {
      const fileContent = await window.go?.desktop?.App?.ReadMemory?.(file.scope);
      setContent(fileContent || "");
      setIsEditing(false);
    } catch (error) {
      console.error("Failed to load file content:", error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const saveContent = useCallback(async () => {
    if (!selectedFile) return;

    try {
      await window.go?.desktop?.App?.WriteMemory?.(selectedFile.scope, content);
      setIsEditing(false);
      loadFiles();
    } catch (error) {
      console.error("Failed to save file:", error);
    }
  }, [selectedFile, content, loadFiles]);

  const handleReconcile = useCallback(async () => {
    try {
      const [indexed, pruned] = await window.go?.desktop?.App?.MemoryReconcile?.();
      alert(`索引完成: ${indexed} 个文件已索引, ${pruned} 个已清理`);
      loadFiles();
    } catch (error) {
      console.error("Memory reconcile failed:", error);
    }
  }, [loadFiles]);

  useEffect(() => {
    loadFiles();
  }, [loadFiles]);

  return (
    <div className="flex h-full">
      {/* File List */}
      <div className="w-64 border-r border-bdr flex flex-col">
        <div className="flex items-center justify-between px-4 py-3 border-b border-bdr">
          <div className="flex items-center gap-2">
            <Database className="w-4 h-4 text-accent" />
            <span className="font-medium text-sm">记忆文件</span>
          </div>
          <button
            onClick={handleReconcile}
            className="p-1.5 rounded hover:bg-elevated transition-colors"
            title="同步索引"
          >
            <RefreshCw className="w-4 h-4 text-txt-m" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto">
          {files.length > 0 ? (
            <div className="p-2 space-y-1">
              {files.map((file) => (
                <button
                  key={file.path}
                  onClick={() => loadFileContent(file)}
                  className={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-left transition-colors ${
                    selectedFile?.path === file.path
                      ? "bg-accent/10 text-accent"
                      : "hover:bg-elevated text-txt"
                  }`}
                >
                  <FileText className="w-4 h-4 flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <div className="text-sm truncate">{file.name}</div>
                    <div className="text-xs text-txt-m truncate">{file.scope}</div>
                  </div>
                </button>
              ))}
            </div>
          ) : (
            <div className="p-4 text-center text-txt-m text-sm">
              暂无记忆文件
            </div>
          )}
        </div>
      </div>

      {/* File Content */}
      <div className="flex-1 flex flex-col">
        {selectedFile ? (
          <>
            <div className="flex items-center justify-between px-4 py-3 border-b border-bdr">
              <div className="flex items-center gap-2">
                <span className="font-medium text-sm">{selectedFile.name}</span>
                <span className="text-xs px-1.5 py-0.5 rounded bg-accent/20 text-accent">
                  {selectedFile.scope}
                </span>
              </div>
              <div className="flex items-center gap-2">
                {isEditing ? (
                  <>
                    <button
                      onClick={saveContent}
                      className="px-3 py-1.5 bg-accent text-white rounded text-sm hover:bg-accent-light"
                    >
                      保存
                    </button>
                    <button
                      onClick={() => setIsEditing(false)}
                      className="px-3 py-1.5 bg-elevated text-txt rounded text-sm hover:bg-elevated/80"
                    >
                      取消
                    </button>
                  </>
                ) : (
                  <button
                    onClick={() => setIsEditing(true)}
                    className="flex items-center gap-1.5 px-3 py-1.5 bg-elevated text-txt rounded text-sm hover:bg-elevated/80"
                  >
                    <Edit className="w-3.5 h-3.5" />
                    编辑
                  </button>
                )}
              </div>
            </div>

            <div className="flex-1 p-4 overflow-y-auto">
              {isLoading ? (
                <div className="text-center py-8 text-txt-m text-sm">加载中...</div>
              ) : isEditing ? (
                <textarea
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  className="w-full h-full min-h-[400px] p-4 bg-elevated border border-bdr rounded-lg text-sm font-mono focus:outline-none focus:border-accent resize-none"
                  placeholder="输入记忆内容..."
                />
              ) : (
                <div className="prose prose-sm max-w-none text-txt whitespace-pre-wrap">
                  {content || "空文件"}
                </div>
              )}
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-txt-m text-sm">
            选择一个文件查看内容
          </div>
        )}
      </div>
    </div>
  );
}
