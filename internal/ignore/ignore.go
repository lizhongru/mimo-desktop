package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// pattern represents a single ignore pattern
type pattern struct {
	raw     string
	negated bool
	dirOnly bool
	text    string // cleaned pattern text (without ! and trailing /)
}

// Matcher checks paths against ignore patterns from multiple sources
type Matcher struct {
	mu       sync.RWMutex
	patterns []pattern
}

// New creates a Matcher with built-in default patterns
func New() *Matcher {
	m := &Matcher{}
	defaults := []string{
		".git",
		".mimo",
		"node_modules",
		"__pycache__",
		".venv",
		".idea",
		".vscode",
		"vendor",
		"dist",
		"build",
		"bin",
	}
	m.AddPatterns(defaults)
	return m
}

// AddPatterns adds patterns from a slice of strings
func (m *Matcher) AddPatterns(lines []string) {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m.addOne(line)
	}
}

// addOne adds a single pattern (internal, no lock)
func (m *Matcher) addOne(raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "#") {
		return
	}

	p := pattern{raw: raw}

	if strings.HasPrefix(raw, "!") {
		p.negated = true
		raw = raw[1:]
	}
	if strings.HasSuffix(raw, "/") {
		p.dirOnly = true
		raw = strings.TrimSuffix(raw, "/")
	}
	p.text = raw

	m.mu.Lock()
	m.patterns = append(m.patterns, p)
	m.mu.Unlock()
}

// LoadFile loads patterns from a .mimoignore file
func (m *Matcher) LoadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	m.AddPatterns(lines)
	return scanner.Err()
}

// ShouldIgnore checks if a path should be ignored.
// name = basename, isDir = whether it's a directory.
// rootDir = project root, fullPath = absolute path of the entry.
func (m *Matcher) ShouldIgnore(name string, isDir bool, rootDir, fullPath string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	relPath, err := filepath.Rel(rootDir, fullPath)
	if err != nil {
		relPath = name
	}
	relPath = filepath.ToSlash(relPath)

	ignored := false
	for _, p := range m.patterns {
		if p.dirOnly && !isDir {
			continue
		}
		if matchPattern(p.text, name, relPath) {
			if p.negated {
				ignored = false
			} else {
				ignored = true
			}
		}
	}
	return ignored
}

// ShouldIgnoreName is a simplified version when only the name is available.
func (m *Matcher) ShouldIgnoreName(name string, isDir bool) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ignored := false
	for _, p := range m.patterns {
		if p.dirOnly && !isDir {
			continue
		}
		if matchPattern(p.text, name, name) {
			if p.negated {
				ignored = false
			} else {
				ignored = true
			}
		}
	}
	return ignored
}

// matchPattern matches a single cleaned pattern against name and relPath.
// Gitignore rules:
//   - Pattern without '/' matches against basename only (at any depth)
//   - Pattern with '/' matches against the full relative path
//   - Leading '/' anchors to the root
func matchPattern(pat, name, relPath string) bool {
	// Strip leading / (root-anchored)
	pat = strings.TrimPrefix(pat, "/")

	if strings.Contains(pat, "/") {
		// Pattern has slash → match against full relative path
		return simpleGlob(pat, relPath)
	}
	// No slash → match against basename only
	return simpleGlob(pat, name)
}

// simpleGlob does glob matching where * does NOT cross path separators.
// Supports * and ? only (no **).
func simpleGlob(pattern, name string) bool {
	matched, err := filepath.Match(pattern, name)
	return err == nil && matched
}

// Patterns returns all current patterns (for debugging)
func (m *Matcher) Patterns() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.patterns))
	for i, p := range m.patterns {
		result[i] = p.raw
	}
	return result
}
