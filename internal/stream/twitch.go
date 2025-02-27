package stream

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/Perkovec/StatiStream/internal/config"
	"github.com/Perkovec/StatiStream/internal/helpers"
)

const (
	DefaultTwitchEndpoint = "rtmp://ingest.global-contribute.live-video.net/app/"
)

type twitchStream struct {
	ffmpegPath  string
	token       string
	nextVideoCb func()

	streamProcess      *exec.Cmd
	streamProcessStdin io.WriteCloser
	stopCh             chan struct{}
	nextCh             chan struct{}

	ctx    context.Context
	cancel context.CancelFunc
}

type TwitchStreamParams struct {
	FfmpegPath string
}

func NewTwitchStream(params TwitchStreamParams) Stream {
	return &twitchStream{
		ffmpegPath: params.FfmpegPath,
		stopCh:     make(chan struct{}),
		nextCh:     make(chan struct{}),
	}
}

func (s *twitchStream) GetPlatform() config.Platform {
	return config.PlatformTwitch
}

func (s *twitchStream) HasToken() bool {
	return s.token != ""
}

func (s *twitchStream) SetStreamToken(token string) {
	s.token = token
}

func (s *twitchStream) IsStarted() bool {
	return s.streamProcess != nil
}

func (s *twitchStream) Stop() error {
	if !s.IsStarted() {
		return nil
	}

	if s.cancel != nil {
		s.cancel()
	}

	return s.streamProcess.Process.Kill()
}

func (s *twitchStream) SetVideo(video io.ReadCloser) {
	if !s.IsStarted() {
		return
	}

	videoCtx := helpers.NewReader(s.ctx, video, s.nextCh)

	go io.Copy(s.streamProcessStdin, videoCtx)
}

func (s *twitchStream) Start() error {
	if s.IsStarted() {
		return nil
	}

	var command = []string{
		"-loglevel", "warning", // only log warnings
		"-hide_banner", // don't bother echoing out the codecs and build information
		"-re",          // do this in real time
		"-i", "pipe:0", // read from stdin
		"-c", "copy", // don't actually encode
		"-f", "flv", // output format
		"-flvflags", "no_duration_filesize", // don't complain about not being
		DefaultTwitchEndpoint + s.token,
	}

	r := exec.Command(s.ffmpegPath, command...)

	stdin, err := r.StdinPipe()
	if err != nil {
		return fmt.Errorf("TwitchStream.Start.StdinPipe: %w", err)
	}

	stderr, err := r.StderrPipe()
	if err != nil {
		return fmt.Errorf("TwitchStream.Start.StderrPipe: %w", err)
	}
	if err = r.Start(); err != nil {
		return fmt.Errorf("TwitchStream.Start.Start: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel

	go s.captureOutput(ctx, stderr)
	go s.eofCapture(ctx)

	s.streamProcess = r
	s.streamProcessStdin = stdin

	return nil
}

func (s *twitchStream) NextVideo() <-chan struct{} {
	c := make(chan struct{})

	go func() {
		defer close(c)

		c <- <-s.nextCh
	}()

	return c
}

func (s *twitchStream) eofCapture(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.nextCh:
			s.nextVideoCb()
		}
	}
}

func (s *twitchStream) captureOutput(ctx context.Context, r io.Reader) {
	reader := bufio.NewReader(r)
	var line string
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, err = reader.ReadString('\n')
			if err != nil && err != io.EOF {
				fmt.Println("unable to read ffmpeg output")
				return
			}
			line = strings.TrimSpace(line)
			if line != "" {
				fmt.Printf("ffmpeg output: %s\n", line)
			}
			if err != nil {
				return
			}
		}
	}
}
