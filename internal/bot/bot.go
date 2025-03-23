package bot

import (
	"context"

	"github.com/Perkovec/StatiStream/internal/storage"
	"github.com/Perkovec/StatiStream/internal/stream"
	telegramBot "github.com/go-telegram/bot"
	"github.com/rs/zerolog"
)

type ButtonType string

const (
	ButtonTypeNext       ButtonType = "➡️ Переключить видео"
	ButtonTypeReloadList ButtonType = "🔄 Перезагрузить данные"
	ButtonTypeStatistics ButtonType = "📊 Статистика"
	ButtonTypeStart      ButtonType = "🟢 Запустить"
	ButtonTypeStop       ButtonType = "⛔️ Остановить"
	ButtonTypeQueue      ButtonType = "📄 Очередь"
)

type streamBot struct {
	AcceptedUsers []int64
	VideoStorage  storage.Storage
	Streams       stream.Streams

	StreamTokensMap map[string]string
}

type BotParams struct {
	AcceptedUsers []int64
	Token         string
	VideoStorage  storage.Storage
	Streams       stream.Streams
}

func NewBot(ctx context.Context, cfg BotParams) (*telegramBot.Bot, error) {
	streamBot := &streamBot{
		AcceptedUsers:   cfg.AcceptedUsers,
		VideoStorage:    cfg.VideoStorage,
		Streams:         cfg.Streams,
		StreamTokensMap: map[string]string{},
	}

	opts := []telegramBot.Option{
		telegramBot.WithDefaultHandler(streamBot.handleInline),
	}

	b, err := telegramBot.New(cfg.Token, opts...)
	if err != nil {
		return nil, err
	}

	b.RegisterHandler(telegramBot.HandlerTypeMessageText, "/start", telegramBot.MatchTypeExact, streamBot.handleStart)
	b.RegisterHandler(telegramBot.HandlerTypeMessageText, string(ButtonTypeReloadList), telegramBot.MatchTypeExact, streamBot.handleReloadStorage)
	b.RegisterHandler(telegramBot.HandlerTypeMessageText, string(ButtonTypeNext), telegramBot.MatchTypeExact, streamBot.preHandleNextVideo)
	b.RegisterHandler(telegramBot.HandlerTypeMessageText, string(ButtonTypeStart), telegramBot.MatchTypeExact, streamBot.preStartStream)
	b.RegisterHandler(telegramBot.HandlerTypeMessageText, string(ButtonTypeStop), telegramBot.MatchTypeExact, streamBot.preStopStream)
	b.RegisterHandler(telegramBot.HandlerTypeMessageText, string(ButtonTypeQueue), telegramBot.MatchTypeExact, streamBot.handleQueue)
	b.RegisterHandler(telegramBot.HandlerTypeMessageText, "stream_key:", telegramBot.MatchTypePrefix, streamBot.handleSetStreamKey)

	b.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData, NextVideoCallbackPrefix, telegramBot.MatchTypePrefix, streamBot.handleNextVideo)
	b.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData, StartStreamCallbackPrefix, telegramBot.MatchTypePrefix, streamBot.handleStartStream)
	b.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData, QueueCallbackPrefix, telegramBot.MatchTypePrefix, streamBot.handleAddVideoQueue)

	go streamBot.captureNextVideo(ctx)

	return b, nil
}

func (s *streamBot) captureNextVideo(ctx context.Context) {
	logger := zerolog.Ctx(ctx)

	for {
	out:
		for _, stream := range s.Streams {
			<-stream.NextVideo()
			video, contentLength, videoMeta := s.VideoStorage.GetNextVideo()
			logger.Info().Msgf("Start video \"%s\": %d\n", videoMeta.Filename, contentLength)
			stream.SetVideo(video, contentLength)
			break out
		}
	}
}
