package service

import (
	"context"
	"fmt"

	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"github.com/movsb/taoblog/service/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type TwitterPostPublisher interface {
	Preview(*proto.Post) (*proto.TwitterPostPreview, error)
	Publish(context.Context, *proto.Post) (string, error)
}

func (s *Service) postForTwitter(ctx context.Context, id int64) (*proto.Post, error) {
	account := user.MustNotBeGuest(ctx)
	if s.twitterPostPublisher == nil {
		return nil, status.Error(codes.FailedPrecondition, "Twitter 同步尚未配置")
	}
	post, err := s.GetPost(ctx, &proto.GetPostRequest{
		Id: int32(id),
		GetPostOptions: &proto.GetPostOptions{ContentOptions: &proto.PostContentOptions{
			WithContent: true,
		}},
	})
	if err != nil {
		return nil, err
	}
	if account.User.ID != int64(post.UserId) {
		return nil, status.Error(codes.PermissionDenied, noPerm)
	}
	if post.Type != "tweet" || post.Status != models.PostStatusPublic {
		return nil, status.Error(codes.FailedPrecondition, "只有公开的碎碎念可以同步到 Twitter")
	}
	return post, nil
}

func (s *Service) PreviewPostForTwitter(ctx context.Context, in *proto.PreviewTwitterPostRequest) (*proto.TwitterPostPreview, error) {
	post, err := s.postForTwitter(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	preview, err := s.twitterPostPublisher.Preview(post)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("生成 Twitter 预览失败：%v", err))
	}
	return preview, nil
}

func (s *Service) PublishPostToTwitter(ctx context.Context, in *proto.PublishPostToTwitterRequest) (*proto.PublishPostToTwitterResponse, error) {
	s.twitterPostMutex.Lock()
	defer s.twitterPostMutex.Unlock()

	post, err := s.postForTwitter(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if id := post.GetMetas().GetTwitterPostId(); id != "" {
		return twitterPostResponse(id), nil
	}
	id, err := s.twitterPostPublisher.Publish(ctx, post)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("发布到 Twitter 失败：%v", err))
	}
	if post.Metas == nil {
		post.Metas = &proto.Metas{}
	}
	post.Metas.TwitterPostId = id
	if _, err := s.UpdatePost(ctx, &proto.UpdatePostRequest{
		Post:       post,
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"metas"}},
		DoNotTouch: true,
	}); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("保存 Twitter Post ID 失败：%v", err))
	}
	return twitterPostResponse(id), nil
}

func twitterPostResponse(id string) *proto.PublishPostToTwitterResponse {
	return &proto.PublishPostToTwitterResponse{
		TwitterPostId: id,
		Url:           "https://twitter.com/i/web/status/" + id,
	}
}
