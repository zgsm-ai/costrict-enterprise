// Package storage defines abstractions for ChatLog persistence backends.
package storage

// StorageBackend abstracts the underlying storage mechanism for ChatLog persistence.
// All ChatLog write operations must go through this interface, allowing the system
// to swap between different storage backends (e.g., local filesystem, S3) without
// changing the caller logic.
type StorageBackend interface {
	// Write persists data under the given key.
	// The key is a relative object path (e.g., "2026/04/03/chat-abc123.json")
	// whose semantics are unified across backends — it maps to a file path for
	// local storage or an object key for S3-compatible storage.
	Write(key string, data []byte) error

	// Close releases any resources held by the storage backend (connections,
	// file handles, etc.) and should be called during graceful shutdown.
	Close() error
}
