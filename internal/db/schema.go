// schema.go
package db

import (
	"time"

	"gorm.io/gorm"
)

// User represents a telegram user
type User struct {
	ChatID     int64     `gorm:"primaryKey;column:chat_id"`
	Username   string    `gorm:"column:username;index"`
	TgName     string    `gorm:"column:tg_name"`
	SavedName  string    `gorm:"column:saved_name"`
	LichessID  string    `gorm:"column:lichess_id"`
	ChessComID string    `gorm:"column:chesscom_id"`
	AddedAt    time.Time `gorm:"column:added_at;autoCreateTime"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook - runs before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.AddedAt.IsZero() {
		u.AddedAt = time.Now().UTC()
	}
	return nil
}

// add more models below as your project grows
// example:
// type Message struct {
//     ID        uint      `gorm:"primaryKey"`
//     UserID    int64     `gorm:"index;not null"`
//     Content   string    `gorm:"type:text"`
//     CreatedAt time.Time `gorm:"autoCreateTime"`
//     User      User      `gorm:"foreignKey:UserID;references:ChatID"`
// }
