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
	QueueCallbackPrefix      = "queue"
	AddVideoQueueCallback    = QueueCallbackPrefix + "_add"
	SelectVideoQueueCallback = QueueCallbackPrefix + "_select"
	VideosPageQueueCallback  = QueueCallbackPrefix + "_page"
)

var (
	addVideoQueueKeyboard = models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "Добавить видео",
					CallbackData: AddVideoQueueCallback,
				},
			},
		},
	}
)

func (s *streamBot) handleQueue(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	logger := zerolog.Ctx(ctx)

	if slices.Contains(s.AcceptedUsers, update.Message.From.ID) {
		logger.Info().
			Int64("user", update.Message.From.ID).
			Msgf("Handle videos queue")

		currentQueue := s.VideoStorage.GetQueue()

		if len(currentQueue) == 0 {
			b.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "На данный момент очередь пуста, следующие видео будут выбираться случайным образом",
				ReplyMarkup: addVideoQueueKeyboard,
			})
			return
		}

		text := "Очередь для транслирования:\n\n"

		for _, item := range currentQueue {
			text += fmt.Sprintf("- %s\n", item)
		}

		b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        text,
			ParseMode:   models.ParseModeMarkdown,
			ReplyMarkup: addVideoQueueKeyboard,
		})
	}
}

func (s *streamBot) handleAddVideoQueue(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
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

	if slices.Contains(s.AcceptedUsers, update.CallbackQuery.From.ID) {
		if update.CallbackQuery.Data == AddVideoQueueCallback {
			logger.Info().
				Int64("user", update.CallbackQuery.From.ID).
				Msgf("Handle add video queue")
			keyboard := models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{},
			}

			files := s.VideoStorage.GetFilesList()
			totalFilesCount := len(files)

			files = files[:min(5, len(files))]
			for _, file := range files {
				keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []models.InlineKeyboardButton{
					{
						Text:         file,
						CallbackData: fmt.Sprintf("%s:%s", SelectVideoQueueCallback, file),
					},
				})
			}

			if totalFilesCount > 5 {
				keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []models.InlineKeyboardButton{
					{
						Text:         "Следующая ▶️",
						CallbackData: fmt.Sprintf("%s:%d", VideosPageQueueCallback, 2),
					},
				})
			}

			b.EditMessageReplyMarkup(ctx, &telegramBot.EditMessageReplyMarkupParams{
				ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
				MessageID:   update.CallbackQuery.Message.Message.ID,
				ReplyMarkup: keyboard,
			})
		} else if strings.HasPrefix(update.CallbackQuery.Data, SelectVideoQueueCallback) {
			logger.Info().
				Int64("user", update.CallbackQuery.From.ID).
				Msgf("Handle select video to queue")

			parts := strings.Split(update.CallbackQuery.Data, ":")
			if len(parts) != 2 {
				logger.Warn().Msgf("Invalid callback data: %s", update.CallbackQuery.Data)
				return
			}

			s.VideoStorage.AddToQueue(parts[1])

			currentQueue := s.VideoStorage.GetQueue()

			if len(currentQueue) == 0 {
				b.EditMessageText(ctx, &telegramBot.EditMessageTextParams{
					ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
					MessageID:   update.CallbackQuery.Message.Message.ID,
					Text:        "На данный момент очередь пуста, следующие видео будут выбираться случайным образом",
					ReplyMarkup: addVideoQueueKeyboard,
				})
				return
			}

			text := "Очередь для транслирования:\n\n"

			for _, item := range currentQueue {
				text += fmt.Sprintf("- %s\n", item)
			}

			fmt.Println(update.CallbackQuery.Message.Message.Chat.ID, update.CallbackQuery.Message.Message.ID, text, models.ParseModeMarkdown, addVideoQueueKeyboard)

			b.EditMessageText(ctx, &telegramBot.EditMessageTextParams{
				ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
				MessageID:   update.CallbackQuery.Message.Message.ID,
				Text:        text,
				ReplyMarkup: addVideoQueueKeyboard,
			})
		}
	}
}
