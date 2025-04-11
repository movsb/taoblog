package notify

import (
	"context"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/logs"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/notify/instant"
	"github.com/movsb/taoblog/service/modules/notify/mailer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Notify struct {
	proto.UnimplementedNotifyServer

	mailer  *mailer.MailerLogger
	instant *instant.NotifyLogger
}

var _ proto.NotifyServer = (*Notify)(nil)

type With func(n *Notify)

func New(ctx context.Context, sr grpc.ServiceRegistrar, options ...With) *Notify {
	n := Notify{}

	for _, opt := range options {
		opt(&n)
	}

	proto.RegisterNotifyServer(sr, &n)

	return &n
}

func (n *Notify) SendEmail(ctx context.Context, in *proto.SendEmailRequest) (*proto.SendEmailResponse, error) {
	auth.MustNotBeGuest(ctx)

	if n.mailer == nil {
		return nil, status.Error(codes.Unimplemented, `未实现邮件服务。`)
	}

	users := utils.Map(in.Users, func(u *proto.SendEmailRequest_User) mailer.User {
		return mailer.User{
			Name:    u.Name,
			Address: u.Address,
		}
	})

	n.mailer.Queue(in.Subject, in.Body, in.FromName, users)

	return &proto.SendEmailResponse{}, nil
}

func (n *Notify) SendInstant(ctx context.Context, in *proto.SendInstantRequest) (*proto.SendInstantResponse, error) {
	auth.MustNotBeGuest(ctx)

	if n.instant == nil {
		return nil, status.Error(codes.Unimplemented, `未实现即时通知服务。`)
	}

	err := n.instant.Notify(in.Subject, in.Body)

	return &proto.SendInstantResponse{}, err
}

func WithMailerLogger(store logs.Logger, mail mailer.Mailer) With {
	return func(n *Notify) {
		n.mailer = mailer.NewMailerLogger(store, mail)
	}
}

func WithInstantLogger(store logs.Logger, inst instant.Notifier) With {
	return func(n *Notify) {
		l := instant.NewNotifyLogger(store)
		l.SetNotifier(inst)
		n.instant = l
	}
}
