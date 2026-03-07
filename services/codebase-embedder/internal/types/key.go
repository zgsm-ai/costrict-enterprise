package types

import "fmt"

const syncVersionKeyPrefixFmt = "codebase_embedder:sync_version:%d"

// SyncVersionKey returns the Redis key for storing versions
func SyncVersionKey(syncId int32) string {
	return fmt.Sprintf(syncVersionKeyPrefixFmt, syncId)
}
