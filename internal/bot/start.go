package bot

import (
	"context"
	"slices"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog"
)

func (s *streamBot) handleStart(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	logger := zerolog.Ctx(ctx)

	if slices.Contains(s.AcceptedUsers, update.Message.From.ID) {
		logger.Info().
			Int64("user", update.Message.From.ID).
			Msgf("Handle start bot")

		keyboard := models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{
					{
						Text: string(ButtonTypeNext),
					},
					{
						Text: string(ButtonTypeQueue),
					},
				},
				{
					{
						Text: string(ButtonTypeReloadList),
					},
					{
						Text: string(ButtonTypeStatistics),
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
