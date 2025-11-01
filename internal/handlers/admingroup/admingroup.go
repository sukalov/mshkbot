package admingroup

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sukalov/mshkbot/internal/bot"
	"github.com/sukalov/mshkbot/internal/db"
	"github.com/sukalov/mshkbot/internal/utils"
)

// GetHandlers returns handler set for admin group
func GetHandlers() bot.HandlerSet {
	return bot.HandlerSet{
		Commands: map[string]func(b *bot.Bot, update tgbotapi.Update) error{
			"help":               handleHelp,
			"tournament":         handleTournament,
			"create_tournament":  handleCreateTournament,
			"remove_tournament":  handleRemoveTournament,
			"suspend_from_green": handleSuspendFromGreen,
			"ban_player":         handleBanPlayer,
		},
		Messages: []func(b *bot.Bot, update tgbotapi.Update) error{
			handleAdminMessage,
		},
		Callbacks: map[string]func(b *bot.Bot, update tgbotapi.Update) error{
			"suspend_duration": handleSuspendDuration,
			"ban_duration":     handleBanDuration,
		},
	}
}

func handleHelp(b *bot.Bot, update tgbotapi.Update) error {
	return b.SendMessage(update.Message.Chat.ID, "команды администратора:\n\n/tournament - показать состояние турнира\n\n/create_tournament - сделать турнир\n\n/remove_tournament - удалить турнир\n\n/suspend_from_green - отстранить пользователя от зелёных турниров\n\n/ban_player - забанить пользователя")
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
	if err := b.Tournament.CreateTournament(ctx, 26, 0, 0, 0); err != nil {
		return err
	}
	return b.GiveReaction(update.Message.Chat.ID, update.Message.MessageID, utils.ApproveEmoji())
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
	return b.GiveReaction(update.Message.Chat.ID, update.Message.MessageID, utils.ApproveEmoji())
}

func handleAdminMessage(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	adminChatID := update.Message.From.ID

	suspendProcess, suspendExists := b.GetSuspensionProcess(adminChatID)
	banProcess, banExists := b.GetBanProcess(adminChatID)

	if !suspendExists && !banExists {
		log.Printf("admin group message: %s", update.Message.Text)
		return nil
	}

	username := strings.TrimPrefix(strings.TrimSpace(update.Message.Text), "@")
	if username == "" {
		b.ClearSuspensionProcess(adminChatID)
		b.ClearBanProcess(adminChatID)
		return b.SendMessage(update.Message.Chat.ID, "юзернейм не может быть пустым")
	}

	user, err := db.GetByUsername(username)
	if err != nil {
		b.ClearSuspensionProcess(adminChatID)
		b.ClearBanProcess(adminChatID)
		return b.SendMessage(update.Message.Chat.ID, fmt.Sprintf("пользователь с юзернеймом %s не найден", username))
	}

	var until *time.Time
	now := time.Now().UTC()

	if suspendExists {
		switch suspendProcess.Duration {
		case "month":
			t := now.AddDate(0, 1, 0)
			until = &t
		case "forever":
			t := now.AddDate(100, 0, 0)
			until = &t
		default:
			b.ClearSuspensionProcess(adminChatID)
			return b.SendMessage(update.Message.Chat.ID, "неизвестная длительность")
		}

		if err := db.SetNotGreenUntil(user.ChatID, until); err != nil {
			b.ClearSuspensionProcess(adminChatID)
			return b.SendMessage(update.Message.Chat.ID, fmt.Sprintf("ошибка при обновлении статуса: %v", err))
		}

		b.ClearSuspensionProcess(adminChatID)

		durationText := "навсегда"
		if suspendProcess.Duration == "month" {
			durationText = "на месяц"
		}

		return b.SendMessage(update.Message.Chat.ID, fmt.Sprintf("пользователь %s отстранён от зелёных %s", username, durationText))
	}

	if banExists {
		switch banProcess.Duration {
		case "month":
			t := now.AddDate(0, 1, 0)
			until = &t
		case "forever":
			t := now.AddDate(100, 0, 0)
			until = &t
		default:
			b.ClearBanProcess(adminChatID)
			return b.SendMessage(update.Message.Chat.ID, "неизвестная длительность")
		}

		if err := db.SetBannedUntil(user.ChatID, until); err != nil {
			b.ClearBanProcess(adminChatID)
			return b.SendMessage(update.Message.Chat.ID, fmt.Sprintf("ошибка при обновлении статуса: %v", err))
		}

		b.ClearBanProcess(adminChatID)

		durationText := "навсегда"
		if banProcess.Duration == "month" {
			durationText = "на месяц"
		}

		return b.SendMessage(update.Message.Chat.ID, fmt.Sprintf("пользователь %s забанен %s", username, durationText))
	}

	return nil
}

func handleSuspendFromGreen(b *bot.Bot, update tgbotapi.Update) error {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("месяц", "suspend_duration:month"),
			tgbotapi.NewInlineKeyboardButtonData("навсегда", "suspend_duration:forever"),
			tgbotapi.NewInlineKeyboardButtonData("отмена", "suspend_duration:cancel"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "выберите длительность отстранения:")
	msg.ReplyMarkup = keyboard

	_, err := b.Client.Send(msg)
	return err
}

func handleSuspendDuration(b *bot.Bot, update tgbotapi.Update) error {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := b.Request(callback); err != nil {
		log.Printf("failed to answer callback: %v", err)
	}

	adminChatID := update.CallbackQuery.From.ID
	data := update.CallbackQuery.Data

	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return fmt.Errorf("invalid callback data: %s", data)
	}

	duration := parts[1]

	if duration == "cancel" {
		b.ClearSuspensionProcess(adminChatID)
		return b.SendMessage(update.CallbackQuery.Message.Chat.ID, "отменено")
	}

	b.SetSuspensionProcess(adminChatID, duration)

	return b.SendMessage(update.CallbackQuery.Message.Chat.ID, "введите telegram username пользователя:")
}

func handleBanPlayer(b *bot.Bot, update tgbotapi.Update) error {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("месяц", "ban_duration:month"),
			tgbotapi.NewInlineKeyboardButtonData("навсегда", "ban_duration:forever"),
			tgbotapi.NewInlineKeyboardButtonData("отмена", "ban_duration:cancel"),
		),
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "выберите длительность бана:")
	msg.ReplyMarkup = keyboard

	_, err := b.Client.Send(msg)
	return err
}

func handleBanDuration(b *bot.Bot, update tgbotapi.Update) error {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := b.Request(callback); err != nil {
		log.Printf("failed to answer callback: %v", err)
	}

	adminChatID := update.CallbackQuery.From.ID
	data := update.CallbackQuery.Data

	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return fmt.Errorf("invalid callback data: %s", data)
	}

	duration := parts[1]

	if duration == "cancel" {
		b.ClearBanProcess(adminChatID)
		return b.SendMessage(update.CallbackQuery.Message.Chat.ID, "отменено")
	}

	b.SetBanProcess(adminChatID, duration)

	return b.SendMessage(update.CallbackQuery.Message.Chat.ID, "введите telegram username пользователя:")
}
