package config

import (
	"bytes"
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Settable interface {
	// 设置之前进行校验。
	// paths：如果为空，则表示整体设置。
	//        如果不为空，则表示设置对应名称的字段的值。
	// obj：对象到具体的值（非值的指针）。
	//
	// 注意：
	// - 具体的校验方式自己决定。如果 key 之前不相互依赖/影响，则可以单独判断。
	// - 无需判断 path.Index 是否会对应到普通字段类型（非 struct、map）。
	// - 无需判断值的类型，只要 path 对应，类型一定可强制断言且赋值。
	// - 无需判断 path 是否存在。
	BeforeSet(paths Segments, obj any) error
}

type AfterSettable interface {
	AfterSet(paths Segments, obj any)
}

type Saver interface {
	CanSave()
}

type Updater struct {
	obj any
}

type Segment struct {
	Key   string
	Index any
}

type Segments []Segment

func (s Segments) At(i int) Segment {
	if i < len(s) && i >= 0 {
		return s[i]
	}
	return Segment{}
}

var reSplit = regexp.MustCompile(`(\w+)\[(\w+|\d+)\]|\w+|\.`)

func NewUpdater(ptr any) *Updater {
	return &Updater{obj: ptr}
}

func (u *Updater) MustApply(path string, value string, save func(path string, value string)) {
	segments := u.parse(path)
	saver, saveSegmentIndex, settable, settableSegments, p := u.find(u.obj, segments, 0, nil, -1, nil, -1)

	// 创建值的副本，在设置之前先检查是否合法。
	new := reflect.New(reflect.TypeOf(p).Elem())
	u.set(new.Interface(), value)
	newVal := new.Elem().Interface()

	if saver == nil {
		saver, _ = u.obj.(Saver)
		if saver == nil {
			panic("尝试修改的值找不到存储者。")
		}
	}

	if settable == nil {
		settable, _ = u.obj.(Settable)
		if settable == nil {
			settable, _ = u.obj.(AfterSettable)
		}
	}

	if bs, ok := settable.(Settable); ok {
		if err := bs.BeforeSet(segments[settableSegments+1:], newVal); err != nil {
			panic(err)
		}
	}

	u.set(p, value)

	if as, ok := settable.(AfterSettable); ok {
		as.AfterSet(segments[settableSegments+1:], newVal)
	}

	var saverPath string
	for i := 0; i <= saveSegmentIndex; i++ {
		if i != saveSegmentIndex && segments[i].Index != nil {
			panic(`数组或对象必须整体存储。`)
		}
		if saverPath != "" {
			saverPath += "."
		}
		saverPath += segments[i].Key
	}

	if b, err := json.Marshal(saver); err != nil {
		panic(err)
	} else {
		save(saverPath, string(b))
	}
}

func (u *Updater) EachSaver(fn func(path string, obj any)) {
	recurseSaver(u.obj, "", func(path string, obj any) {
		fn(path[1:], obj)
	})
}

func recurseSaver(obj any, path string, fn func(path string, obj any)) {
	value := reflect.ValueOf(obj)
	// 不是也行，那就是简单赋值了，无意义。
	if value.Type().Kind() != reflect.Pointer {
		panic(`expect pointer`)
	}

	if value.Type().Implements(reflect.TypeOf((*Saver)(nil)).Elem()) {
		fn(path, obj)
		return
	}

	elemTy := value.Elem().Type()
	if elemTy.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < elemTy.NumField(); i++ {
		tag := elemTy.Field(i).Tag.Get(`yaml`)
		if tag == "" {
			continue
		}
		fields := strings.SplitN(tag, ",", 2)
		if len(fields) <= 0 {
			continue
		}
		recurseSaver(value.Elem().Field(i).Addr().Interface(), path+"."+fields[0], fn)
	}
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
	_, _, _, _, p := u.find(u.obj, segments, 0, nil, -1, nil, -1)
	return p
}

// 在 obj 对象中查找 segments 依次对应的元素。
// save 用于记录谁可以保存此元素。
// settable 用于记录谁可以校验此元素。
// 如果 saver, settable 为 nil，则有可能是 obj 本身。
func (u *Updater) find(obj any, segments []Segment, index int, saver Saver, saverSegment int, settable any, settableSegments int) (Saver, int, any, int, any) {
	if len(segments[index:]) < 1 {
		return saver, saverSegment, settable, settableSegments, obj
	}

	value := reflect.ValueOf(obj)
	if value.Type().Kind() != reflect.Pointer {
		panic(`expect pointer`)
	}

	seg := segments[index]

	// 必须通过 tag 拿，否则可能出现 json、yaml 不一致。
	// field := value.Elem().FieldByName(seg.Key)
	var field reflect.Value
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

	var ownerField reflect.Value

	if index, ok := seg.Index.(int); ok {
		if field.Kind() != reflect.Slice {
			panic(`not slice`)
		}
		ownerField = field
		field = field.Index(index)
	} else if index, ok := seg.Index.(string); ok {
		if field.Kind() != reflect.Map {
			panic(`not map`)
		}
		ownerField = field
		field = field.MapIndex(reflect.ValueOf(index))
	} else {
		// ownerField = value.Elem()
	}

	if !field.IsValid() {
		panic(`invalid field: ` + seg.Key)
	}

	if ownerField.IsValid() {
		if reflect.PointerTo(ownerField.Type()).Implements(reflect.TypeOf((*Settable)(nil)).Elem()) ||
			reflect.PointerTo(ownerField.Type()).Implements(reflect.TypeOf((*AfterSettable)(nil)).Elem()) {
			reflect.ValueOf(&settable).Elem().Set(ownerField.Addr())
			settableSegments = index
		}
		if reflect.PointerTo(ownerField.Type()).Implements(reflect.TypeOf((*Saver)(nil)).Elem()) {
			if saver != nil {
				panic(`发现上级可保存。`)
			}
			reflect.ValueOf(&saver).Elem().Set(ownerField.Addr())
			saverSegment = index
		}
	}
	if reflect.PointerTo(field.Type()).Implements(reflect.TypeOf((*Settable)(nil)).Elem()) ||
		reflect.PointerTo(field.Type()).Implements(reflect.TypeOf((*AfterSettable)(nil)).Elem()) {
		reflect.ValueOf(&settable).Elem().Set(field.Addr())
		settableSegments = index
	}
	if reflect.PointerTo(field.Type()).Implements(reflect.TypeOf((*Saver)(nil)).Elem()) {
		if saver != nil {
			panic(`发现上级可保存。`)
		}
		reflect.ValueOf(&saver).Elem().Set(field.Addr())
		saverSegment = index
	}

	return u.find(field.Addr().Interface(), segments, index+1, saver, saverSegment, settable, settableSegments)
}

func (*Updater) set(p any, value string) {
	vpe := reflect.ValueOf(p).Elem()

	value = strings.TrimRight(value, "\n")

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
