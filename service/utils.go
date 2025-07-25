package service

import (
	"context"
	"time"

	_ "time/tzdata"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/dialers"
	"github.com/movsb/taoblog/modules/geo"
	"github.com/movsb/taoblog/modules/globals"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/xeonx/timeago"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var fixedZone = time.Now().Local().Location()

type Utils struct {
	proto.UnimplementedUtilsServer

	geoLocationResolver geo.GeoLocationResolver
}

func NewUtils(options ...UtilOption) *Utils {
	u := &Utils{}

	for _, opt := range options {
		opt(u)
	}

	return u
}

type UtilOption func(u *Utils)

func WithBaidu(ak, ref string) UtilOption {
	return func(u *Utils) {
		u.geoLocationResolver = geo.NewBaidu(ak, ref)
	}
}

func (u *Utils) FormatTime(ctx context.Context, in *proto.FormatTimeRequest) (*proto.FormatTimeResponse, error) {
	formatted := make([]*proto.FormatTimeResponse_Formatted, len(in.Times))
	for i, ts := range in.Times {
		r := proto.FormatTimeResponse_Formatted{}
		t := time.Unix(int64(ts.Unix), 0)

		// TODO 用浏览器时区。
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

func (u *Utils) ResolveGeoLocation(ctx context.Context, in *proto.ResolveGeoLocationRequest) (_ *proto.ResolveGeoLocationResponse, outErr error) {
	defer utils.CatchAsError(&outErr)

	auth.MustNotBeGuest(ctx)

	if u.geoLocationResolver == nil {
		panic(status.Errorf(codes.Unavailable, `未初始化。`))
	}

	names := utils.Must1(u.geoLocationResolver.ResolveGeoLocation(ctx, in.Latitude, in.Longitude))
	return &proto.ResolveGeoLocationResponse{Names: names}, nil
}
