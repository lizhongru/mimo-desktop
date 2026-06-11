package tools

import (
	"context"
	"os"
	"path/filepath"
)

type workingDirContextKey struct{}

// WithWorkingDir attaches the selected workspace directory to a tool context.
func WithWorkingDir(ctx context.Context, dir string) context.Context {
	if dir == "" {
		return ctx
	}
	return context.WithValue(ctx, workingDirContextKey{}, dir)
}

// WorkingDir returns the selected workspace directory, falling back to process cwd.
func WorkingDir(ctx context.Context) string {
	if dir, ok := ctx.Value(workingDirContextKey{}).(string); ok && dir != "" {
		return dir
	}
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// ResolvePath resolves a relative path against the selected workspace directory.
func ResolvePath(ctx context.Context, path string) string {
	if path == "" {
		path = "."
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(WorkingDir(ctx), path)
}
