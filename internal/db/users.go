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

type UsersType struct{}

var Users = &UsersType{}

func (u *UsersType) Register(update tgbotapi.Update) error {
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

func (u *UsersType) GetByChatID(chatID int64) (User, error) {
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

func (u *UsersType) UpdateSavedName(chatID int64, newName string) error {
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

// GetAll returns all users
func (u *UsersType) GetAll() ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var users []User
	result := Database.WithContext(ctx).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all users: %w", result.Error)
	}

	return users, nil
}

// Delete removes a user by chat ID
func (u *UsersType) Delete(chatID int64) error {
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
