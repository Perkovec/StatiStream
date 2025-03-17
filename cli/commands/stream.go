package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Perkovec/StatiStream/internal/bot"
	"github.com/Perkovec/StatiStream/internal/config"
	"github.com/Perkovec/StatiStream/internal/storage"
	"github.com/Perkovec/StatiStream/internal/stream"
	telegramBot "github.com/go-telegram/bot"
	"github.com/hashicorp/cli"
)

type StreamCommand struct {
}

func (f *CommandsFactory) NewStreamCommand() (cli.Command, error) {
	return &StreamCommand{}, nil
}

func (c *StreamCommand) Help() string {
	return ""
}

func (c *StreamCommand) Synopsis() string {
	return ""
}

func (c *StreamCommand) Run(args []string) int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := c.getConfig(args)
	if err != nil {
		log.Fatal(err)
	}

	videoStorage, err := c.initStorage(cfg.Source)
	if err != nil {
		log.Fatal(err)
	}

	streams := c.initStreams(cfg)

	bot, err := c.initTelegramBot(
		cfg.Bot,
		videoStorage,
		streams,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting bot...")
	bot.Start(ctx)

	return 0
}

func (c *StreamCommand) getConfig(args []string) (*config.Config, error) {
	configPath := "./config.yaml"
	if len(args) > 0 {
		configPath = args[0]
	}

	return config.ParseConfigFromFile(configPath)
}

func (c *StreamCommand) initTelegramBot(cfg config.ConfigBot, storage storage.Storage, streams stream.Streams) (*telegramBot.Bot, error) {
	botToken, err := os.ReadFile(cfg.Token)
	if err != nil {
		log.Fatal(err)
	}

	token := strings.ReplaceAll(string(botToken), "\n", "")
	token = strings.TrimSpace(token)

	return bot.NewBot(bot.BotParams{
		AcceptedUsers: cfg.AcceptedUsers,
		Token:         token,
		VideoStorage:  storage,
		Streams:       streams,
	})
}

func (c *StreamCommand) initStorage(cfg config.ConfigSource) (storage.Storage, error) {
	switch cfg.Type {
	case config.SourceTypeS3:
		return storage.NewS3Storage(storage.S3StorageParams{
			Bucket:            cfg.S3Bucket,
			Endpoint:          cfg.S3Endpoint,
			CredentialsID:     cfg.S3Credentials.ID,
			CredentialsSecret: cfg.S3Credentials.Secret,
			Region:            cfg.S3Region,
			PickStrategy:      cfg.PickStrategy,
			DirectoryPath:     cfg.DirectoryPath,
			Files:             cfg.Files,
		})
	case config.SourceTypeDisk:
		return storage.NewDiskStorage(storage.DiskStorageParams{}), nil
	default:
		return nil, fmt.Errorf("unknown storage type '%s'", cfg.Type)
	}
}

func (c *StreamCommand) initStreams(cfg *config.Config) stream.Streams {
	streams := make(stream.Streams, len(cfg.Platform))
	for _, platform := range cfg.Platform {
		switch platform {
		case config.PlatformTwitch:
			streams[platform] = stream.NewTwitchStream(stream.TwitchStreamParams{
				FfmpegPath: cfg.FfmpegPath,
			})
		}
	}

	return streams
}
