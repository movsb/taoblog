package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func ToInt64(s string) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	return n, err
}

func MustToInt64(s string) int64 {
	n, err := ToInt64(s)
	if err != nil {
		panic(err)
	}
	return n
}

func JoinInts(ints []int64, delim string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ints)), delim), "[]")
}
