package datetime

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	f := fmt.Println
	f(MyLocal())
	f(My2Gmt("2018-06-07 13:00:27"))
}
