package storage

import "io"

type VideoMeta struct {
	Filename string
}

type Storage interface {
	GetNextVideo() (io.ReadCloser, int64, *VideoMeta)
	UpdateFilesList() error
}
