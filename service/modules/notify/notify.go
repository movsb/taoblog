package notify

import (
	"cmp"
	"context"
	"database/sql"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/logs"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/notify/instant"
	"github.com/movsb/taoblog/service/modules/notify/mailer"
	"github.com/movsb/taorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Notify struct {
	proto.UnimplementedNotifyServer

	db           *taorm.DB
	mailer       mailer.MailSender
	storedNotify *instant.NotifyLogger

	defaultChanifyToken string
}

var _ proto.NotifyServer = (*Notify)(nil)

type With func(n *Notify)

func New(ctx context.Context, sr grpc.ServiceRegistrar, db *sql.DB, options ...With) *Notify {
	n := Notify{
		db:           taorm.NewDB(db),
		storedNotify: instant.NewNotifyLogger(logs.NewLogStore(db)),
	}

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

	n.mailer.Send(in.Subject, in.Body, in.FromName, users)

	return &proto.SendEmailResponse{}, nil
}

func (n *Notify) SendInstant(ctx context.Context, in *proto.SendInstantRequest) (*proto.SendInstantResponse, error) {
	auth.MustNotBeGuest(ctx)

	if n.defaultChanifyToken == `` && in.ChanifyToken == `` {
		return nil, status.Error(codes.Unimplemented, `未实现即时通知服务。`)
	}

	err := n.storedNotify.Notify(in.Subject, in.Body, cmp.Or(in.ChanifyToken, n.defaultChanifyToken))

	return &proto.SendInstantResponse{}, err
}

func WithMailer(mail mailer.MailSender) With {
	return func(n *Notify) {
		n.mailer = mail
	}
}

func WithDefaultChanifyToken(t string) With {
	return func(n *Notify) {
		n.defaultChanifyToken = t
	}
}
