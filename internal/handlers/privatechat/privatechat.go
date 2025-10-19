package privatechat

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sukalov/mshkbot/internal/bot"
	"github.com/sukalov/mshkbot/internal/db"
)

// GetHandlers returns handler set for private messages
func GetHandlers() bot.HandlerSet {
	return bot.HandlerSet{
		Commands: map[string]func(b *bot.Bot, update tgbotapi.Update) error{
			"start": handleStart,
			"help":  handleHelp,
		},
		Messages: []func(b *bot.Bot, update tgbotapi.Update) error{
			handlePrivateMessage,
		},
		Callbacks: map[string]func(b *bot.Bot, update tgbotapi.Update) error{},
	}
}

func handleStart(b *bot.Bot, update tgbotapi.Update) error {
	if err := db.Users.Register(update); err != nil {
		log.Printf("failed to register user: %v", err)
	}

	return b.SendMessage(update.Message.Chat.ID, "welcome")
}

func handleHelp(b *bot.Bot, update tgbotapi.Update) error {
	return b.SendMessage(update.Message.Chat.ID, "private help text")
}

func handlePrivateMessage(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}
	log.Printf("private message from %d: %s", update.Message.From.ID, update.Message.Text)
	return nil
}
