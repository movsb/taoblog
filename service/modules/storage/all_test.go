package storage

import (
	"testing"

	"github.com/movsb/taoblog/service/models"
)

func TestDigest(t *testing.T) {
	if models.Digest([]byte{'1'}) != `c4ca4238a0b923820dcc509a6f75849b` {
		panic(`digest error`)
	}
}
