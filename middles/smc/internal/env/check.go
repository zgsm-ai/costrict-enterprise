package env

import "strings"

/**
 *	Check if val exists in set
 */
func InSet(val string, sets ...string) bool {
	for _, s := range sets {
		if s == val {
			return true
		}
	}
	return false
}

/**
 *	Check if val is in string sets (case-insensitive)
 */
func InNcaseSet(val string, sets ...string) bool {
	for _, s := range sets {
		if strings.EqualFold(val, s) {
			return true
		}
	}
	return false
}
