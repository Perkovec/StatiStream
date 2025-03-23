package storage

import (
	"context"
	"io"
)

type VideoMeta struct {
	Filename string
}

type Storage interface {
	GetNextVideo() (io.ReadCloser, int64, *VideoMeta)
	UpdateFilesList(context.Context) error
	GetQueue() []string
	AddToQueue(key string)
	GetFilesList() []string
}
