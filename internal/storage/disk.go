package storage

import "io"

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

func (s *diskStorage) UpdateFilesList() error {
	return nil
}
