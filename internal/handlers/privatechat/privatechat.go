package privatechat

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sukalov/mshkbot/internal/bot"
	"github.com/sukalov/mshkbot/internal/db"
	"github.com/sukalov/mshkbot/internal/utils"
)

// GetHandlers returns handler set for private messages
func GetHandlers() bot.HandlerSet {
	return bot.HandlerSet{
		Commands: map[string]func(b *bot.Bot, update tgbotapi.Update) error{
			"start":     handleStart,
			"help":      handleHelp,
			"me":        handleMe,
			"myratings": handleMyRatings,
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

	// Get or create user in one operation
	user, isNew, err := db.GetOrCreateUser(update)
	if err != nil {
		log.Printf("failed to get/create user: %v", err)
		return err
	}

	if !isNew {
		// User exists, check their state
		switch user.State {
		case db.StateCompleted:
			return b.SendMessage(chatID, "вы уже зарегистрированы!")
		case db.StateAskedLichess:
			return b.SendMessage(chatID, "введите ваш никнейм на lichess:")
		case db.StateAskedChessCom:
			return b.SendMessage(chatID, "введите ваш никнейм на chess.com:")
		case db.StateAskedSavedName:
			return b.SendMessage(chatID, "введите ваш никнейм для турниров:")
		}
	}

	// Show registration options for new users
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

func handleMe(b *bot.Bot, update tgbotapi.Update) error {
	chatID := update.Message.Chat.ID

	if user, err := db.GetByChatID(chatID); err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	} else {
		return b.SendMessage(chatID, db.Stringify(user, false))
	}
}

func handleMyRatings(b *bot.Bot, update tgbotapi.Update) error {
	chatID := update.Message.Chat.ID
	var lichess, chesscom string
	if user, err := db.GetByChatID(chatID); err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	} else {

		if user.Lichess == nil || *user.Lichess == "" {
			lichess = "личес не указан"
		}
		if user.ChessCom == nil || *user.ChessCom == "" {
			chesscom = "чесском не указан"
		}

		if user.Lichess != nil {
			lichessTopRatings, err := utils.GetLichessAllTimeHigh(*user.Lichess)
			if err != nil {
				return fmt.Errorf("ошибка при запросе к базе личеса: %w", err)
			}

			lichess = fmt.Sprintf("пиковые рейтинги на личесе: блиц %d, рапид %d, классика %d", lichessTopRatings.Blitz, lichessTopRatings.Rapid, lichessTopRatings.Classical)
		}
		if user.ChessCom != nil {
			chesscomTopRatings, err := utils.GetChessComAllTimeHigh(*user.ChessCom)
			if err != nil {
				return fmt.Errorf("ошибка при запросе к базе чесскома: %w", err)
			}
			chesscom = fmt.Sprintf("пиковые рейтинги на чесскоме: блиц %d, рапид %d, классика %d", chesscomTopRatings.Blitz, chesscomTopRatings.Rapid, chesscomTopRatings.Classical)
		}

		return b.SendMessage(chatID, fmt.Sprintf("%s\n%s", lichess, chesscom))
	}
}

func handlePrivateMessage(b *bot.Bot, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	chatID := update.Message.Chat.ID

	user, err := db.GetUser(chatID) // DB CALL 1
	if err != nil {
		log.Printf("failed to get user state: %v", err)
		return nil
	}

	switch user.State {
	case db.StateAskedLichess:
		username := strings.TrimPrefix(strings.TrimSpace(update.Message.Text), "@")
		if username == "" {
			return b.SendMessage(chatID, "юзернейм не может быть пустым")
		}

		allTimeHigh, err := utils.GetLichessAllTimeHigh(username)
		if err != nil {
			return b.SendMessage(chatID, "произошла ошибка, попробуйте ещё раз")
		}
		log.Printf("all time high: %d", allTimeHigh)

		// save the username
		if err := db.UpdateLichess(chatID, username); err != nil { // DB CALL 2
			log.Printf("failed to update lichess username: %v", err)
			return b.SendMessage(chatID, fmt.Sprintf("произошла ошибка, попробуйте ещё раз: %v", err))
		}

		// ask for saved name
		if err := db.UpdateState(chatID, db.StateAskedSavedName); err != nil { // DB CALL 3
			return fmt.Errorf("failed to update state: %w", err)
		}

		return b.SendMessage(chatID, "введите ваш никнейм для турниров:")

	case db.StateAskedChessCom:
		username := strings.TrimPrefix(strings.TrimSpace(update.Message.Text), "@")
		if username == "" {
			return b.SendMessage(chatID, "юзернейм не может быть пустым")
		}

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
