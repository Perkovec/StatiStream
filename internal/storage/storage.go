package storage

import "io"

type Storage interface {
	GetNextVideo() io.ReadCloser
	UpdateFilesList() error
}
