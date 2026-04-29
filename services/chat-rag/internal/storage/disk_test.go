package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Compile-time check: DiskStorage must satisfy StorageBackend.
var _ StorageBackend = (*DiskStorage)(nil)

func TestWrite_BasicContent(t *testing.T) {
	dir := t.TempDir()
	ds := NewDiskStorage(dir)

	key := "hello.txt"
	data := []byte("hello, disk storage")

	if _, err := ds.Write(key, data); err != nil {
		t.Fatalf("Write(%q) returned error: %v", key, err)
	}

	got, err := os.ReadFile(filepath.Join(dir, key))
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("content mismatch: got %q, want %q", got, data)
	}
}

func TestWrite_NestedKeyCreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	ds := NewDiskStorage(dir)

	key := "2026-04/03/user/file.json"
	data := []byte(`{"event":"test"}`)

	if _, err := ds.Write(key, data); err != nil {
		t.Fatalf("Write(%q) returned error: %v", key, err)
	}

	// Verify the intermediate directories were created.
	parentDir := filepath.Join(dir, "2026-04", "03", "user")
	info, err := os.Stat(parentDir)
	if err != nil {
		t.Fatalf("expected directory %q to exist: %v", parentDir, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", parentDir)
	}

	// Verify file content.
	got, err := os.ReadFile(filepath.Join(dir, key))
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("content mismatch: got %q, want %q", got, data)
	}
}

func TestWrite_OverwriteExistingFile(t *testing.T) {
	dir := t.TempDir()
	ds := NewDiskStorage(dir)

	key := "overwrite.txt"
	original := []byte("original content")
	updated := []byte("updated content")

	if _, err := ds.Write(key, original); err != nil {
		t.Fatalf("first Write(%q) returned error: %v", key, err)
	}
	if _, err := ds.Write(key, updated); err != nil {
		t.Fatalf("second Write(%q) returned error: %v", key, err)
	}

	got, err := os.ReadFile(filepath.Join(dir, key))
	if err != nil {
		t.Fatalf("failed to read overwritten file: %v", err)
	}
	if !bytes.Equal(got, updated) {
		t.Errorf("content mismatch after overwrite: got %q, want %q", got, updated)
	}
}

func TestClose_ReturnsNil(t *testing.T) {
	ds := NewDiskStorage(t.TempDir())

	if err := ds.Close(); err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}
}

func TestWrite_PathTraversalRejected(t *testing.T) {
	dir := t.TempDir()
	ds := NewDiskStorage(dir)

	traversalKeys := []string{
		"../../etc/passwd",
		"../outside",
		"foo/../../bar/../../../etc/shadow",
	}

	for _, key := range traversalKeys {
		if _, err := ds.Write(key, []byte("exploit")); err == nil {
			t.Errorf("Write(%q) should have returned an error for path traversal, but got nil", key)
		}
	}

	// Verify that a valid nested key still works.
	if _, err := ds.Write("safe/nested/file.txt", []byte("ok")); err != nil {
		t.Errorf("Write with safe key returned unexpected error: %v", err)
	}
}

func TestWrite_AbsoluteKeyRejected(t *testing.T) {
	dir := t.TempDir()
	ds := NewDiskStorage(dir)

	absKeys := []string{
		"/2026/04/log.json",
		"/etc/passwd",
		"/tmp/something",
	}

	for _, key := range absKeys {
		if _, err := ds.Write(key, []byte("data")); err == nil {
			t.Errorf("Write(%q) should have returned an error for absolute key, but got nil", key)
		}
	}
}

func TestWrite_SymlinkEscapeRejected(t *testing.T) {
	// Create two temp directories: one is the storage root, the other is "outside".
	storageDir := t.TempDir()
	outsideDir := t.TempDir()

	// Create a symlink inside storageDir that points to outsideDir.
	linkPath := filepath.Join(storageDir, "linked")
	if err := os.Symlink(outsideDir, linkPath); err != nil {
		t.Skipf("cannot create symlink (unsupported OS/permissions): %v", err)
	}

	ds := NewDiskStorage(storageDir)

	// Attempt to write through the symlink — should be rejected.
	_, err := ds.Write("linked/evil.json", []byte("escape attempt"))
	if err == nil {
		t.Fatal("Write through symlink pointing outside base path should have been rejected, but got nil")
	}

	// Verify the file was NOT written to the outside directory.
	if _, statErr := os.Stat(filepath.Join(outsideDir, "evil.json")); statErr == nil {
		t.Error("file was written to the outside directory despite the guard")
	}
}
