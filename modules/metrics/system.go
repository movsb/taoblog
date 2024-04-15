package metrics

var (
	namespace = `taoblog`
	subsystem = ``
)

func bool2string(v bool) string {
	if v {
		return `1`
	}
	return `0`
}
