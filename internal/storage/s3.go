package storage

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"math/rand"

	"github.com/Perkovec/StatiStream/internal/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rs/zerolog"
)

type S3StorageParams struct {
	Bucket        string
	PickStrategy  config.PickStrategy
	DirectoryPath string
	Files         []string

	Endpoint          string
	CredentialsID     string
	CredentialsSecret string
	Region            string
}

type s3Storage struct {
	s3Service *s3.S3
	bucket    string

	pickStrategy  config.PickStrategy
	directoryPath string
	filesList     []string
	queue         []string
}

func boolPrt(value bool) *bool {
	return &value
}

func NewS3Storage(ctx context.Context, params S3StorageParams) (Storage, error) {
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         &params.Endpoint,
		Credentials:      credentials.NewStaticCredentials(params.CredentialsID, params.CredentialsSecret, ""),
		S3ForcePathStyle: boolPrt(true),
		Region:           &params.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("NewS3Storage.NewSession: %w", err)
	}

	s3Service := s3.New(sess)

	st := &s3Storage{
		s3Service:     s3Service,
		directoryPath: params.DirectoryPath,
		pickStrategy:  params.PickStrategy,
		bucket:        params.Bucket,
		queue:         []string{},
	}

	if len(params.Files) == 0 {
		err := st.UpdateFilesList(ctx)
		if err != nil {
			return nil, fmt.Errorf("S3Storage.UpdateFilesList: %w", err)
		}
	}

	return st, nil
}

func (s *s3Storage) GetNextVideo() (io.ReadCloser, int64, *VideoMeta) {
	var key string

	if len(s.queue) > 0 {
		key, s.queue = s.queue[0], s.queue[1:]
	} else {
		switch s.pickStrategy {
		case config.PickStrategyRandom:
			key = s.getRandomVideoKey()
		default:
			return nil, 0, nil
		}
	}

	res, err := s.s3Service.GetObject(&s3.GetObjectInput{
		Key:    &key,
		Bucket: &s.bucket,
	})
	if err != nil {
		return nil, 0, nil
	}

	var bodyLength int64 = 0
	if res.ContentLength != nil {
		bodyLength = *res.ContentLength
	}

	return res.Body, bodyLength, &VideoMeta{
		Filename: key,
	}
}

func (s *s3Storage) getRandomVideoKey() string {
	if len(s.filesList) == 1 {
		return s.filesList[0]
	}

	source := rand.NewSource(time.Now().Unix())
	r := rand.New(source)

	return s.filesList[r.Intn(len(s.filesList))]
}

func (s *s3Storage) UpdateFilesList(ctx context.Context) error {
	logger := zerolog.Ctx(ctx)

	directory := strings.TrimRight(s.directoryPath, "/")
	list, err := s.s3Service.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &s.bucket,
		Prefix: &directory,
	})
	if err != nil {
		return fmt.Errorf("NewS3Storage.UpdateFilesList.ListObjectsV2: %w", err)
	}

	videoList := make([]string, 0, len(list.Contents))
	for _, object := range list.Contents {
		if strings.HasSuffix(*object.Key, ".ts") {
			videoList = append(videoList, *object.Key)
		}
	}

	s.filesList = videoList

	logger.Info().
		Strs("files", videoList).
		Msg("Files list updated")

	return nil
}

func (s *s3Storage) GetQueue() []string {
	return s.queue
}

func (s *s3Storage) AddToQueue(key string) {
	if slices.Contains(s.filesList, key) {
		s.queue = append(s.queue, key)
	}
}

func (s *s3Storage) GetFilesList() []string {
	return s.filesList
}
