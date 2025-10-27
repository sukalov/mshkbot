package maingroup

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sukalov/mshkbot/internal/bot"
	"github.com/sukalov/mshkbot/internal/db"
	"github.com/sukalov/mshkbot/internal/utils"
)

// GetHandlers returns handler set for main group
func GetHandlers() bot.HandlerSet {
	return bot.HandlerSet{
		Commands: map[string]func(b *bot.Bot, update tgbotapi.Update) error{
			"checkin":  handleCheckIn,
			"checkout": handleCheckOut,
			"help":     handleHelp,
		},
		Messages: []func(b *bot.Bot, update tgbotapi.Update) error{
			handleRegularMessage,
		},
		Callbacks: map[string]func(b *bot.Bot, update tgbotapi.Update) error{
			"action": handleAction,
		},
	}
}

func handleHelp(b *bot.Bot, update tgbotapi.Update) error {
	return b.SendMessage(update.Message.Chat.ID, "/checkin — записаться на турнир\n\n/checkout — выход из турнира")
}

func handleCheckIn(b *bot.Bot, update tgbotapi.Update) error {
	user, err := db.GetUser(update.Message.From.ID)
	if err != nil {
		if err.Error() == "user not found" {
			return b.ReplyToMessage(update.Message.Chat.ID, update.Message.MessageID, "напишите мне в личку чтобы зарегистрироваться")
		}
		return b.SendMessage(update.Message.From.ID, fmt.Sprintf("ошибка: %v. попробуйте ещё раз и если ничего не получается, напишите @sukalov", err))
	}
	if user.State != db.StateCompleted {
		return b.ReplyToMessage(update.Message.Chat.ID, update.Message.MessageID, "мы с вами в личке ещё не закончили регистрацию")
	}

	// TODO: check if user is already checked in
	// TODO: check if tournament is full
	// TODO: add user to tournament

	

	return b.GiveReaction(update.Message.Chat.ID, update.Message.MessageID, utils.ApproveEmoji())
}

func handleCheckOut(b *bot.Bot, update tgbotapi.Update) error {
	return b.SendMessage(update.Message.Chat.ID, "check out")
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
