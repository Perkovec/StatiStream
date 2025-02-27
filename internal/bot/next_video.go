package bot

import (
	"context"
	"slices"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	NextVideoCallbackPrefix  = "next_video"
	ApproveNextVideoCallback = NextVideoCallbackPrefix + "_approve"
	CancelNextVideoCallback  = NextVideoCallbackPrefix + "_cancel"
)

func (s *streamBot) preHandleNextVideo(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	if slices.Contains(s.AcceptedUsers, update.Message.From.ID) {
		keyboard := models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         "Подтвердить",
						CallbackData: ApproveNextVideoCallback,
					},
					{
						Text:         "Отмена",
						CallbackData: CancelNextVideoCallback,
					},
				},
			},
		}

		b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Вы точно уверены, что хотите переключить видео?",
			ReplyMarkup: keyboard,
		})
	}
}

func (s *streamBot) handleNextVideo(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
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
			editText := "Операция отменена"
			if update.CallbackQuery.Data == ApproveNextVideoCallback {
				editText = "Видео переключено"
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
