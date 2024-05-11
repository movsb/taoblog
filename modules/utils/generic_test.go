package utils_test

import (
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func TestChainFunc(t *testing.T) {
	f := utils.ChainFuncs(1,
		func(i int) int {
			return i + 1
		},
		func(i int) int {
			return i * 2
		},
	)
	if f != 3 {
		t.Fatal(`not equal`)
	}
}
