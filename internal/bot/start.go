package bot

import (
	"context"
	"slices"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (s *streamBot) handleStart(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	if slices.Contains(s.AcceptedUsers, update.Message.From.ID) {
		keyboard := models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{
					{
						Text: string(ButtonTypeNext),
					},
					{
						Text: string(ButtonTypeStatistics),
					},
				},
				{
					{
						Text: string(ButtonTypeReloadList),
					},
				},
				{
					{
						Text: string(ButtonTypeStart),
					},
					{
						Text: string(ButtonTypeStop),
					},
				},
			},
		}

		b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Для управления потоком воспользуйтесь кнопками",
			ReplyMarkup: keyboard,
		})
	}
}
