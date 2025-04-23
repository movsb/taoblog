package oss

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	cc "github.com/movsb/taoblog/cmd/config"
)

// Cloudflare R2
func NewR2(c *cc.OSSConfig) (*s3.Client, error) {
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
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(c.Endpoint)
	}), nil
}

// 腾讯云 COS
func NewCOS(c *cc.OSSConfig) (*s3.Client, error) {
	// https://cloud.tencent.com/document/product/436/37421
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(`auto`),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			c.AccessKeyID, c.AccessKeySecret, ``,
		)),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(c.Endpoint)
	}), nil
}
