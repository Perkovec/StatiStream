package storage

import (
	"fmt"
	"io"
	"strings"

	"math/rand"

	"github.com/Perkovec/StatiStream/internal/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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
}

func boolPrt(value bool) *bool {
	return &value
}

func NewS3Storage(params S3StorageParams) (Storage, error) {
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
	}

	if len(params.Files) == 0 {
		err := st.UpdateFilesList()
		if err != nil {
			return nil, fmt.Errorf("S3Storage.UpdateFilesList: %w", err)
		}
	}

	return st, nil
}

func (s *s3Storage) GetNextVideo() io.ReadCloser {
	var key string
	switch s.pickStrategy {
	case config.PickStrategyRandom:
		key = s.getRandomVideoKey()
	default:
		return nil
	}

	res, err := s.s3Service.GetObject(&s3.GetObjectInput{
		Key:    &key,
		Bucket: &s.bucket,
	})
	if err != nil {
		return nil
	}
	return res.Body
}

func (s *s3Storage) getRandomVideoKey() string {
	if len(s.filesList) == 1 {
		return s.filesList[0]
	}
	return s.filesList[rand.Intn(len(s.filesList)-1)]
}

func (s *s3Storage) UpdateFilesList() error {
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
		fmt.Println(*object.Key)
		if strings.HasSuffix(*object.Key, ".flv") {
			videoList = append(videoList, *object.Key)
		}
	}

	s.filesList = videoList

	return nil
}
