package taorm

import (
	"go/ast"
	"reflect"
	"regexp"
	"strings"
	"unsafe"
)

// https://gist.github.com/stoewer/fbe273b711e6a06315d19552dd4d33e60
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func isColumnField(field reflect.StructField) bool {
	if !ast.IsExported(field.Name) {
		return false
	}
	switch field.Type.Kind() {
	case reflect.Bool, reflect.String:
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func getColumnName(field reflect.StructField) string {
	tag := field.Tag.Get("taorm")
	kvs := strings.Split(tag, ",")
	for _, kv := range kvs {
		s := strings.Split(kv, ":")
		switch s[0] {
		case "name":
			if len(s) > 1 {
				return s[1]
			}
			return ""
		}
	}
	return toSnakeCase(field.Name)
}

type _EmptyEface struct {
	typ *struct{}
	ptr unsafe.Pointer
}

func baseFromInterface(ptr interface{}) uintptr {
	return uintptr((*_EmptyEface)(unsafe.Pointer(&ptr)).ptr)
}

func ptrToInterface(ptr uintptr, kind reflect.Kind) interface{} {
	var i interface{}
	var p = unsafe.Pointer(ptr)
	switch kind {
	case reflect.Bool:
		i = (*bool)(p)
	case reflect.String:
		i = (*string)(p)
	case reflect.Int:
		i = (*int)(p)
	case reflect.Int8:
		i = (*int8)(p)
	case reflect.Int16:
		i = (*int16)(p)
	case reflect.Int32:
		i = (*int32)(p)
	case reflect.Int64:
		i = (*int64)(p)
	case reflect.Uint:
		i = (*uint)(p)
	case reflect.Uint8:
		i = (*uint8)(p)
	case reflect.Uint16:
		i = (*uint16)(p)
	case reflect.Uint32:
		i = (*uint32)(p)
	case reflect.Uint64:
		i = (*uint64)(p)
	default:
		panic("unknown kind")
	}
	return i
}
