package utils

func StrInSlice(str []string, s string) bool {
	for _, b := range str {
		if b == s {
			return true
		}
	}
	return false
}
