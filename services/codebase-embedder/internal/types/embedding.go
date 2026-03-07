package types

type SourceFile struct {
	CodebaseId   int32
	CodebasePath string
	CodebaseName string
	Name         string
	Path         string
	Content      []byte
	Language     string
}

// CodeChunk represents a chunk of code with associated metadata.
type CodeChunk struct {
	CodebaseId   int32
	CodebasePath string
	CodebaseName string
	Language     string
	Content      []byte // The actual code snippet
	FilePath     string // The BasePath to the file this block came from
	Range        []int  // start from zero, startLine, startColumn, endLine, endColumn
	TokenCount   int    // The number of tokens in this block
}

// CodeChunkPathUpdate represents a request to update a code chunk's file path
type CodeChunkPathUpdate struct {
	CodebaseId  int32  // The ID of the codebase
	OldFilePath string // The old file path
	NewFilePath string // The new file path
}
