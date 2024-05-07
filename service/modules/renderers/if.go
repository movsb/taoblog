package renderers

type Renderer interface {
	Render(source string) (string, string, error)
}

// 解析文件系统路径为进程的相对路径。
// TODO：移除，内部解析出这个路径只是为了打开文件，为啥不直接返回文件系统对象？
type PathResolver interface {
	Resolve(path string) string
}
