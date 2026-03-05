package parser

type Parser interface {
	IsCodeSyntax(code string) bool
	InterceptSyntaxErrorCode(choicesText, prefix, suffix string) string
	ExtractAccurateBlockPrefixSuffix(prefix, suffix string) (string, string)
}
