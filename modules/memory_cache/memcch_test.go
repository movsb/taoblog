package memory_cache

import (
	"fmt"
	"testing"
	"time"
)

func Test_Memcch(t *testing.T) {
	m := NewMemoryCache(time.Second * 3)
	m.Set("1", "str")
	fmt.Println(m.Get("1"))
	m.Stop()
	time.Sleep(time.Second)
}
