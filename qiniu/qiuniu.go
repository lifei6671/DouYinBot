package qiniu

import (
	"bytes"
	"context"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

type Bucket struct {
	mac  *qbox.Mac
	Zone *storage.Zone
}

func NewBucket(accessKey, secretKey string) *Bucket {
	mac := qbox.NewMac(accessKey, secretKey)

	return &Bucket{
		mac:  mac,
		Zone: &storage.ZoneHuanan,
	}
}

func (b *Bucket) UploadFile(bucket string, key string, localFile string) error {
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	upToken := putPolicy.UploadToken(b.mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = b.Zone
	// 是否使用https域名
	cfg.UseHTTPS = true
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	// 构建表单上传的对象
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	// 可选配置
	putExtra := storage.PutExtra{
		Params: map[string]string{
			"x:name": key,
		},
	}
	err := formUploader.PutFile(context.Background(), &ret, upToken, key, localFile, &putExtra)
	if err != nil {
		return err
	}
	return nil
}
func (b *Bucket) Upload(bucket string, key string, data []byte) error {
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	upToken := putPolicy.UploadToken(b.mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = b.Zone
	// 是否使用https域名
	cfg.UseHTTPS = true
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{
		Params: map[string]string{
			"x:name": key,
		},
	}
	dataLen := int64(len(data))
	err := formUploader.Put(context.Background(), &ret, upToken, key, bytes.NewReader(data), dataLen, &putExtra)
	if err != nil {
		return err
	}
	return nil
}
