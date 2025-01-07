package service

import (
	"context"
	"errors"
	"log"
	"time"

	_ "time/tzdata"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/dialers"
	"github.com/movsb/taoblog/modules/globals"
	"github.com/movsb/taoblog/modules/notify"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/xeonx/timeago"
)

var fixedZone = time.Now().Local().Location()

type Utils struct {
	proto.UnimplementedUtilsServer

	instantNotifier notify.Notifier
}

func NewUtils(instantNotifier notify.Notifier) *Utils {
	u := &Utils{
		instantNotifier: instantNotifier,
	}
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

func (u *Utils) InstantNotify(ctx context.Context, in *proto.InstantNotifyRequest) (*proto.InstantNotifyResponse, error) {
	log.Println(`即时通知：`, in.Title, in.Message)
	ac := auth.Context(ctx)
	if !ac.User.IsSystem() && !ac.User.IsAdmin() && !(ac.User.ID == auth.Notify.ID) {
		return nil, errors.New(`此操作无权限。`)
	}
	if u.instantNotifier != nil {
		u.instantNotifier.Notify(in.Title, in.Message)
	}
	return &proto.InstantNotifyResponse{}, nil
}

func (u *Utils) DialRemote(s proto.Utils_DialRemoteServer) error {
	dialer := dialers.NewRemoteDialerManager(s)
	return dialer.Run()
}
