package memory_cache

import (
	"fmt"
	"testing"
	"time"
)

func Test_Memcch(t *testing.T) {
	m := NewMemoryCache(time.Second * 3)
	fmt.Println(m.Get("1", func(key string) (interface{}, error) {
		return "str", nil
	}))
	m.Stop()
	time.Sleep(time.Second)
}
