package r2

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// TODO 引入 AWS-SDK-GO 使得二进制大了 10MB

type R2 struct {
	client     *s3.Client
	bucketName string
}

func New(accountID, accessKeyID, accessKeySecret, bucketName string) (*R2, error) {
	// https://developers.cloudflare.com/r2/examples/aws/aws-sdk-go/
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID, accessKeySecret, ``,
		)),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		endpoint := fmt.Sprintf(`https://%s.r2.cloudflarestorage.com`, accountID)
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &R2{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (r2 *R2) Upload(ctx context.Context, path string, r io.Reader) error {
	output, err := r2.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &r2.bucketName,
		Key:    &path,
		Body:   r,
	})
	if err != nil {
		return err
	}
	_ = output
	return nil
}
