package globals

import "sync"

// 遇到了 GraphViz 的并发问题，所以这里用全局锁来保护。
// https://github.com/goccy/go-graphviz/issues/117
var GraphVizLock sync.Mutex
