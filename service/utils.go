package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/notify"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/xeonx/timeago"
)

var fixedZone = time.FixedZone(``, 8*60*60)

type Utils struct {
	proto.UnimplementedUtilsServer

	instantNotifier notify.InstantNotifier
}

func NewUtils(instantNotifier notify.InstantNotifier) *Utils {
	return &Utils{
		instantNotifier: instantNotifier,
	}
}

func (u *Utils) FormatTime(ctx context.Context, in *proto.FormatTimeRequest) (*proto.FormatTimeResponse, error) {
	formatted := make([]*proto.FormatTimeResponse_Formatted, len(in.Unix))
	for i, u := range in.Unix {
		r := proto.FormatTimeResponse_Formatted{}
		t := time.Unix(int64(u), 0).In(fixedZone)
		r.Friendly = timeago.Chinese.Format(t)
		r.Rfc3339 = t.Format(time.RFC3339)
		formatted[i] = &r
	}
	return &proto.FormatTimeResponse{
		Formatted: formatted,
	}, nil
}

func (u *Utils) InstantNotify(ctx context.Context, in *proto.InstantNotifyRequest) (*proto.InstantNotifyResponse, error) {
	ac := auth.Context(ctx)
	if !ac.User.IsSystem() && !ac.User.IsAdmin() && !(ac.User.ID == auth.Notify.ID) {
		return nil, errors.New(`此操作无权限。`)
	}
	if u.instantNotifier == nil {
		log.Println(`即时通知：`, in.Title, in.Message)
	} else {
		u.instantNotifier.InstantNotify(in.Title, in.Message)
	}
	return &proto.InstantNotifyResponse{}, nil
}
