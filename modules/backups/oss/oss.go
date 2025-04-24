package oss

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	cc "github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/utils"

	alioss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	aliosscrendentials "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// TODO 引入 AWS-SDK-GO 使得二进制大了 10MB

type Client interface {
	Upload(ctx context.Context, path string, size int64, r io.Reader, contentType string, digest []byte) error
	GetFileURL(ctx context.Context, path string, di []byte) string
}

func New(provider string, c *cc.OSSConfig) (Client, error) {
	switch provider {
	case `r2`:
		return NewR2(c)
	case `aliyun`:
		return NewAliyun(c)
	default:
		return nil, fmt.Errorf(`未知存储服务：%s`, provider)
	}
}

func digest2contentMD5(digest []byte) string {
	return base64.StdEncoding.EncodeToString(digest)
}

type S3Compatible struct {
	client     *s3.Client
	presign    *s3.PresignClient
	bucketName string
}

func NewR2(c *cc.OSSConfig) (Client, error) {
	// https://developers.cloudflare.com/r2/examples/aws/aws-sdk-go/
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			c.AccessKeyID, c.AccessKeySecret, ``,
		)),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(c.Endpoint)
	})
	return &S3Compatible{
		client:     client,
		presign:    s3.NewPresignClient(client),
		bucketName: c.BucketName,
	}, nil
}

func (oss *S3Compatible) Upload(ctx context.Context, path string, size int64, r io.Reader, contentType string, digest []byte) error {
	contentType = utils.IIF(contentType == "", `application/octet-stream`, contentType)
	output, err := oss.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &oss.bucketName,
		Key:           &path,
		Body:          r,
		ContentType:   &contentType,
		ContentLength: &size,
		ContentMD5:    aws.String(digest2contentMD5(digest)),
	})
	if err != nil {
		return err
	}
	_ = output
	// log.Println(*output.ETag)
	return nil
}

func (oss *S3Compatible) GetFileURL(ctx context.Context, path string, md5 []byte) string {
	_, err := oss.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:  &oss.bucketName,
		Key:     &path,
		IfMatch: aws.String(digest2contentMD5(md5)),
	})
	if err != nil {
		// log.Println(err)
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

type Aliyun struct {
	client     *alioss.Client
	bucketName string
}

var _ Client = (*Aliyun)(nil)

func NewAliyun(c *cc.OSSConfig) (Client, error) {
	if c.Endpoint != `` {
		return nil, fmt.Errorf(`阿里云不能传 endpoint`)
	}
	// Using the SDK's default configuration
	// loading credentials values from the environment variables
	cfg := alioss.LoadDefaultConfig().
		WithCredentialsProvider(aliosscrendentials.NewStaticCredentialsProvider(c.AccessKeyID, c.AccessKeySecret)).
		WithRegion(c.Region)

	client := alioss.NewClient(cfg)
	return &Aliyun{
		client:     client,
		bucketName: c.BucketName,
	}, nil
}

func (oss *Aliyun) Upload(ctx context.Context, path string, size int64, r io.Reader, contentType string, digest []byte) error {
	contentType = utils.IIF(contentType == "", `application/octet-stream`, contentType)
	output, err := oss.client.PutObject(ctx, &alioss.PutObjectRequest{
		Bucket:        &oss.bucketName,
		Key:           &path,
		Body:          r,
		ContentLength: &size,
		ContentMD5:    alioss.Ptr(digest2contentMD5(digest)),
		ContentType:   &contentType,
	})
	if err != nil {
		return err
	}
	_ = output
	return nil
}

func (oss *Aliyun) GetFileURL(ctx context.Context, path string, md5 []byte) string {
	output1, err := oss.client.HeadObject(ctx, &alioss.HeadObjectRequest{
		Bucket:  &oss.bucketName,
		Key:     &path,
		IfMatch: alioss.Ptr(digest2contentMD5(md5)),
	})
	if err != nil {
		return ``
	}
	_ = output1

	output, err := oss.client.Presign(ctx, &alioss.GetObjectRequest{
		Bucket:  &oss.bucketName,
		Key:     &path,
		IfMatch: alioss.Ptr(digest2contentMD5(md5)),
	})
	if err != nil {
		log.Println(err)
		return ``
	}
	return output.URL
}
