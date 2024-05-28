package utils

import (
	"crypto/rand"
	"fmt"
)

// 搁这套娃🪆🪆🪆？
// P：Prototype
func ChainFuncs[P func(H) H, H any](h H, ps ...P) H {
	for i := len(ps) - 1; i >= 0; i-- {
		h = ps[i](h)
	}
	return h
}

func Must[A any](a A, e error) A {
	if e != nil {
		panic(e)
	}
	return a
}

// Go 语言多少有点儿大病，以至于我需要写这种东西。
// 是谁当初说不需要三元运算符的？我打断他的 🐶 腿。
// https://en.wikipedia.org/wiki/IIf
// https://blog.twofei.com/716/#没有条件运算符
func IIF[Condition ~bool, Any any](cond Condition, first, second Any) Any {
	if cond {
		return first
	}
	return second
}

func RandomString() string {
	b := [4]byte{}
	rand.Read(b[:])
	return fmt.Sprintf(`xx-%x`, b)
}

func DropLast1[First any, Last any](f First, l Last) First {
	return f
}

func CatchAsError(err *error) {
	if er := recover(); er != nil {
		if er2, ok := er.(error); ok {
			*err = er2
			return
		}
		*err = fmt.Errorf(`%v`, er)
	}
}
