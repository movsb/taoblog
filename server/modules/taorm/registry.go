package taorm

import (
	"database/sql"
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
	fields map[string]_FieldInfo
}

func newStructInfo() *_StructInfo {
	return &_StructInfo{
		fields: make(map[string]_FieldInfo),
	}
}

// FieldPointers ...
func (s *_StructInfo) FieldPointers(base uintptr, fields []string) (ptrs []interface{}) {
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
func register(_struct interface{}) *_StructInfo {
	rwLock.Lock()
	defer rwLock.Unlock()
	t := structType(_struct)
	typeName := t.String()
	structInfo := newStructInfo()
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
	return register(_struct)
}

func getPointers(out interface{}, rows *sql.Rows) []interface{} {
	fields, _ := rows.Columns()
	base := baseFromInterface(out)
	return getRegistered(out).FieldPointers(base, fields)
}
