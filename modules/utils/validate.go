package utils

import (
	"net/url"
	"regexp"
	"strings"
)

// hand write regex, not tested well.
var regexpValidEmail = regexp.MustCompile(`^[+-_.a-zA-Z0-9]+@[[:alnum:]]+(\.[[:alnum:]]+)+$`)

func IsEmail(email string) bool {
	return regexpValidEmail.MatchString(email)
}

func IsURL(Url string, addScheme bool) bool {
	if !strings.Contains(Url, `://`) && addScheme {
		Url = "http://" + Url
	}
	u, err := url.Parse(Url)
	if err != nil {
		return false
	}
	if !strings.Contains(u.Host, ".") {
		return false
	}
	return true
}

func Must[A any, Error error](a A, e Error) A {
	if error(e) != nil {
		panic(e)
	}
	return a
}

// Go 语言多少有点儿大病，以至于我需要写这种东西。
// 是谁当初说不需要三元运算符的？我打断他的 🐶 腿。
// https://en.wikipedia.org/wiki/IIf
// https://blog.twofei.com/716/#没有条件运算符
func IIF[Condition bool, Any any](cond Condition, first, second Any) Any {
	if cond {
		return first
	}
	return second
}
