package storage

import (
	"context"
	"fmt"
	"io"
)

// File 一个文件
type File struct {
	ContentLength *int64
	ContentType   *string
	Body          io.ReadCloser
}

type Storage interface {
	// OpenFile 打开文件
	OpenFile(ctx context.Context, filename string) (*File, error)

	// Delete 删除一个文件
	Delete(ctx context.Context, filename string) error

	// WriteFile 写入一个文件
	WriteFile(ctx context.Context, r io.Reader, filename string) (string, error)
}

func Factory(name string, opts ...OptionsFunc) (Storage, error) {
	switch name {
	case "local":
		return NewDiskStorage(), nil
	case "cloudflare":
		return NewCloudflare(opts...)
	}
	return nil, fmt.Errorf("unknown storage: %s", name)
}
