// users.go
package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func Register(update tgbotapi.Update) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message := update.Message
	userName := message.From.UserName
	tgName := strings.TrimSpace(message.From.FirstName + " " + message.From.LastName)

	user := User{
		ChatID:   message.Chat.ID,
		Username: userName,
		TgName:   tgName,
	}

	// FirstOrCreate to avoid duplicates
	result := Database.WithContext(ctx).Where(User{ChatID: message.Chat.ID}).FirstOrCreate(&user)
	if result.Error != nil {
		return fmt.Errorf("failed to register user: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		log.Printf("new user registered: id: %d, username: %s", message.Chat.ID, userName)
	}

	return nil
}

func GetByChatID(chatID int64) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	result := Database.WithContext(ctx).Where("chat_id = ?", chatID).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return User{}, fmt.Errorf("user not found: %d", chatID)
		}
		return User{}, fmt.Errorf("failed to retrieve user: %w", result.Error)
	}

	return user, nil
}

func GetByUsername(username string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	result := Database.WithContext(ctx).Where("username = ?", username).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return User{}, fmt.Errorf("user not found")
		}
		return User{}, fmt.Errorf("failed to retrieve user: %w", result.Error)
	}

	return user, nil
}

func UpdateSavedName(chatID int64, newName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := Database.WithContext(ctx).
		Model(&User{}).
		Where("chat_id = ?", chatID).
		Update("saved_name", newName)

	if result.Error != nil {
		return fmt.Errorf("failed to update saved name: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with chat id: %d", chatID)
	}

	return nil
}

func UpdateLichess(chatID int64, lichess string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var value *string
	if lichess == "" {
		return fmt.Errorf("update lichess with ''")
	}
	value = &lichess

	result := Database.WithContext(ctx).
		Model(&User{}).
		Where("chat_id = ?", chatID).
		Update("lichess", value)

	if result.Error != nil {
		return fmt.Errorf("failed to update lichess: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with chat id: %d", chatID)
	}

	return nil
}

func UpdateChessCom(chatID int64, chessCom string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var value *string
	if chessCom == "" {
		return fmt.Errorf("update chesscom with ''")
	}
	value = &chessCom

	result := Database.WithContext(ctx).
		Model(&User{}).
		Where("chat_id = ?", chatID).
		Update("chesscom", value)

	if result.Error != nil {
		return fmt.Errorf("failed to update chesscom: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with chat id: %d", chatID)
	}

	return nil
}

func UpdateState(chatID int64, state State) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := Database.WithContext(ctx).
		Model(&User{}).
		Where("chat_id = ?", chatID).
		Update("state", state)

	if result.Error != nil {
		return fmt.Errorf("failed to update state: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with chat id: %d", chatID)
	}

	return nil
}

func SetBannedUntil(chatID int64, until *time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := Database.WithContext(ctx).
		Model(&User{}).
		Where("chat_id = ?", chatID).
		Update("banned_until", until)

	if result.Error != nil {
		return fmt.Errorf("failed to update ban status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with chat id: %d", chatID)
	}

	return nil
}

func SetNotGreenUntil(chatID int64, until *time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := Database.WithContext(ctx).
		Model(&User{}).
		Where("chat_id = ?", chatID).
		Update("not_green_until", until)

	if result.Error != nil {
		return fmt.Errorf("failed to update green status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with chat id: %d", chatID)
	}

	return nil
}

// GetAll returns all users
func GetAll() ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var users []User
	result := Database.WithContext(ctx).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all users: %w", result.Error)
	}

	return users, nil
}

func GetUser(chatID int64) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	result := Database.WithContext(ctx).
		Select("state").
		Where("chat_id = ?", chatID).
		First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return User{}, fmt.Errorf("user not found")
		}
		return User{}, fmt.Errorf("failed to get user state: %w", result.Error)
	}

	return user, nil
}

// Delete removes a user by chat ID
func Delete(chatID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := Database.WithContext(ctx).Where("chat_id = ?", chatID).Delete(&User{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with chat id: %d", chatID)
	}

	return nil
}

func Stringify(u User) string {
	builder := strings.Builder{}

	if u.SavedName != "" {
		builder.WriteString(fmt.Sprintf("ник: %s\n", u.SavedName))
	}
	if u.Lichess != nil && *u.Lichess != "" {
		builder.WriteString(fmt.Sprintf("lichess: [%s](https://lichess.org/@/%s)\n", *u.Lichess, *u.Lichess))
	}
	if u.ChessCom != nil && *u.ChessCom != "" {
		builder.WriteString(fmt.Sprintf("chess.com: [%s](https://www.chess.com/member/%s)\n", *u.ChessCom, *u.ChessCom))
	}

	return builder.String()
}

// UpdateLichessAndState updates lichess username and state in one transaction
func UpdateLichessAndState(chatID int64, lichess string, newState State) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if lichess == "" {
		return fmt.Errorf("update lichess with ''")
	}

	result := Database.WithContext(ctx).
		Model(&User{}).
		Where("chat_id = ?", chatID).
		Updates(map[string]interface{}{
			"lichess": &lichess,
			"state":   newState,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update lichess and state: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with chat id: %d", chatID)
	}

	return nil
}

// UpdateChessComAndState updates chess.com username and state in one transaction
func UpdateChessComAndState(chatID int64, chessCom string, newState State) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if chessCom == "" {
		return fmt.Errorf("update chesscom with ''")
	}

	result := Database.WithContext(ctx).
		Model(&User{}).
		Where("chat_id = ?", chatID).
		Updates(map[string]interface{}{
			"chesscom": &chessCom,
			"state":    newState,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update chesscom and state: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with chat id: %d", chatID)
	}

	return nil
}

// GetOrCreateUser combines getting and creating user in one operation
func GetOrCreateUser(update tgbotapi.Update) (User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message := update.Message
	userName := message.From.UserName
	tgName := strings.TrimSpace(message.From.FirstName + " " + message.From.LastName)

	user := User{
		ChatID:   message.Chat.ID,
		Username: userName,
		TgName:   tgName,
	}

	result := Database.WithContext(ctx).Where(User{ChatID: message.Chat.ID}).FirstOrCreate(&user)
	if result.Error != nil {
		return User{}, false, fmt.Errorf("failed to get/create user: %w", result.Error)
	}

	isNew := result.RowsAffected > 0
	if isNew {
		log.Printf("new user registered: id: %d, username: %s", message.Chat.ID, userName)
	}

	return user, isNew, nil
}
