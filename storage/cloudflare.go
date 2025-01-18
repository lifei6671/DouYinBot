package storage

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
)

type Cloudflare struct {
	opts   *Options
	client *s3.Client
}

func (c *Cloudflare) OpenFile(ctx context.Context, filename string) (*File, error) {
	input := &s3.GetObjectInput{}
	object, err := c.client.GetObject(ctx, input)
	if err != nil {
		return nil, err
	}
	return &File{
		ContentLength: object.ContentLength,
		ContentType:   object.ContentType,
		Body:          object.Body,
	}, nil
}

func (c *Cloudflare) Delete(ctx context.Context, filename string) error {

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.opts.BucketName),
		Key:    aws.String(filename),
	}
	_, err := c.client.DeleteObject(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cloudflare) WriteFile(ctx context.Context, r io.Reader, filename string) (string, error) {
	mimeBuf := &bytes.Buffer{}
	inputBuf := &bytes.Buffer{}

	w := io.MultiWriter(mimeBuf, inputBuf)

	if _, err := io.Copy(w, r); err != nil {
		return "", fmt.Errorf("failed to copy file to cloudflare: %w", err)
	}
	mimeType, err := mimetype.DetectReader(mimeBuf)
	if err != nil {
		return "", fmt.Errorf("detect file %s error: %w", filename, err)
	}

	putObjectInput := &s3.PutObjectInput{
		Bucket:      &c.opts.BucketName,
		Key:         aws.String(filename),
		Body:        inputBuf,
		ContentType: aws.String(mimeType.String()),
	}

	_, pErr := c.client.PutObject(ctx, putObjectInput)
	if pErr != nil {
		return "", fmt.Errorf("upload file %s error: %w", filename, pErr)
	}
	return c.opts.Domain + strings.ReplaceAll("/"+filename, "//", "/"), nil
}

func NewCloudflare(opts ...OptionsFunc) (Storage, error) {
	var o Options
	for _, opt := range opts {
		if err := opt(&o); err != nil {
			return nil, err
		}
	}
	// 自定义 HTTP 客户端以支持所需的 TLS 配置
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS11,
		},
		Proxy: http.ProxyFromEnvironment,
	}
	customHTTPClient := &http.Client{
		Transport: customTransport,
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(o.AccessKeyID, o.AccessKeySecret, "")),
		config.WithRegion("auto"),
		config.WithHTTPClient(customHTTPClient),
		config.WithRetryMaxAttempts(5),
	)
	if err != nil {
		return nil, fmt.Errorf("load s3 config err:%w", err)
	}
	endpoint := o.Endpoint
	if o.Endpoint == "" {
		o.Endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", o.AccountID)
		endpoint = o.Endpoint
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})
	return &Cloudflare{
		opts:   &o,
		client: client,
	}, nil
}
