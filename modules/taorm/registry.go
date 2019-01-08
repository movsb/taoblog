package taorm

import (
	"fmt"
	"reflect"
	"sync"
)

type _FieldInfo struct {
	offset uintptr
	kind   reflect.Kind
}

// StructInfo ...
type _StructInfo struct {
	tableName string
	fields    map[string]_FieldInfo
}

func newStructInfo() *_StructInfo {
	return &_StructInfo{
		fields: make(map[string]_FieldInfo),
	}
}

func (s *_StructInfo) mustBeTable() *_StructInfo {
	if s.tableName == "" {
		panic("not table")
	}
	return s
}

// FieldPointers ...
func (s *_StructInfo) FieldPointers(base uintptr, fields []string) (ptrs []interface{}) {
	ptrs = make([]interface{}, 0, len(fields))
	for _, field := range fields {
		if fs, ok := s.fields[field]; ok {
			i := ptrToInterface(base+fs.offset, fs.kind)
			ptrs = append(ptrs, i)
			continue
		}
		panic(fmt.Errorf("no place to save field: %s", field))
	}
	return
}

var structs = make(map[string]*_StructInfo)
var rwLock = &sync.RWMutex{}

func structType(_struct interface{}) reflect.Type {
	t := reflect.TypeOf(_struct)
	if t == nil {
		panic("Register with nil")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic("Register with non-struct")
	}
	return t
}

// register ...
func register(_struct interface{}, tableName string) *_StructInfo {
	rwLock.Lock()
	defer rwLock.Unlock()
	t := structType(_struct)
	typeName := t.PkgPath() + "." + t.Name()
	if si, ok := structs[typeName]; ok {
		return si
	}
	structInfo := newStructInfo()
	structInfo.tableName = tableName
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if isColumnField(f) {
			columnName := getColumnName(f)
			if columnName == "" {
				continue
			}
			fieldInfo := _FieldInfo{
				offset: f.Offset,
				kind:   f.Type.Kind(),
			}
			structInfo.fields[columnName] = fieldInfo
		}
	}
	structs[typeName] = structInfo
	fmt.Printf("taorm: registered: %s\n", typeName)
	return structInfo
}

func getRegistered(_struct interface{}) *_StructInfo {
	name := structType(_struct).String()
	rwLock.RLock()
	if si, ok := structs[name]; ok {
		rwLock.RUnlock()
		return si
	}
	rwLock.RUnlock()
	return register(_struct, "")
}

func getPointers(out interface{}, columns []string) []interface{} {
	base := baseFromInterface(out)
	info := getRegistered(out)
	return info.FieldPointers(base, columns)
}
