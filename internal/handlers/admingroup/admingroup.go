package admingroup

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sukalov/mshkbot/internal/bot"
	"github.com/sukalov/mshkbot/internal/utils"
)

// GetHandlers returns handler set for admin group
func GetHandlers() bot.HandlerSet {
	return bot.HandlerSet{
		Commands: map[string]func(b *bot.Bot, update tgbotapi.Update) error{
			"help":              handleHelp,
			"tournament":        handleTournament,
			"create_tournament": handleCreateTournament,
			"remove_tournament": handleRemoveTournament,
		},
		Messages: []func(b *bot.Bot, update tgbotapi.Update) error{
			handleAdminMessage,
		},
		Callbacks: map[string]func(b *bot.Bot, update tgbotapi.Update) error{},
	}
}

func handleHelp(b *bot.Bot, update tgbotapi.Update) error {
	return b.SendMessage(update.Message.Chat.ID, "команды администратора:\n\n/tournament - показать состояние турнира\n\n/create_tournament - сделать турнир\n\n/remove_tournament - удалить турнир")
}

func handleTournament(b *bot.Bot, update tgbotapi.Update) error {
	jsonStr, err := b.Tournament.GetTournamentJSON()
	if err != nil {
		return err
	}
	return b.SendMessageWithMarkdown(update.Message.Chat.ID, fmt.Sprintf("```json\n%s```", jsonStr), true)
}

func handleCreateTournament(b *bot.Bot, update tgbotapi.Update) error {
	ctx := context.Background()
	if b.Tournament.Metadata.Exists {
		return b.SendMessage(update.Message.Chat.ID, "турнир уже создан")
	}
	if err := b.Tournament.CreateTournament(ctx, 26, 0, 0); err != nil {
		return err
	}
	return b.GiveReaction(update.Message.Chat.ID, update.Message.MessageID, utils.RandomApproveEmoji())
}

func handleRemoveTournament(b *bot.Bot, update tgbotapi.Update) error {
	ctx := context.Background()
	if !b.Tournament.Metadata.Exists {
		return b.SendMessage(update.Message.Chat.ID, "его и так нет")
	}
	announcementMessageID := b.Tournament.Metadata.AnnouncementMessageID
	if announcementMessageID != 0 {
		if err := b.UnpinMessage(b.GetMainGroupID(), announcementMessageID); err != nil {
			log.Printf("failed to unpin message: %v", err)
		}
	}
	if err := b.Tournament.RemoveTournament(ctx); err != nil {
		return err
	}
	return b.GiveReaction(update.Message.Chat.ID, update.Message.MessageID, utils.RandomApproveEmoji())
}

func handleAdminMessage(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}
	log.Printf("admin group message: %s", update.Message.Text)
	return nil
}
