package admingroup

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sukalov/mshkbot/internal/bot"
)

// GetHandlers returns handler set for admin group
func GetHandlers() bot.HandlerSet {
	return bot.HandlerSet{
		Commands: map[string]func(b *bot.Bot, update tgbotapi.Update) error{
			"start": handleStart,
		},
		Messages: []func(b *bot.Bot, update tgbotapi.Update) error{
			handleAdminMessage,
		},
		Callbacks: map[string]func(b *bot.Bot, update tgbotapi.Update) error{},
	}
}

func handleStart(b *bot.Bot, update tgbotapi.Update) error {
	return b.SendMessage(update.Message.Chat.ID, "admin start handler")
}

func handleAdminMessage(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}
	log.Printf("admin group message: %s", update.Message.Text)
	return nil
}
