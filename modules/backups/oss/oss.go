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
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	cc "github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/utils"

	alioss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	aliosscrendentials "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// TODO 引入 AWS-SDK-GO 使得二进制大了 10MB

type Client interface {
	Upload(ctx context.Context, path string, size int64, r io.Reader, contentType string, digest []byte) error
	GetFileURL(ctx context.Context, path string, digest []byte) string
	DeleteByPrefix(ctx context.Context, prefix string)
}

func New(provider string, c *cc.OSSConfig) (Client, error) {
	switch provider {
	case `r2`:
		return NewS3(c, false)
	case `aliyun`:
		return NewAliyun(c)
	case `minio`:
		return NewS3(c, true)
	default:
		return nil, fmt.Errorf(`未知存储服务：%s`, provider)
	}
}

type Digest []byte

func NewDigestFromString(s string) Digest {
	if len(s) != 32 {
		panic(`bad digest:` + s)
	}
	var b []byte
	if _, err := fmt.Sscanf(s, `%x`, &b); err != nil {
		panic(`bad digest:` + s)
	}
	if len(b) != 16 {
		panic(`bad digest:` + s)
	}
	return Digest(b)
}

func (d Digest) ToContentMD5() string {
	return base64.StdEncoding.EncodeToString([]byte(d))
}

func (d Digest) ToETag(upperCase bool) string {
	f := `"%x"`
	if upperCase {
		f = `"%X"`
	}
	return fmt.Sprintf(f, []byte(d))
}

type S3Compatible struct {
	client     *s3.Client
	presign    *s3.PresignClient
	bucketName string
}

func NewS3(c *cc.OSSConfig, pathStyle bool) (Client, error) {
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
		o.UsePathStyle = pathStyle
		// o.ClientLogMode = aws.LogRequest
		// o.Logger = logging.LoggerFunc(func(classification logging.Classification, format string, v ...interface{}) {
		// 	log.Println(classification, fmt.Sprintf(format, v...))
		// })
	})
	return &S3Compatible{
		client:     client,
		presign:    s3.NewPresignClient(client),
		bucketName: c.BucketName,
	}, nil
}

const privateCache = `private, no-cache, must-revalidate`

func (oss *S3Compatible) Upload(ctx context.Context, path string, size int64, r io.Reader, contentType string, digest []byte) error {
	// 先判断是否存在，避免重复上传。
	// NOTE：不使用 PutObject 的 IfNoneMatch，看文档写的是判断文件不存在才上传，
	// 无法区别文件变化。
	_, err := oss.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:  &oss.bucketName,
		Key:     &path,
		IfMatch: aws.String(Digest(digest).ToETag(false)),
	})
	if err == nil {
		log.Println(`oss.Upload:`, path, `already exists. Won't upload.`)
		return nil
	}
	contentType = utils.IIF(contentType == "", `application/octet-stream`, contentType)
	output, err := oss.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        &oss.bucketName,
		Key:           &path,
		Body:          r,
		ContentType:   &contentType,
		ContentLength: &size,
		ContentMD5:    aws.String(Digest(digest).ToContentMD5()),
		CacheControl:  aws.String(privateCache),
	})
	if err != nil {
		return err
	}
	log.Println(`oss.Upload: ETag:`, path, *output.ETag)
	return nil
}

func (oss *S3Compatible) GetFileURL(ctx context.Context, path string, digest []byte) string {
	headOutput, err := oss.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:  &oss.bucketName,
		Key:     &path,
		IfMatch: aws.String(Digest(digest).ToETag(false)),
	})
	if err != nil {
		// log.Println(err)
		return ``
	}
	_ = headOutput
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

func (oss *S3Compatible) DeleteByPrefix(ctx context.Context, prefix string) {
	var toDelete []types.ObjectIdentifier

	paginator := s3.NewListObjectsV2Paginator(oss.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(oss.bucketName),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			log.Println("ListObjectsV2 error:", err)
			return
		}
		for _, obj := range page.Contents {
			toDelete = append(toDelete, types.ObjectIdentifier{Key: obj.Key})
		}
	}

	if len(toDelete) == 0 {
		log.Println("No objects to delete.")
		return
	}

	// 批量删除会报缺少 ContentMD5，不知道怎么传。
	for _, del := range toDelete {
		_, err := oss.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(oss.bucketName),
			Key:    del.Key,
		})
		if err != nil {
			log.Println("DeleteObject error:", err)
		} else {
			log.Println(`DeleteObject:`, *del.Key)
		}
	}
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
	// 先判断是否存在，避免重复上传。
	_, err := oss.client.HeadObject(ctx, &alioss.HeadObjectRequest{
		Bucket:  &oss.bucketName,
		Key:     &path,
		IfMatch: alioss.Ptr(Digest(digest).ToETag(true)),
	})
	if err == nil {
		log.Println(`oss.Upload:`, path, `already exists. Won't upload.`)
		return nil
	}
	contentType = utils.IIF(contentType == "", `application/octet-stream`, contentType)
	output, err := oss.client.PutObject(ctx, &alioss.PutObjectRequest{
		Bucket:        &oss.bucketName,
		Key:           &path,
		Body:          r,
		ContentLength: &size,
		ContentMD5:    alioss.Ptr(Digest(digest).ToContentMD5()),
		ContentType:   &contentType,
		CacheControl:  alioss.Ptr(privateCache),
		ProgressFn: func(increment, transferred, total int64) {
			log.Printf(`oss.Upload: progress: %s (%.2f%%)`, path, float64(transferred)/float64(total)*100)
		},
	})
	if err != nil {
		return err
	}
	log.Println(`oss.Upload: Content-MD5 & ETag:`, path, *output.ContentMD5, *output.ETag)
	return nil
}

func (oss *Aliyun) GetFileURL(ctx context.Context, path string, digest []byte) string {
	output1, err := oss.client.HeadObject(ctx, &alioss.HeadObjectRequest{
		Bucket:  &oss.bucketName,
		Key:     &path,
		IfMatch: alioss.Ptr(Digest(digest).ToETag(true)),
	})
	if err != nil {
		return ``
	}
	_ = output1

	output, err := oss.client.Presign(ctx, &alioss.GetObjectRequest{
		Bucket:  &oss.bucketName,
		Key:     &path,
		IfMatch: alioss.Ptr(Digest(digest).ToETag(true)),
	})
	if err != nil {
		log.Println(err)
		return ``
	}

	return output.URL
}

func (oss *Aliyun) DeleteByPrefix(ctx context.Context, prefix string) {
	var toDelete []string

	listInput := &alioss.ListObjectsRequest{
		Bucket: &oss.bucketName,
		Prefix: &prefix,
	}

	for {
		resp, err := oss.client.ListObjects(ctx, listInput)
		if err != nil {
			log.Println("ListObjects error:", err)
			return
		}

		for _, obj := range resp.Contents {
			toDelete = append(toDelete, *obj.Key)
		}

		if !resp.IsTruncated || resp.NextMarker == nil {
			break
		}

		listInput.Marker = resp.NextMarker
	}

	if len(toDelete) == 0 {
		log.Println("No objects to delete.")
		return
	}

	_, err := oss.client.DeleteMultipleObjects(ctx, &alioss.DeleteMultipleObjectsRequest{
		Bucket: &oss.bucketName,
		Objects: utils.Map(toDelete, func(key string) alioss.DeleteObject {
			return alioss.DeleteObject{Key: &key}
		}),
	})
	if err != nil {
		log.Println("DeleteMultipleObjects error:", err)
	} else {
		log.Printf("Deleted %d objects with prefix %s\n", len(toDelete), prefix)
	}
}
