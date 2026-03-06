package prompt

import "strings"

const (
	PREFIX_ENVIRONS   = "shenma:environs:"
	PREFIX_TOOLS      = "shenma:tools:"
	PREFIX_EXTENSIONS = "shenma:extensions:"
	PREFIX_TEMPLATES  = "shenma:templates:"
)

func KeyToID(key, prefix string) string {
	return strings.ReplaceAll(strings.TrimPrefix(key, prefix), ":", ".")
}

func IDtoKey(id, prefix string) string {
	return prefix + strings.ReplaceAll(id, ".", ":")
}
