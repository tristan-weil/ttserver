package utils

import (
	"path/filepath"
	"strings"
)

func FilePathClean(s *string) *string {
	return String(filepath.Clean(StringValue(s)))
}

func IsStringEmpty(v *string) bool {
	return v == nil || StringValue(v) == ""
}

func IsStringSliceEmpty(v []string) bool {
	return len(v) == 0
}

func NotStringEmpty(v *string) bool {
	return !IsStringEmpty(v)
}

func NotStringSliceEmpty(v []string) bool {
	return !IsStringSliceEmpty(v)
}

func SplitRune(s string, separators ...rune) []string {
	f := func(r rune) bool {
		for _, s := range separators {
			if r == s {
				return true
			}
		}

		return false
	}

	return strings.FieldsFunc(s, f)
}
