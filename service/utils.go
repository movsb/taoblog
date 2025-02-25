package service

import (
	"context"
	"time"

	_ "time/tzdata"

	"github.com/movsb/taoblog/modules/dialers"
	"github.com/movsb/taoblog/modules/globals"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/xeonx/timeago"
)

var fixedZone = time.Now().Local().Location()

type Utils struct {
	proto.UnimplementedUtilsServer
}

func NewUtils() *Utils {
	u := &Utils{}
	return u
}

func (u *Utils) FormatTime(ctx context.Context, in *proto.FormatTimeRequest) (*proto.FormatTimeResponse, error) {
	formatted := make([]*proto.FormatTimeResponse_Formatted, len(in.Times))
	for i, ts := range in.Times {
		r := proto.FormatTimeResponse_Formatted{}
		t := time.Unix(int64(ts.Unix), 0)
		r.Friendly = timeago.Chinese.Format(t)
		r.Server = t.In(fixedZone).Format(time.RFC3339)

		if ts.Timezone != `` {
			loc := globals.LoadTimezoneOrDefault(ts.Timezone, nil)
			if loc == nil {
				r.Original = r.Server
			} else {
				r.Original = t.In(loc).Format(time.RFC3339)
			}
		}

		if in.Device != `` {
			loc := globals.LoadTimezoneOrDefault(in.Device, nil)
			if loc != nil {
				r.Device = t.In(loc).Format(time.RFC3339)
			}
		}
		formatted[i] = &r
	}
	return &proto.FormatTimeResponse{
		Formatted: formatted,
	}, nil
}

func (u *Utils) DialRemote(s proto.Utils_DialRemoteServer) error {
	dialer := dialers.NewRemoteDialerManager(s)
	return dialer.Run()
}
