package r2

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/movsb/taoblog/modules/utils"
)

// TODO 引入 AWS-SDK-GO 使得二进制大了 10MB

type R2 struct {
	client     *s3.Client
	presign    *s3.PresignClient
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
		presign:    s3.NewPresignClient(client),
		bucketName: bucketName,
	}, nil
}

func (r2 *R2) Upload(ctx context.Context, path string, r io.Reader, contentType string) error {
	contentType = utils.IIF(contentType == "", `application/octet-stream`, contentType)
	output, err := r2.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &r2.bucketName,
		Key:         &path,
		Body:        r,
		ContentType: &contentType,
	})
	if err != nil {
		return err
	}
	_ = output
	log.Println(*output.ETag)
	return nil
}

func (r2 *R2) GetFileURL(ctx context.Context, path string, md5 string) string {
	_, err := r2.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:  &r2.bucketName,
		Key:     &path,
		IfMatch: &md5,
	})
	if err != nil {
		log.Println(err)
		return ``
	}
	output, err := r2.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &r2.bucketName,
		Key:    &path,
	}, s3.WithPresignExpires(time.Minute*15))
	if err != nil {
		log.Println(err)
		return ``
	}
	return output.URL
}
