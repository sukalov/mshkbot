package privatechat

import (
	"fmt"
	"log"
	"strings"

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
		Callbacks: map[string]func(b *bot.Bot, update tgbotapi.Update) error{
			"register": handleRegister,
		},
	}
}

func handleStart(b *bot.Bot, update tgbotapi.Update) error {
	chatID := update.Message.Chat.ID

	// check user state
	state, err := db.GetUserState(chatID)
	if err != nil {
		log.Printf("failed to get user state: %v", err)
		// user doesn't exist yet, continue with registration
	} else {
		// user exists, check their state
		switch state {
		case db.StateCompleted:
			return b.SendMessage(chatID, "вы уже зарегистрированы!")

		case db.StateAskedLichess:
			platform := "lichess"
			return b.SendMessage(chatID, fmt.Sprintf("введите ваш никнейм на %s:", platform))

		case db.StateAskedChessCom:
			platform := "chess.com"
			return b.SendMessage(chatID, fmt.Sprintf("введите ваш никнейм на %s:", platform))

		case db.StateAskedSavedName:
			return b.SendMessage(chatID, "введите ваш никнейм для турниров:")
		}
	}

	if err := db.Register(update); err != nil {
		log.Printf("failed to register user: %v", err)
	}

	row := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("lichess", "register:lichess"),
		tgbotapi.NewInlineKeyboardButtonData("chess.com", "register:chess.com"),
	}
	row2 := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("нигде не играю (честное слово)", "register:none"),
	}

	return b.SendMessageWithButtons(chatID, "привет! чтобы записываться на турниры нужно показать свой шахматный уровень. где вы играете?", tgbotapi.NewInlineKeyboardMarkup(row, row2))
}

func handleRegister(b *bot.Bot, update tgbotapi.Update) error {
	chatID := update.CallbackQuery.Message.Chat.ID
	data := update.CallbackQuery.Data

	// answer callback query to remove loading state
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := b.Request(callback); err != nil {
		log.Printf("failed to answer callback: %v", err)
	}

	// parse option from callback data
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return fmt.Errorf("invalid callback data: %s", data)
	}

	option := parts[1]

	switch option {
	case "lichess":
		editMsg := tgbotapi.NewEditMessageText(
			chatID,
			update.CallbackQuery.Message.MessageID,
			"введите ваш никнейм на lichess:",
		)
		if _, err := b.Request(editMsg); err != nil {
			return fmt.Errorf("failed to edit message: %w", err)
		}
		if err := db.UpdateState(chatID, db.StateAskedLichess); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}

	case "chess.com":
		editMsg := tgbotapi.NewEditMessageText(
			chatID,
			update.CallbackQuery.Message.MessageID,
			"введите ваш никнейм на chess.com:",
		)
		if _, err := b.Request(editMsg); err != nil {
			return fmt.Errorf("failed to edit message: %w", err)
		}
		if err := db.UpdateState(chatID, db.StateAskedChessCom); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}

	case "none":
		editMsg := tgbotapi.NewEditMessageText(
			chatID,
			update.CallbackQuery.Message.MessageID,
			"введите ваш никнейм для турниров:",
		)
		if _, err := b.Request(editMsg); err != nil {
			return fmt.Errorf("failed to edit message: %w", err)
		}
		if err := db.UpdateState(chatID, db.StateAskedSavedName); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}

	default:
		return fmt.Errorf("unknown register option: %s", option)
	}

	return nil
}

func handleHelp(b *bot.Bot, update tgbotapi.Update) error {
	return b.SendMessage(update.Message.Chat.ID, "help")
}

func handlePrivateMessage(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	chatID := update.Message.Chat.ID

	// get user state
	state, err := db.GetUserState(chatID)
	if err != nil {
		log.Printf("failed to get user state: %v", err)
		return nil
	}

	switch state {
	case db.StateAskedLichess:
		username := strings.TrimSpace(update.Message.Text)

		// save the username
		if err := db.UpdateLichess(chatID, username); err != nil {
			log.Printf("failed to update lichess username: %v", err)
			return b.SendMessage(chatID, "произошла ошибка, попробуйте еще раз")
		}

		// ask for saved name
		if err := db.UpdateState(chatID, db.StateAskedSavedName); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}

		return b.SendMessage(chatID, "введите ваш никнейм для турниров:")

	case db.StateAskedChessCom:
		username := strings.TrimSpace(update.Message.Text)

		// save the username
		if err := db.UpdateChessCom(chatID, username); err != nil {
			log.Printf("failed to update lichess username: %v", err)
			return b.SendMessage(chatID, "произошла ошибка, попробуйте еще раз")
		}

		// ask for saved name
		if err := db.UpdateState(chatID, db.StateAskedSavedName); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}

		return b.SendMessage(chatID, "введите ваш никнейм для турниров:")

	case db.StateAskedSavedName:
		// user sent their saved name
		savedName := strings.TrimSpace(update.Message.Text)

		if err := db.UpdateSavedName(chatID, savedName); err != nil {
			log.Printf("failed to update saved name: %v", err)
			return b.SendMessage(chatID, "произошла ошибка, попробуйте еще раз")
		}

		// registration complete
		if err := db.UpdateState(chatID, db.StateCompleted); err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}

		return b.SendMessage(chatID, "отлично! регистрация завершена. теперь можете записываться на турниры в чате @moscowchessclub")

	default:
		log.Printf("private message from %d: %s", update.Message.From.ID, update.Message.Text)
	}

	return nil
}
