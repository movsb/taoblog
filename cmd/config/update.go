package config

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Updater struct {
	obj any
}

type Segment struct {
	Key   string
	Index any
}

var reSplit = regexp.MustCompile(`(\w+)\[(\w+|\d+)\]|\w+|\.`)

func NewUpdater(ptr any) *Updater {
	return &Updater{obj: ptr}
}

func (u *Updater) MustApply(path string, value string) {
	segments := u.parse(path)
	p := u.find(u.obj, segments)
	u.set(p, value)
}

func (u *Updater) Apply(path string, value string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()
	u.MustApply(path, value)
	return nil
}

func (u *Updater) parse(path string) []Segment {
	segments := reSplit.FindAllStringSubmatch(path, -1)
	typed := []Segment{}
	buf := bytes.NewBuffer(nil)
	lastIsDot := true
	for _, seg := range segments {
		if seg[0] == "." && lastIsDot {
			panic(`unexpected dot`)
		}
		lastIsDot = seg[0] == "."
		for i, key := range seg {
			if i == 0 {
				buf.WriteString(key)
			}
		}

		// 奇怪，为什么会 match 到空字符串？
		switch len(seg) {
		default:
			panic(`won't be here`)
		case 3:
			if seg[0] == "." {
				continue
			}
			if seg[1] == "" {
				typed = append(typed, Segment{Key: seg[0]})
				continue
			}
			n, err := strconv.Atoi(seg[2])
			if err == nil {
				typed = append(typed, Segment{Key: seg[1], Index: n})
			} else {
				typed = append(typed, Segment{Key: seg[1], Index: seg[2]})
			}
		}
	}
	if buf.String() != path {
		panic(`bad path`)
	}
	return typed
}

func (u *Updater) Find(path string) any {
	segments := u.parse(path)
	p := u.find(u.obj, segments)
	return p
}

func (u *Updater) find(obj any, segments []Segment) any {
	value := reflect.ValueOf(obj)

	// 不是也行，那就是简单赋值了，无意义。
	if value.Type().Kind() != reflect.Pointer {
		panic(`expect pointer`)
	}

	if len(segments) < 1 {
		return obj
	}

	seg := segments[0]

	field := value.Elem().FieldByName(seg.Key)
	if !field.IsValid() {
		ty := value.Elem().Type()
		for i, n := 0, ty.NumField(); i < n; i++ {
			tag := ty.Field(i).Tag.Get(`yaml`)
			if tag == "" {
				continue
			}
			fields := strings.SplitN(tag, ",", 2)
			if len(fields) <= 0 {
				continue
			}
			if fields[0] == seg.Key {
				field = value.Elem().Field(i)
				break
			}
		}
	}
	if !field.IsValid() {
		panic(`invalid field: ` + seg.Key)
	}

	if index, ok := seg.Index.(int); ok {
		if field.Kind() != reflect.Slice {
			panic(`not slice`)
		}
		field = field.Index(index)
	} else if index, ok := seg.Index.(string); ok {
		if field.Kind() != reflect.Map {
			panic(`not map`)
		}
		field = field.MapIndex(reflect.ValueOf(index))
	}

	if !field.IsValid() {
		panic(`invalid field: ` + seg.Key)
	}

	return u.find(field.Addr().Interface(), segments[1:])
}

func (u *Updater) set(p any, value string) {
	vpe := reflect.ValueOf(p).Elem()

	switch vpe.Type().Kind() {
	case reflect.Bool:
		var b bool
		switch strings.ToLower(value) {
		case `true`, `yes`, `1`:
			b = true
		case `false`, `no`, `0`:
			b = false
		default:
			panic(`invalid bool value`)
		}
		vpe.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			panic(err)
		}
		vpe.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			panic(err)
		}
		vpe.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic(err)
		}
		vpe.SetFloat(f)
	case reflect.String:
		vpe.SetString(value)
	case reflect.Struct, reflect.Slice:
		vpe.SetZero()
		a := vpe.Interface()
		_ = a
		if err := yaml.UnmarshalStrict([]byte(value), vpe.Addr().Interface()); err != nil {
			panic(err)
		}
	default:
		panic(`unknown value type`)
	}
}
