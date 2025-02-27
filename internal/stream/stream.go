package stream

import (
	"fmt"
	"io"

	"github.com/Perkovec/StatiStream/internal/config"
)

type Stream interface {
	Start() error
	Stop() error
	SetVideo(video io.ReadCloser)
	SetStreamToken(token string)
	HasToken() bool
	IsStarted() bool
	NextVideo() <-chan struct{}
}

type Streams map[config.Platform]Stream

func (s Streams) Start() error {
	for _, stream := range s {
		err := stream.Start()
		if err != nil {
			s.Stop()
			return fmt.Errorf("Streams.Start: %w", err)
		}
	}

	return nil
}

func (s Streams) Stop() error {
	var sErr error
	for _, stream := range s {
		err := stream.Stop()
		if err != nil {
			sErr = fmt.Errorf("Streams.Stop: %w", err)
		}
	}

	return sErr
}

func (s Streams) SetVideo(video io.ReadCloser) {
	for _, stream := range s {
		stream.SetVideo(video)
	}
}
