import { useState, useCallback, useEffect } from "react";
import {
  Folder,
  FolderOpen,
  File,
  ChevronRight,
  ChevronDown,
  RefreshCw,
  Loader2,
} from "lucide-react";

interface FileNode {
  name: string;
  path: string;
  isDir: boolean;
  children?: FileNode[];
}

interface Props {
  root?: string;
  onFileClick?: (path: string) => void;
}

export function FileTree({ root, onFileClick }: Props) {
  const [rootNode, setRootNode] = useState<FileNode | null>(null);
  const [childrenMap, setChildrenMap] = useState<Map<string, FileNode[]>>(
    new Map(),
  );
  const [expandedDirs, setExpandedDirs] = useState<Set<string>>(new Set());
  const [loadingDirs, setLoadingDirs] = useState<Set<string>>(new Set());
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadRoot = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const workingDir =
        root || (await window.go?.desktop?.App?.GetWorkingDir?.()) || ".";
      const children: FileNode[] =
        (await window.go?.desktop?.App?.ListWorkspaceFiles?.(workingDir, 1)) ||
        [];
      setRootNode({
        name: workingDir.split(/[\\/]/).pop() || workingDir,
        path: workingDir,
        isDir: true,
      });
      setChildrenMap(new Map([[workingDir, children]]));
    } catch (err) {
      console.error("Failed to load file tree:", err);
      setError("Failed to load files");
    } finally {
      setIsLoading(false);
    }
  }, [root]);

  useEffect(() => {
    loadRoot();
  }, [loadRoot]);

  const toggleDir = useCallback(
    async (path: string) => {
      const willExpand = !expandedDirs.has(path);
      setExpandedDirs((prev) => {
        const next = new Set(prev);
        if (next.has(path)) {
          next.delete(path);
        } else {
          next.add(path);
        }
        return next;
      });

      if (willExpand && !childrenMap.has(path)) {
        setLoadingDirs((prev) => new Set(prev).add(path));
        try {
          const children: FileNode[] =
            (await window.go?.desktop?.App?.ListDirChildren?.(path)) || [];
          setChildrenMap((prev) => new Map(prev).set(path, children));
        } catch (err) {
          console.error("Failed to load directory children:", err);
        } finally {
          setLoadingDirs((prev) => {
            const next = new Set(prev);
            next.delete(path);
            return next;
          });
        }
      }
    },
    [childrenMap, expandedDirs],
  );

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 p-4 text-xs text-txt-m">
        <RefreshCw className="w-3 h-3 animate-spin" />
        <span>Loading...</span>
      </div>
    );
  }

  if (error) {
    return <div className="p-4 text-xs text-red-400">{error}</div>;
  }

  if (!rootNode) {
    return <div className="p-4 text-xs text-txt-m">No files</div>;
  }

  return (
    <div className="py-1">
      <TreeNode
        node={rootNode}
        childrenMap={childrenMap}
        expandedDirs={expandedDirs}
        loadingDirs={loadingDirs}
        onToggleDir={toggleDir}
        onFileClick={onFileClick}
        depth={0}
        isRoot
      />
    </div>
  );
}

export function FileTreeRefreshButton({ onClick }: { onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="p-1 rounded hover:bg-elevated text-txt-m cursor-pointer"
      title="Refresh"
    >
      <RefreshCw className="w-3 h-3" />
    </button>
  );
}

function TreeNode({
  node,
  childrenMap,
  expandedDirs,
  loadingDirs,
  onToggleDir,
  onFileClick,
  depth,
  isRoot,
}: {
  node: FileNode;
  childrenMap: Map<string, FileNode[]>;
  expandedDirs: Set<string>;
  loadingDirs: Set<string>;
  onToggleDir: (path: string) => void;
  onFileClick?: (path: string) => void;
  depth: number;
  isRoot?: boolean;
}) {
  const isExpanded = expandedDirs.has(node.path);
  const isLoading = loadingDirs.has(node.path);
  const children = childrenMap.get(node.path);

  const handleClick = () => {
    if (node.isDir) {
      onToggleDir(node.path);
    } else {
      onFileClick?.(node.path);
    }
  };

  return (
    <div>
      <div
        onClick={handleClick}
        className="flex items-center gap-1 px-2 py-[3px] text-xs hover:bg-elevated rounded cursor-pointer select-none group"
        style={{ paddingLeft: `${depth * 14 + 6}px` }}
      >
        {node.isDir && (
          <>
            {isLoading ? (
              <Loader2 className="w-3 h-3 text-txt-m animate-spin flex-shrink-0" />
            ) : isExpanded ? (
              <ChevronDown className="w-3 h-3 text-txt-m flex-shrink-0" />
            ) : (
              <ChevronRight className="w-3 h-3 text-txt-m flex-shrink-0" />
            )}
            {isExpanded ? (
              <FolderOpen className="w-3.5 h-3.5 text-accent flex-shrink-0" />
            ) : (
              <Folder className="w-3.5 h-3.5 text-accent flex-shrink-0" />
            )}
          </>
        )}
        {!node.isDir && (
          <File className="w-3.5 h-3.5 text-txt-m flex-shrink-0 ml-[12px]" />
        )}
        <span className="truncate text-txt-2 group-hover:text-txt">
          {node.name}
        </span>
      </div>

      {node.isDir && isExpanded && children && (
        <div>
          {children.map((child) => (
            <TreeNode
              key={child.path}
              node={child}
              childrenMap={childrenMap}
              expandedDirs={expandedDirs}
              loadingDirs={loadingDirs}
              onToggleDir={onToggleDir}
              onFileClick={onFileClick}
              depth={depth + 1}
            />
          ))}
          {children.length === 0 && (
            <div
              className="text-[11px] text-txt-m italic py-1"
              style={{ paddingLeft: `${(depth + 1) * 14 + 22}px` }}
            >
              Empty
            </div>
          )}
        </div>
      )}
    </div>
  );
}
