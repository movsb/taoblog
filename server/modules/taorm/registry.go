package taorm

import (
	"database/sql"
	"reflect"
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
		panic("no place to save field")
	}
	return
}

var structs = make(map[string]*_StructInfo)

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

// Register ...
func Register(_struct interface{}) {
	t := structType(_struct)
	typeName := t.String()
	structInfo := newStructInfo()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if isColumnField(f) {
			columnName := toSnakeCase(f.Name)
			fieldInfo := _FieldInfo{
				offset: f.Offset,
				kind:   f.Type.Kind(),
			}
			structInfo.fields[columnName] = fieldInfo
		}
	}
	structs[typeName] = structInfo
}

func getRegistered(_struct interface{}) *_StructInfo {
	t := structType(_struct)
	typeName := t.String()
	if si, ok := structs[typeName]; ok {
		return si
	}
	panic("not registered")
}

func getPointers(out interface{}, rows *sql.Rows) []interface{} {
	fields, _ := rows.Columns()
	base := baseFromInterface(out)
	return getRegistered(out).FieldPointers(base, fields)
}

/*
func Test() {
	Register(Comment{})
	var err error
	dataSource := fmt.Sprintf("%[1]s:%[1]s@/%[1]s", "taoblog")
	gdb, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}
	defer gdb.Close()

	rows, err := gdb.Query("select id,author,email,ip,date from comments")
	if err != nil {
		panic(err)
	}
	var c Comment
	rows.Next()
	if err := Scan(rows, &c); err != nil {
		panic(err)
	}
	fmt.Println(c)
}
*/
