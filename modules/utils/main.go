package utils

import "strconv"

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
