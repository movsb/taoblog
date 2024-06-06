package service

import (
	"context"
	"time"

	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/xeonx/timeago"
)

var fixedZone = time.FixedZone(``, 8*60*60)

type Utils struct {
	proto.UnimplementedUtilsServer
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
