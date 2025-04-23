package oss

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/utils"
)

// TODO 引入 AWS-SDK-GO 使得二进制大了 10MB

type OSS struct {
	client     *s3.Client
	presign    *s3.PresignClient
	bucketName string
}

func New(provider string, cfg *config.OSSConfig) (*OSS, error) {
	var err error
	var client *s3.Client

	switch provider {
	case `r2`:
		client, err = NewR2(cfg)
		if err != nil {
			return nil, err
		}
	case `cos`:
		client, err = NewCOS(cfg)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf(`unknown provider: %s`, provider)
	}
	return &OSS{
		client:     client,
		presign:    s3.NewPresignClient(client),
		bucketName: cfg.BucketName,
	}, nil
}

func (oss *OSS) Upload(ctx context.Context, path string, r io.Reader, contentType string) error {
	contentType = utils.IIF(contentType == "", `application/octet-stream`, contentType)
	output, err := oss.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &oss.bucketName,
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

func (oss *OSS) GetFileURL(ctx context.Context, path string, md5 string) string {
	_, err := oss.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:  &oss.bucketName,
		Key:     &path,
		IfMatch: &md5,
	})
	if err != nil {
		log.Println(err)
		return ``
	}
	output, err := oss.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &oss.bucketName,
		Key:    &path,
	}, s3.WithPresignExpires(time.Minute*15))
	if err != nil {
		log.Println(err)
		return ``
	}
	return output.URL
}
