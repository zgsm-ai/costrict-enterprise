package types

type IndexType string

var (
	Embedding IndexType = "embedding"
	CodeGraph IndexType = "codegraph"
	All       IndexType = "all"
)
