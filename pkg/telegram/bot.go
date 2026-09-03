package telegram

import (
	"context"
	"strings"

	"github.com/Kudzeri/job-prep-backend/internal/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api  *tgbotapi.BotAPI
	repo *repository.AuthRepository
}

func NewBot(token string, repo *repository.AuthRepository) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		api:  api,
		repo: repo,
	}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		ctx := context.Background()
		msg := update.Message

		// Обработка команды /start <auth_token>
		if msg.Command() == "start" {
			args := strings.TrimSpace(msg.CommandArguments())

			if args == "" {
				reply := tgbotapi.NewMessage(msg.Chat.ID, "Привет! Для входа на сайт перейдите по ссылке авторизации с нашего веб-приложения.")
				b.api.Send(reply)
				continue
			}

			authToken := args
			telegramID := msg.From.ID
			username := msg.From.UserName

			// Подтверждаем сессию
			err := b.repo.ConfirmTelegramSession(ctx, authToken, telegramID)
			if err != nil {
				reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ Ошибка авторизации или срок действия ссылки истёк.")
				b.api.Send(reply)
				continue
			}

			// Создаем или получаем пользователя
			_, err = b.repo.GetUserByTelegramID(ctx, telegramID)
			if err != nil {
				_, _ = b.repo.CreateUserWithTelegram(ctx, telegramID, username)
			}

			reply := tgbotapi.NewMessage(msg.Chat.ID, "✅ Вы успешно авторизовались! Вернитесь на сайт для продолжения.")
			b.api.Send(reply)
		}
	}
}