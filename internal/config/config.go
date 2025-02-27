package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Platform string
type PickStrategy string
type SourceType string

const (
	PlatformTwitch Platform = "twitch"
)

const (
	PickStrategyRandom PickStrategy = "random"
)

const (
	SourceTypeDisk SourceType = "disk"
	SourceTypeS3   SourceType = "s3"
)

type ConfigStreamKeyFile struct {
	Twitch string `yaml:"twitch"`
}

type ConfigSource struct {
	Type          SourceType          `yaml:"type"`
	DirectoryPath string              `yaml:"directory_path"`
	Files         []string            `yaml:"files"`
	PickStrategy  PickStrategy        `yaml:"pick_strategy"`
	S3Bucket      string              `yaml:"s3bucket"`
	S3Endpoint    string              `yaml:"s3endpoint"`
	S3Credentials ConfigS3Credentials `yaml:"s3credentials"`
	S3Region      string              `yaml:"s3region"`
}

type ConfigBot struct {
	Token         string  `yaml:"token"`
	AcceptedUsers []int64 `yaml:"accepted_users"`
}

type ConfigS3Credentials struct {
	ID     string `yaml:"id"`
	Secret string `yaml:"secret"`
}

type Config struct {
	StreamKeyFile ConfigStreamKeyFile `yaml:"stream_key_file"`
	Platform      []Platform          `yaml:"platforms"`
	Source        ConfigSource        `yaml:"source"`
	Bot           ConfigBot           `yaml:"bot"`
	FfmpegPath    string              `yaml:"ffmpeg_path"`
}

func ParseConfigFromFile(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ParseConfigFromFile.ReadFile: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return nil, fmt.Errorf("ParseConfigFromFile.Unmarshal: %w", err)
	}

	err = validateConfig(&config)
	if err != nil {
		return nil, fmt.Errorf("ParseConfigFromFile.validateConfig: %w", err)
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	// Проверяем что указаны платформы для стриминга
	if len(config.Platform) == 0 {
		return errors.New("empty 'platforms' list")
	}

	// Проверяем что по указанным платформам есть пути ключа трансляции
	for _, platform := range config.Platform {
		var keyPath string
		switch platform {
		case PlatformTwitch:
			keyPath = config.StreamKeyFile.Twitch
		}

		if len(keyPath) == 0 {
			return fmt.Errorf("empty stream key for platform: %s", platform)
		}
	}

	// Проверяем что указан источник видео
	if !isValidSourceType(config.Source.Type) {
		return fmt.Errorf("invalid source type: %s", config.Source.Type)
	}

	// Проверяем что указана папка с видео или список файлов
	if len(config.Source.DirectoryPath) == 0 && len(config.Source.Files) == 0 {
		return fmt.Errorf("not specified source directory path or files list")
	}

	// Проверяем что есть ключ бота
	if len(config.Bot.Token) == 0 {
		return errors.New("telegram bot token not specified")
	}

	// Проверяем что для бота указаны одобренные пользователи
	if len(config.Bot.AcceptedUsers) == 0 {
		return errors.New("list of accepted users for telegram bot is empty")
	}

	return nil
}

func isValidSourceType(rawSource SourceType) bool {
	return rawSource == SourceTypeDisk || rawSource == SourceTypeS3
}
