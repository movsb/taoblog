package geoip

import (
	_ "embed"
	"fmt"
	"net"
	"net/netip"
	"sync"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/yl2chen/cidranger"
)

//go:embed china_ipv4_range.bin
var data []byte

var tree = sync.OnceValue(func() cidranger.Ranger {
	ranger := cidranger.NewPCTrieRanger()

	_ = data[len(data)-1]

	for i, n := 0, len(data); i < n; i += 5 {
		_, cidr, _ := net.ParseCIDR(fmt.Sprintf(
			`%d.%d.%d.%d/%d`,
			data[i+0], data[i+1], data[i+2], data[i+3], data[i+4],
		))
		ranger.Insert(cidranger.NewBasicRangerEntry(*cidr))
	}

	return ranger
})

func IsInChina(ip netip.Addr) bool {
	old := net.IP(ip.AsSlice())
	return utils.DropLast1(tree().Contains(old))
}
