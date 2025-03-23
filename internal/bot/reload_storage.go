package bot

import (
	"context"
	"fmt"
	"slices"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog"
)

func (s *streamBot) handleReloadStorage(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	logger := zerolog.Ctx(ctx)

	if slices.Contains(s.AcceptedUsers, update.Message.From.ID) {
		logger.Info().
			Int64("user", update.Message.From.ID).
			Msgf("Handle reload video storage")

		b.SendChatAction(ctx, &telegramBot.SendChatActionParams{
			ChatID: update.Message.Chat.ID,
			Action: models.ChatActionTyping,
		})

		err := s.VideoStorage.UpdateFilesList(ctx)

		msgText := "Список видеозаписей успешно обновлен"
		if err != nil {
			msgText = fmt.Sprintf("Ошибка обновления видеозаписей: \n ```%v```", err)
		}

		b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      msgText,
			ParseMode: models.ParseModeMarkdown,
		})
	}
}
