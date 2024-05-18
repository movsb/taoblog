package commentgeo_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	commentgeo "github.com/movsb/taoblog/service/modules/comment_geo"
)

func TestAll(t *testing.T) {
	cg := commentgeo.New(context.TODO())
	cg.Queue(`8.8.8.8`)
	cg.Queue(`1.1.1.1`)
	time.Sleep(time.Second)
	fmt.Println(cg.Get(`1.1.1.1`))
}
