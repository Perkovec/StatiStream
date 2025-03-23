package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Perkovec/StatiStream/internal/bot"
	"github.com/Perkovec/StatiStream/internal/config"
	"github.com/Perkovec/StatiStream/internal/storage"
	"github.com/Perkovec/StatiStream/internal/stream"
	telegramBot "github.com/go-telegram/bot"
	"github.com/hashicorp/cli"
	"github.com/rs/zerolog"
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

	ctx, logger, err := initLogger(ctx)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := c.getConfig(ctx, args)
	if err != nil {
		log.Fatal(err)
	}

	videoStorage, err := c.initStorage(ctx, cfg.Source)
	if err != nil {
		log.Fatal(err)
	}

	streams := c.initStreams(ctx, cfg)

	bot, err := c.initTelegramBot(
		ctx,
		cfg.Bot,
		videoStorage,
		streams,
	)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info().Msg("Starting bot")
	bot.Start(ctx)

	return 0
}

func (c *StreamCommand) getConfig(ctx context.Context, args []string) (*config.Config, error) {
	logger := zerolog.Ctx(ctx)
	configPath := "./config.yaml"
	if len(args) > 0 {
		configPath = args[0]
	}

	logger.Info().Msgf("Parse config file: %s", configPath)

	return config.ParseConfigFromFile(configPath)
}

func (c *StreamCommand) initTelegramBot(ctx context.Context, cfg config.ConfigBot, storage storage.Storage, streams stream.Streams) (*telegramBot.Bot, error) {
	botToken, err := os.ReadFile(cfg.Token)
	if err != nil {
		log.Fatal(err)
	}

	token := strings.ReplaceAll(string(botToken), "\n", "")
	token = strings.TrimSpace(token)

	return bot.NewBot(ctx, bot.BotParams{
		AcceptedUsers: cfg.AcceptedUsers,
		Token:         token,
		VideoStorage:  storage,
		Streams:       streams,
	})
}

func (c *StreamCommand) initStorage(ctx context.Context, cfg config.ConfigSource) (storage.Storage, error) {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msgf("Init storage: %s", cfg.Type)

	switch cfg.Type {
	case config.SourceTypeS3:
		return storage.NewS3Storage(ctx, storage.S3StorageParams{
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

func (c *StreamCommand) initStreams(ctx context.Context, cfg *config.Config) stream.Streams {
	logger := zerolog.Ctx(ctx)
	streams := make(stream.Streams, len(cfg.Platform))
	for _, platform := range cfg.Platform {
		logger.Info().Msgf("Init stream manager: %s", platform)
		switch platform {
		case config.PlatformTwitch:
			streams[platform] = stream.NewTwitchStream(stream.TwitchStreamParams{
				FfmpegPath: cfg.FfmpegPath,
			})
		}
	}

	return streams
}

func initLogger(ctx context.Context) (context.Context, zerolog.Logger, error) {
	err := os.MkdirAll(filepath.Join(".", "logs"), os.ModePerm)
	if err != nil {
		return ctx, zerolog.Logger{}, err
	}

	logFile, _ := os.OpenFile(
		fmt.Sprintf("./logs/statistream_%s", time.Now().Format("2006-01-02")),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)

	multiLogOutput := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stdout}, logFile)
	logger := zerolog.New(multiLogOutput).With().Timestamp().Logger()
	ctx = logger.WithContext(ctx)

	return ctx, logger, nil
}
