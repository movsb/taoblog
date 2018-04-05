package main

import (
	"fmt"
	"strings"
)

func joinInts(ints []int64, delim string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ints)), delim), "[]")
}

func strInSlice(str []string, s string) bool {
	for _, b := range str {
		if b == s {
			return true
		}
	}
	return false
}
