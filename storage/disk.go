package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type DiskStorage struct{}

func (d *DiskStorage) OpenFile(_ context.Context, filename string) (*File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	stat, sErr := f.Stat()
	if sErr != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	return &File{
		ContentLength: aws.Int64(stat.Size()),
		Body:          f,
	}, nil
}

func (d *DiskStorage) Delete(_ context.Context, filename string) error {
	return os.Remove(filename)
}

func (d *DiskStorage) WriteFile(ctx context.Context, r io.Reader, filename string) (string, error) {
	f, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return filename, nil
}

func NewDiskStorage() Storage {
	return &DiskStorage{}
}
