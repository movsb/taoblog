package service

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"

	_ "time/tzdata"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/dialers"
	"github.com/movsb/taoblog/modules/notify"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/phuslu/lru"
	"github.com/xeonx/timeago"
)

var fixedZone = time.Now().Local().Location()

type Utils struct {
	proto.UnimplementedUtilsServer

	instantNotifier notify.InstantNotifier

	RemoteDialer atomic.Pointer[dialers.RemoteDialerManager]

	timeLocations *lru.TTLCache[string, *time.Location]
}

func NewUtils(instantNotifier notify.InstantNotifier) *Utils {
	u := &Utils{
		instantNotifier: instantNotifier,
		timeLocations:   lru.NewTTLCache[string, *time.Location](16),
	}
	u.RemoteDialer.Store(nil)
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
			loc, err, _ := u.timeLocations.GetOrLoad(ctx, ts.Timezone, func(ctx context.Context, s string) (*time.Location, time.Duration, error) {
				loc, err := time.LoadLocation(s)
				// log.Println(`加载时区：`, s, loc, err)
				return loc, time.Hour, err
			})
			if err != nil {
				log.Println(err)
				r.Original = r.Server
			} else {
				// log.Println(`时区：`, loc)
				r.Original = t.In(loc).Format(time.RFC3339)
			}
		}

		if in.Device != `` {
			loc, err, _ := u.timeLocations.GetOrLoad(ctx, in.Device, func(ctx context.Context, s string) (*time.Location, time.Duration, error) {
				loc, err := time.LoadLocation(s)
				// log.Println(`加载时区：`, s, loc, err)
				return loc, time.Hour, err
			})
			if err != nil {
				log.Println(err)
			} else {
				// log.Println(`时区：`, loc)
				r.Device = t.In(loc).Format(time.RFC3339)
			}
		}
		formatted[i] = &r
	}
	return &proto.FormatTimeResponse{
		Formatted: formatted,
	}, nil
}

func (u *Utils) InstantNotify(ctx context.Context, in *proto.InstantNotifyRequest) (*proto.InstantNotifyResponse, error) {
	log.Println(`即时通知：`, in.Title, in.Message)
	ac := auth.Context(ctx)
	if !ac.User.IsSystem() && !ac.User.IsAdmin() && !(ac.User.ID == auth.Notify.ID) {
		return nil, errors.New(`此操作无权限。`)
	}
	if u.instantNotifier != nil {
		u.instantNotifier.InstantNotify(in.Title, in.Message)
	}
	return &proto.InstantNotifyResponse{}, nil
}

func (u *Utils) DialRemote(s proto.Utils_DialRemoteServer) error {
	dialer := dialers.NewRemoteDialerManager(s)
	u.RemoteDialer.Store(dialer)
	defer u.RemoteDialer.CompareAndSwap(dialer, nil)
	return dialer.Run()
}
