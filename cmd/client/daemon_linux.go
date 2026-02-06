//go:build linux

package client

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func statDisk(wd string) (int64, int64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(wd, &stat); err != nil {
		log.Println(err)
		return 0, 0
	}
	total := stat.Blocks * uint64(stat.Bsize)
	avail := stat.Bavail * uint64(stat.Bsize)
	return int64(total), int64(avail)
}

func statMemory() (int64, int64) {
	all, err := os.ReadFile(`/proc/meminfo`)
	if err != nil {
		log.Println(err)
		return 0, 0
	}
	var total, avail int64
	scanner := bufio.NewScanner(bytes.NewReader(all))

	parseSize := func(n, u string) int64 {
		nn, _ := strconv.ParseInt(n, 10, 63)
		switch u {
		case `kB`:
			return nn * 1024
		case `mB`:
			return nn * 1024 * 1024
		}
		return 0
	}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 3 {
			continue
		}
		switch fields[0] {
		case `MemTotal:`:
			total = parseSize(fields[1], fields[2])
		case `MemAvailable:`:
			avail = parseSize(fields[1], fields[2])
		}
	}
	if scanner.Err() != nil {
		log.Println(scanner.Err())
		return 0, 0
	}

	return total, avail
}
