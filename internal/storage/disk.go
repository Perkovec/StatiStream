package storage

import (
	"context"
	"io"
)

type DiskStorageParams struct {
}

type diskStorage struct {
}

func NewDiskStorage(params DiskStorageParams) Storage {
	return &diskStorage{}
}

func (s *diskStorage) GetNextVideo() (io.ReadCloser, int64, *VideoMeta) {
	return nil, 0, nil
}

func (s *diskStorage) UpdateFilesList(_ context.Context) error {
	return nil
}

func (s *diskStorage) GetQueue() []string {
	return []string{}
}

func (s *diskStorage) AddToQueue(_ string) {

}

func (s *diskStorage) GetFilesList() []string {
	return []string{}
}
