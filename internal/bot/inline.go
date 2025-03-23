package bot

import (
	"context"
	"fmt"
	"slices"
	"strings"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	TokenAcceptPrefix = "token_"
)

func (s *streamBot) handleInline(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	logger := zerolog.Ctx(ctx)

	if update.InlineQuery == nil {
		return
	}

	if slices.Contains(s.AcceptedUsers, update.InlineQuery.From.ID) {
		logger.Info().
			Int64("user", update.InlineQuery.From.ID).
			Msgf("Handle inline query")

		temporaryKey, err := gonanoid.New()
		if err != nil {
			logger.Err(err)

			b.AnswerInlineQuery(ctx, &telegramBot.AnswerInlineQueryParams{
				InlineQueryID: update.InlineQuery.ID,
				IsPersonal:    true,
				CacheTime:     1,
				Results: []models.InlineQueryResult{
					&models.InlineQueryResultArticle{
						ID:          "error",
						Title:       "Ошибка",
						Description: err.Error(),
						InputMessageContent: &models.InputTextMessageContent{
							MessageText: "Error",
						},
					},
				},
			})
			return
		}

		s.StreamTokensMap[temporaryKey] = update.InlineQuery.Query

		results := make([]models.InlineQueryResult, 0, len(s.Streams))
		for platform := range s.Streams {
			platformName := cases.Title(language.Russian, cases.Compact).String(string(platform))
			results = append(results, &models.InlineQueryResultArticle{
				ID:          TokenAcceptPrefix + string(platform),
				Title:       platformName,
				Description: fmt.Sprintf("Применить ключ трансляции для платформы %s", platformName),
				InputMessageContent: &models.InputTextMessageContent{
					MessageText: fmt.Sprintf("stream_key:%s:%s", platform, temporaryKey),
				},
			})
		}

		b.AnswerInlineQuery(ctx, &telegramBot.AnswerInlineQueryParams{
			InlineQueryID: update.InlineQuery.ID,
			IsPersonal:    true,
			CacheTime:     1,
			Results:       results,
		})
	} else {
		b.AnswerInlineQuery(ctx, &telegramBot.AnswerInlineQueryParams{
			InlineQueryID: update.InlineQuery.ID,
			IsPersonal:    true,
			CacheTime:     1,
			Results: []models.InlineQueryResult{
				&models.InlineQueryResultArticle{
					ID:          "none",
					Title:       "Доступ запрещен",
					Description: "Для получения доступа обратитесь к владельцу",
					InputMessageContent: &models.InputTextMessageContent{
						MessageText: "Error",
					},
				},
			},
		})
	}
}

func (s *streamBot) handleSetStreamKey(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	logger := zerolog.Ctx(ctx)

	if slices.Contains(s.AcceptedUsers, update.Message.From.ID) {
		logger.Info().
			Int64("user", update.Message.From.ID).
			Msgf("Handle set stream key")

		parts := strings.Split(update.Message.Text, ":")
		if len(parts) != 3 {
			b.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Не удалось распарсить данные",
			})
			return
		}

		originalToken, ok := s.StreamTokensMap[parts[2]]
		if !ok {
			b.SendMessage(ctx, &telegramBot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Ключ не найден",
			})
			return
		}

		for platform, stream := range s.Streams {
			if string(platform) == parts[1] {
				stream.SetStreamToken(originalToken)

				b.SendMessage(ctx, &telegramBot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   fmt.Sprintf("Ключ трансляции для платформы %s установлен", platform),
				})
				return
			}
		}

		b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Неизвестная платформа: %s", parts[1]),
		})
	}
}
