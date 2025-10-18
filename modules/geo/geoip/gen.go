//go:build geoip

package main

import (
	"bufio"
	"bytes"
	"math"
	"net/http"
	"net/netip"
	"os"
	"strconv"

	"github.com/movsb/taoblog/modules/utils"
)

func main() {
	rsp := utils.Must1(http.Get(`http://ftp.apnic.net/apnic/stats/apnic/delegated-apnic-latest`))
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		panic(`status != 200`)
	}

	fp := utils.Must1(os.Create(`china_ipv4_range.bin`))
	defer fp.Close()

	scanner := bufio.NewScanner(rsp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if !bytes.Contains(line, []byte(`|CN|ipv4|`)) {
			continue
		}
		parts := bytes.Split(line, []byte{'|'})
		if len(parts) < 5 {
			continue
		}

		start := netip.MustParseAddr(string(parts[3]))
		count := utils.Must1(strconv.Atoi(string(parts[4])))
		bits := 32 - int(math.Log2(float64(count)))

		as4 := start.As4()

		fp.Write(as4[:])
		fp.Write([]byte{byte(bits)})
	}
}
