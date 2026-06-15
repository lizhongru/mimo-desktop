package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileNode represents a single file or directory in the workspace tree.
type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"isDir"`
	Children []FileNode `json:"children,omitempty"`
}

// ListWorkspaceFiles returns the directory tree rooted at `path` up to `maxDepth` levels.
// When `path` is empty, the current working directory is used.
// When `maxDepth` is 0 or negative, it defaults to 2.
func (a *App) ListWorkspaceFiles(path string, maxDepth int) ([]FileNode, error) {
	if path == "" {
		path, _ = os.Getwd()
	}
	if maxDepth <= 0 {
		maxDepth = 2
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("path not found: %s", path)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", path)
	}

	nodes, err := a.walkDir(path, path, 0, maxDepth)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// ListDirChildren returns the immediate children of a directory.
// This is used for lazy-loading: when the frontend expands a node it
// calls this to fetch one level at a time.
func (a *App) ListDirChildren(dirPath string) ([]FileNode, error) {
	if dirPath == "" {
		dirPath, _ = os.Getwd()
	}
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("path not found: %s", dirPath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", dirPath)
	}
	return a.walkDir(dirPath, dirPath, 0, 1)
}

// walkDir recursively walks the directory tree and returns FileNodes.
func (a *App) walkDir(root, dir string, depth, maxDepth int) ([]FileNode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		// Dirs first, then files, both alphabetical
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	var nodes []FileNode
	for _, e := range entries {
		name := e.Name()
		full := filepath.Join(dir, name)

		// Respect ignore rules
		if a.ignoreMatcher != nil && a.ignoreMatcher.ShouldIgnore(name, e.IsDir(), root, full) {
			continue
		}

		node := FileNode{
			Name:  name,
			Path:  full,
			IsDir: e.IsDir(),
		}

		// Recurse into dirs if within depth limit
		if e.IsDir() && depth+1 < maxDepth {
			children, err := a.walkDir(root, full, depth+1, maxDepth)
			if err == nil {
				node.Children = children
			}
		}

		nodes = append(nodes, node)
	}

	// Limit to 200 entries per directory to keep the UI responsive
	if len(nodes) > 200 {
		nodes = nodes[:200]
	}

	return nodes, nil
}
