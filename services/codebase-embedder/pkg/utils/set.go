package utils

import (
	"github.com/emirpasic/gods/sets/treeset"
	"strconv"
)

func NewTimestampTreeSet() *treeset.Set {
	return treeset.NewWith(func(a, b interface{}) int {
		aUnix, _ := strconv.ParseInt(a.(string), 10, 64)
		bUnix, _ := strconv.ParseInt(b.(string), 10, 64)
		switch {
		case aUnix < bUnix:
			return -1
		case aUnix > bUnix:
			return 1
		default:
			return 0
		}
	})
}
