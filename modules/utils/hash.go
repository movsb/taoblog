package utils

import (
	"crypto/md5"
	"fmt"
)

func Md5Str(str string) string {
	md5 := md5.New()
	md5.Write([]byte(str))
	return fmt.Sprintf("%x", md5.Sum(nil))
}
