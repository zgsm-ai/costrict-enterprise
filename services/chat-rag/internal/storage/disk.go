// Package storage defines abstractions for ChatLog persistence backends.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiskStorage implements StorageBackend by writing files to the local filesystem
// under a configured base path. It preserves the existing file write behavior:
// directories are auto-created with mode 0755 and files are written with mode 0644.
type DiskStorage struct {
	basePath string
}

// NewDiskStorage creates a DiskStorage that writes files under basePath.
// The basePath should be an absolute directory path (e.g., "/data/logs").
func NewDiskStorage(basePath string) *DiskStorage {
	return &DiskStorage{basePath: basePath}
}

// Write persists data to {basePath}/{key} on the local filesystem.
// Parent directories are created automatically with mode 0755.
// The file is written with mode 0644.
// The caller is responsible for content formatting (e.g., trailing newlines).
// Keys must be relative paths that resolve within basePath; absolute keys are
// rejected outright. Keys containing ".." segments that escape basePath, or
// keys that traverse symlinks pointing outside basePath, are also rejected.
func (d *DiskStorage) Write(key string, data []byte) (*WriteInfo, error) {
	// Reject absolute keys — callers must always pass relative paths.
	if filepath.IsAbs(key) {
		return nil, fmt.Errorf("disk storage: key %q must be a relative path", key)
	}

	fullPath := filepath.Join(d.basePath, key)

	// Fast pre-check: the cleaned path must stay within basePath.
	cleanBase := filepath.Clean(d.basePath) + string(filepath.Separator)
	cleanFull := filepath.Clean(fullPath)
	if !hasPrefix(cleanFull, cleanBase) && cleanFull != filepath.Clean(d.basePath) {
		return nil, fmt.Errorf("disk storage: key %q resolves outside base path", key)
	}

	// Create parent directories if they don't exist.
	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("disk storage: failed to create directory: %w", err)
	}

	// Resolve symlinks in the parent directory and re-check the prefix.
	// This must happen after MkdirAll because EvalSymlinks requires the path to exist.
	resolvedParent, err := filepath.EvalSymlinks(parentDir)
	if err != nil {
		return nil, fmt.Errorf("disk storage: failed to resolve symlinks: %w", err)
	}
	resolvedBase, err := filepath.EvalSymlinks(filepath.Clean(d.basePath))
	if err != nil {
		return nil, fmt.Errorf("disk storage: failed to resolve base path symlinks: %w", err)
	}
	resolvedBasePrefix := resolvedBase + string(filepath.Separator)
	if !hasPrefix(resolvedParent, resolvedBasePrefix) && resolvedParent != resolvedBase {
		return nil, fmt.Errorf("disk storage: key %q resolves outside base path after symlink resolution", key)
	}

	// Write file contents.
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return nil, fmt.Errorf("disk storage: failed to write file: %w", err)
	}

	return &WriteInfo{FilePath: key}, nil
}

// Close is a no-op for DiskStorage since there are no persistent resources
// (connections, handles, etc.) to release.
func (d *DiskStorage) Close() error {
	return nil
}

// hasPrefix is a filesystem-safe prefix check using cleaned paths.
func hasPrefix(path, prefix string) bool {
	return strings.HasPrefix(path, prefix)
}
