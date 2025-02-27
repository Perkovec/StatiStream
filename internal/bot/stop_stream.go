package bot

import (
	"context"
	"slices"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	StopStreamCallbackPrefix  = "stop_stream"
	ApproveStopStreamCallback = StopStreamCallbackPrefix + "_approve"
	CancelStopStreamCallback  = StopStreamCallbackPrefix + "_cancel"
)

func (s *streamBot) preStopStream(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	if slices.Contains(s.AcceptedUsers, update.Message.From.ID) {
		keyboard := models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         "Подтвердить",
						CallbackData: ApproveStopStreamCallback,
					},
					{
						Text:         "Отмена",
						CallbackData: CancelStopStreamCallback,
					},
				},
			},
		}

		b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Вы точно уверены, что хотите остановить стрим?",
			ReplyMarkup: keyboard,
		})
	}
}
