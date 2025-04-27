package storage

import "testing"

func TestDigest(t *testing.T) {
	if digest([]byte{'1'}) != `c4ca4238a0b923820dcc509a6f75849b` {
		panic(`digest error`)
	}
}
