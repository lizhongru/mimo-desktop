import { useState, useCallback, useEffect } from "react";
import {
  Folder,
  FolderOpen,
  File,
  ChevronRight,
  ChevronDown,
  RefreshCw,
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

export function FileTree({ root = ".", onFileClick }: Props) {
  const [tree, setTree] = useState<FileNode | null>(null);
  const [expandedDirs, setExpandedDirs] = useState<Set<string>>(new Set());
  const [isLoading, setIsLoading] = useState(false);

  const loadTree = useCallback(async () => {
    setIsLoading(true);
    try {
      const workingDir = await window.go?.desktop?.App?.GetWorkingDir?.();
      // For now, use a simple mock tree
      // In real implementation, this would call a backend API
      const mockTree: FileNode = {
        name: workingDir || ".",
        path: workingDir || ".",
        isDir: true,
        children: [],
      };
      setTree(mockTree);
    } catch (error) {
      console.error("Failed to load file tree:", error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadTree();
  }, [loadTree]);

  const toggleDir = useCallback((path: string) => {
    setExpandedDirs((prev) => {
      const next = new Set(prev);
      if (next.has(path)) {
        next.delete(path);
      } else {
        next.add(path);
      }
      return next;
    });
  }, []);

  if (!tree) {
    return (
      <div className="p-4 text-sm text-txt-m">
        {isLoading ? "Loading..." : "No files"}
      </div>
    );
  }

  return (
    <div className="p-2">
      <div className="flex items-center justify-between mb-2 px-2">
        <span className="text-xs font-medium text-txt-m">Files</span>
        <button
          onClick={loadTree}
          className="p-1 rounded hover:bg-elevated text-txt-m"
          title="Refresh"
        >
          <RefreshCw className="w-3 h-3" />
        </button>
      </div>
      <TreeNode
        node={tree}
        expandedDirs={expandedDirs}
        onToggleDir={toggleDir}
        onFileClick={onFileClick}
        depth={0}
      />
    </div>
  );
}

function TreeNode({
  node,
  expandedDirs,
  onToggleDir,
  onFileClick,
  depth,
}: {
  node: FileNode;
  expandedDirs: Set<string>;
  onToggleDir: (path: string) => void;
  onFileClick?: (path: string) => void;
  depth: number;
}) {
  const isExpanded = expandedDirs.has(node.path);

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
        className="flex items-center gap-1 px-2 py-1 text-sm hover:bg-elevated rounded cursor-pointer"
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
      >
        {node.isDir && (
          <>
            {isExpanded ? (
              <ChevronDown className="w-3 h-3 text-txt-m" />
            ) : (
              <ChevronRight className="w-3 h-3 text-txt-m" />
            )}
            {isExpanded ? (
              <FolderOpen className="w-4 h-4 text-accent" />
            ) : (
              <Folder className="w-4 h-4 text-accent" />
            )}
          </>
        )}
        {!node.isDir && <File className="w-4 h-4 text-txt-m ml-4" />}
        <span className="truncate">{node.name}</span>
      </div>

      {node.isDir && isExpanded && node.children && (
        <div>
          {node.children.map((child) => (
            <TreeNode
              key={child.path}
              node={child}
              expandedDirs={expandedDirs}
              onToggleDir={onToggleDir}
              onFileClick={onFileClick}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  );
}
