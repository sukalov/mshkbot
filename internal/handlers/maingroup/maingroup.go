package maingroup

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sukalov/mshkbot/internal/bot"
)

// GetHandlers returns handler set for main group
func GetHandlers() bot.HandlerSet {
	return bot.HandlerSet{
		Commands: map[string]func(b *bot.Bot, update tgbotapi.Update) error{
			"start":  handleStart,
			"help":   handleHelp,
			"status": handleStatus,
		},
		Messages: []func(b *bot.Bot, update tgbotapi.Update) error{
			handleRegularMessage,
		},
		Callbacks: map[string]func(b *bot.Bot, update tgbotapi.Update) error{
			"action": handleAction,
		},
	}
}

func handleStart(b *bot.Bot, update tgbotapi.Update) error {
	return b.SendMessage(update.Message.Chat.ID, "welcome to main group!")
}

func handleHelp(b *bot.Bot, update tgbotapi.Update) error {
	return b.SendMessage(update.Message.Chat.ID, "main group help text")
}

func handleStatus(b *bot.Bot, update tgbotapi.Update) error {
	return b.SendMessage(update.Message.Chat.ID, "main group status")
}

func handleRegularMessage(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}
	log.Printf("main group message: %s", update.Message.Text)
	return nil
}

func handleAction(b *bot.Bot, update tgbotapi.Update) error {
	return b.SendMessage(update.CallbackQuery.Message.Chat.ID, "action in main group")
}
