// Package storage defines abstractions for ChatLog persistence backends.
package storage

// WriteInfo holds metadata returned by a successful Write operation.
// Fields are backend-specific; callers should check for non-empty/non-nil
// values before using them.
type WriteInfo struct {
	// FilePath is the fully-resolved path (or object key) where the data was
	// FilePath eq key
	// persisted.
	//   - DiskStorage: absolute filesystem path, e.g. "/data/logs/2026/04/03/chat-abc123.json"
	//   - S3Storage:   the object key used, e.g. "2026/04/03/chat-abc123.json"
	FilePath string

	// ETag is the entity tag (typically an MD5 hex digest of the uploaded
	// content) returned by S3-compatible backends. Empty for DiskStorage.
	ETag string
}

// StorageBackend abstracts the underlying storage mechanism for ChatLog persistence.
// All ChatLog write operations must go through this interface, allowing the system
// to swap between different storage backends (e.g., local filesystem, S3) without
// changing the caller logic.
type StorageBackend interface {
	// Write persists data under the given key.
	// The key is a relative object path (e.g., "2026/04/03/chat-abc123.json")
	// whose semantics are unified across backends — it maps to a file path for
	// local storage or an object key for S3-compatible storage.
	// On success it returns a WriteInfo populated with backend-specific metadata.
	Write(key string, data []byte) (*WriteInfo, error)

	// Close releases any resources held by the storage backend (connections,
	// file handles, etc.) and should be called during graceful shutdown.
	Close() error
}
