package bot

import (
	"context"
	"fmt"
	"slices"
	"strings"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog"
)

const (
	StartStreamCallbackPrefix  = "start_stream"
	ApproveStartStreamCallback = StartStreamCallbackPrefix + "_approve"
	CancelStartStreamCallback  = StartStreamCallbackPrefix + "_cancel"
)

func (s *streamBot) preStartStream(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	logger := zerolog.Ctx(ctx)

	if slices.Contains(s.AcceptedUsers, update.Message.From.ID) {
		logger.Info().
			Int64("user", update.Message.From.ID).
			Msgf("Prehandle start stream")

		alreadyStartedPlatforms := make([]string, 0, len(s.Streams))
		for platform, stream := range s.Streams {
			if stream.IsStarted() {
				alreadyStartedPlatforms = append(alreadyStartedPlatforms, string(platform))
			}
		}

		if len(alreadyStartedPlatforms) > 0 {
			b.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("На следующих платформах уже запущена трансляция: %s\nСначала остановите все трансляции", strings.Join(alreadyStartedPlatforms, ", ")),
			})
			return
		}

		missedTokenPlatforms := make([]string, 0, len(s.Streams))
		for platform, stream := range s.Streams {
			if !stream.HasToken() {
				missedTokenPlatforms = append(missedTokenPlatforms, string(platform))
			}
		}

		if len(missedTokenPlatforms) > 0 {
			b.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("Не добавлены ключи трансляций для платформ: %s", strings.Join(missedTokenPlatforms, ", ")),
			})
			return
		}

		keyboard := models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         "Подтвердить",
						CallbackData: ApproveStartStreamCallback,
					},
					{
						Text:         "Отмена",
						CallbackData: CancelStartStreamCallback,
					},
				},
			},
		}

		b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Вы точно уверены, что хотите запустить стрим?",
			ReplyMarkup: keyboard,
		})
	}
}

func (s *streamBot) handleStartStream(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	logger := zerolog.Ctx(ctx)

	isAccepted := slices.Contains(s.AcceptedUsers, update.CallbackQuery.From.ID)
	cbAnswerParams := &telegramBot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	}
	if !isAccepted {
		cbAnswerParams.ShowAlert = true
		cbAnswerParams.Text = "Вы не можете управлять трансляцией"
	}

	b.AnswerCallbackQuery(ctx, cbAnswerParams)

	if isAccepted {
		if update.CallbackQuery.Message.Message != nil {
			logger.Info().
				Int64("user", update.CallbackQuery.From.ID).
				Msgf("Handle start stream")

			var editText string
			if update.CallbackQuery.Data == CancelStartStreamCallback {
				editText = "Запуск стрима отменен"
			}

			if update.CallbackQuery.Data == ApproveStartStreamCallback {
				video, contentLength, videoMeta := s.VideoStorage.GetNextVideo()
				if video == nil {
					editText = "Не удалось получить видео для запуска стрима"
				} else {
					err := s.Streams.Start()
					if err != nil {
						editText = fmt.Sprintf("Не удалось запустить стрим:\n%v", err)
					} else {
						editText = "Стрим запущен"
						fmt.Printf("Start video \"%s\": %d\n", videoMeta.Filename, contentLength)
						s.Streams.SetVideo(video, contentLength)
					}
				}
			}
			b.EditMessageText(ctx, &telegramBot.EditMessageTextParams{
				ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
				MessageID: update.CallbackQuery.Message.Message.ID,
				Text:      editText,
				ReplyMarkup: models.InlineKeyboardMarkup{
					InlineKeyboard: [][]models.InlineKeyboardButton{},
				},
			})
		}
	}
}
