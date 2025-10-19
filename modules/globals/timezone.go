package globals

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/phuslu/lru"
)

// 服务器当前时区。
func SystemTimezone() *time.Location {
	return time.Now().Location()
}

var (
	timezoneLocations         *lru.LRUCache[string, *time.Location]
	onceInitTimezoneLocations sync.Once
)

func LoadTimezoneOrDefault(location string, def *time.Location) *time.Location {
	onceInitTimezoneLocations.Do(func() {
		timezoneLocations = lru.NewLRUCache[string, *time.Location](16)
	})
	loc, err, _ := timezoneLocations.GetOrLoad(context.Background(), location, func(ctx context.Context, s string) (*time.Location, error) {
		if s == `` {
			s = `Local`
		}
		return time.LoadLocation(s)
	})
	if err != nil {
		log.Println(`加载时区出错：`, err)
		return def
	}
	return loc
}
