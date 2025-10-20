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

func SetBanned(chatID int64, isBanned bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := Database.WithContext(ctx).
		Model(&User{}).
		Where("chat_id = ?", chatID).
		Update("is_banned", isBanned)

	if result.Error != nil {
		return fmt.Errorf("failed to update ban status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with chat id: %d", chatID)
	}

	return nil
}

func SetNotGreen(chatID int64, isNotGreen bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := Database.WithContext(ctx).
		Model(&User{}).
		Where("chat_id = ?", chatID).
		Update("is_not_green", isNotGreen)

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

func GetUserState(chatID int64) (State, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	result := Database.WithContext(ctx).
		Select("state").
		Where("chat_id = ?", chatID).
		First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("user not found: %d", chatID)
		}
		return "", fmt.Errorf("failed to get user state: %w", result.Error)
	}

	return user.State, nil
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

func Stringify(u User, markdown bool) string {
	builder := strings.Builder{}

	if u.SavedName != "" {
		builder.WriteString(fmt.Sprintf("ник: %s\n", u.SavedName))
	}
	if u.Lichess != nil && *u.Lichess != "" {
		builder.WriteString(fmt.Sprintf("lichess.org/@/%s\n", *u.Lichess))
	}
	if u.ChessCom != nil && *u.ChessCom != "" {
		builder.WriteString(fmt.Sprintf("chess.com/member/%s\n", *u.ChessCom))
	}

	return builder.String()
}
