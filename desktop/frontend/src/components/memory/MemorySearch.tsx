import { useState, useCallback } from "react";
import { Search, FileText, Database, Trash2, RefreshCw } from "lucide-react";
import { t } from "../../lib/i18n";

interface MemoryResult {
  path: string;
  snippet: string;
  score: number;
  scope: string;
  scope_id: string;
  type: string;
}

interface MemoryFileInfo {
  path: string;
  name: string;
  size: number;
  updatedAt: string;
  scope: string;
}

export function MemorySearch() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<MemoryResult[]>([]);
  const [files, setFiles] = useState<MemoryFileInfo[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [activeTab, setActiveTab] = useState<"search" | "files">("search");

  const handleSearch = useCallback(async () => {
    if (!query.trim()) return;

    setIsSearching(true);
    try {
      const searchResults = await window.go?.desktop?.App?.MemorySearch?.(query, "", 10);
      setResults(searchResults || []);
    } catch (error) {
      console.error("Memory search failed:", error);
    } finally {
      setIsSearching(false);
    }
  }, [query]);

  const handleLoadFiles = useCallback(async () => {
    try {
      const fileList = await window.go?.desktop?.App?.ListMemoryFiles?.();
      setFiles(fileList || []);
    } catch (error) {
      console.error("Failed to load memory files:", error);
    }
  }, []);

  const handleReconcile = useCallback(async () => {
    try {
      const [indexed, pruned] = await window.go?.desktop?.App?.MemoryReconcile?.();
      alert(`索引完成: ${indexed} 个文件已索引, ${pruned} 个已清理`);
      if (activeTab === "files") {
        handleLoadFiles();
      }
    } catch (error) {
      console.error("Memory reconcile failed:", error);
    }
  }, [activeTab, handleLoadFiles]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSearch();
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-bdr">
        <div className="flex items-center gap-2">
          <Database className="w-4 h-4 text-accent" />
          <span className="font-medium text-sm">记忆系统</span>
        </div>
        <button
          onClick={handleReconcile}
          className="p-1.5 rounded hover:bg-elevated transition-colors"
          title="同步索引"
        >
          <RefreshCw className="w-4 h-4 text-txt-m" />
        </button>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-bdr">
        <button
          className={`flex-1 px-4 py-2 text-sm transition-colors ${
            activeTab === "search"
              ? "text-accent border-b-2 border-accent"
              : "text-txt-m hover:text-txt"
          }`}
          onClick={() => setActiveTab("search")}
        >
          搜索
        </button>
        <button
          className={`flex-1 px-4 py-2 text-sm transition-colors ${
            activeTab === "files"
              ? "text-accent border-b-2 border-accent"
              : "text-txt-m hover:text-txt"
          }`}
          onClick={() => {
            setActiveTab("files");
            handleLoadFiles();
          }}
        >
          文件
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto">
        {activeTab === "search" ? (
          <div className="p-4">
            {/* Search Input */}
            <div className="flex gap-2 mb-4">
              <div className="flex-1 relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-txt-m" />
                <input
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="搜索记忆..."
                  className="w-full pl-10 pr-4 py-2 bg-elevated border border-bdr rounded-lg text-sm focus:outline-none focus:border-accent"
                />
              </div>
              <button
                onClick={handleSearch}
                disabled={isSearching || !query.trim()}
                className="px-4 py-2 bg-accent text-white rounded-lg text-sm hover:bg-accent-light disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isSearching ? "搜索中..." : "搜索"}
              </button>
            </div>

            {/* Results */}
            {results.length > 0 ? (
              <div className="space-y-3">
                {results.map((result, index) => (
                  <div
                    key={`${result.path}-${index}`}
                    className="p-3 bg-elevated rounded-lg border border-bdr"
                  >
                    <div className="flex items-center gap-2 mb-2">
                      <FileText className="w-4 h-4 text-txt-m" />
                      <span className="text-xs text-txt-m truncate">{result.path}</span>
                      <span className="text-xs px-1.5 py-0.5 rounded bg-accent/20 text-accent">
                        {result.scope}
                      </span>
                      {result.type && (
                        <span className="text-xs px-1.5 py-0.5 rounded bg-blue-500/20 text-blue-400">
                          {result.type}
                        </span>
                      )}
                    </div>
                    <div
                      className="text-sm text-txt"
                      dangerouslySetInnerHTML={{ __html: result.snippet }}
                    />
                    <div className="mt-2 text-xs text-txt-g">
                      分数: {result.score.toFixed(3)}
                    </div>
                  </div>
                ))}
              </div>
            ) : query && !isSearching ? (
              <div className="text-center py-8 text-txt-m text-sm">
                未找到相关记忆
              </div>
            ) : (
              <div className="text-center py-8 text-txt-m text-sm">
                输入关键词搜索项目记忆
              </div>
            )}
          </div>
        ) : (
          <div className="p-4">
            {files.length > 0 ? (
              <div className="space-y-2">
                {files.map((file) => (
                  <div
                    key={file.path}
                    className="flex items-center gap-3 p-3 bg-elevated rounded-lg border border-bdr"
                  >
                    <FileText className="w-4 h-4 text-txt-m flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <div className="text-sm text-txt truncate">{file.name}</div>
                      <div className="text-xs text-txt-m truncate">{file.path}</div>
                    </div>
                    <div className="text-xs text-txt-g">
                      {new Date(file.updatedAt).toLocaleDateString()}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8 text-txt-m text-sm">
                暂无记忆文件
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
