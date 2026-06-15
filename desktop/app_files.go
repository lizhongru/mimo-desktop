package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"encoding/base64"
	"unicode/utf8"
)

// FileNode represents a single file or directory in the workspace tree.
type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"isDir"`
	Children []FileNode `json:"children,omitempty"`
}

// FilePreview represents file metadata and preview content for the frontend.
type FilePreview struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	IsDir     bool   `json:"isDir"`
	SizeBytes int64  `json:"sizeBytes"`
	IsText    bool   `json:"isText"`
	IsImage   bool   `json:"isImage"`
	Truncated bool   `json:"truncated"`
	Content   string `json:"content"`
	Language  string `json:"language"`
	Mime      string `json:"mime"`
}

const maxPreviewBytes = 256 * 1024

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

// ReadFilePreview reads a file for preview purposes in the UI.
func (a *App) ReadFilePreview(path string) (*FilePreview, error) {
	if path == "" {
		return nil, fmt.Errorf("path is required")
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("path not found: %s", path)
	}

	preview := &FilePreview{
		Name:      info.Name(),
		Path:      path,
		IsDir:     info.IsDir(),
		SizeBytes: info.Size(),
	}

	if info.IsDir() {
		preview.IsText = true
		preview.Language = "text"
		return preview, nil
	}

	// Handle image files — return base64
	if isImageExtension(path) {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read image: %w", err)
		}
		preview.IsImage = true
		preview.IsText = false
		preview.Content = base64.StdEncoding.EncodeToString(raw)
		preview.Mime = mimeByExtension(path)
		preview.Language = "image"
		return preview, nil
	}

	if info.Size() == 0 {
		preview.IsText = true
		preview.Language = languageByExtension(path)
		return preview, nil
	}

	readBytes := info.Size()
	truncated := false
	if readBytes > maxPreviewBytes {
		readBytes = maxPreviewBytes
		truncated = true
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if int64(len(raw)) > readBytes {
		raw = raw[:readBytes]
	}

	if !utf8.Valid(raw) {
		preview.IsText = false
		preview.Content = ""
		preview.Language = "binary"
		return preview, nil
	}

	preview.IsText = true
	preview.Truncated = truncated
	preview.Content = string(raw)
	preview.Language = languageByExtension(path)

	return preview, nil
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

func isImageExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".ico", ".svg":
		return true
	}
	return false
}

func mimeByExtension(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".ico":
		return "image/x-icon"
	case ".svg":
		return "image/svg+xml"
	}
	return "application/octet-stream"
}

// languageByExtension returns a lightweight language hint for syntax highlighting.
func languageByExtension(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".js", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx", ".mts", ".cts":
		return "typescript"
	case ".jsx":
		return "jsx"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp", ".hh", ".hxx":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".xml", ".svg":
		return "xml"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".scss", ".sass":
		return "scss"
	case ".less":
		return "less"
	case ".md", ".markdown":
		return "markdown"
	case ".sql":
		return "sql"
	case ".sh", ".bash", ".zsh":
		return "bash"
	case ".ps1", ".psm1":
		return "powershell"
	case ".bat", ".cmd":
		return "bat"
	case ".ini", ".cfg", ".conf":
		return "ini"
	case ".env":
		return "dotenv"
	case ".dockerfile":
		return "dockerfile"
	case ".graphql", ".gql":
		return "graphql"
	case ".proto":
		return "protobuf"
	case ".lua":
		return "lua"
	case ".r":
		return "r"
	case ".dart":
		return "dart"
	case ".vue":
		return "vue"
	case ".svelte":
		return "svelte"
	}
	if strings.EqualFold(filepath.Base(path), "Dockerfile") {
		return "dockerfile"
	}
	if strings.EqualFold(filepath.Base(path), "Makefile") {
		return "makefile"
	}
	return "text"
}
