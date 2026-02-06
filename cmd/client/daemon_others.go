//go:build !linux

package client

func statDisk(wd string) (int64, int64) {
	return 0, 0
}

func statMemory() (int64, int64) {
	return 0, 0
}
