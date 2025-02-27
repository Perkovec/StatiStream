package storage

import "io"

type DiskStorageParams struct {
}

type diskStorage struct {
}

func NewDiskStorage(params DiskStorageParams) Storage {
	return &diskStorage{}
}

func (s *diskStorage) GetNextVideo() io.ReadCloser {
	return nil
}

func (s *diskStorage) UpdateFilesList() error {
	return nil
}
