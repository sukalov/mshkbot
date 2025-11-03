package cron

import (
	"context"
	"log"
	"time"

	"github.com/sukalov/mshkbot/internal/bot"
)

type Scheduler struct {
	bot         *bot.Bot
	mainGroupID int64
	stopChan    chan struct{}
	timezone    *time.Location
}

// scheduledTask represents a task that runs at a specific time each week
type scheduledTask struct {
	weekday time.Weekday
	hour    int
	minute  int
	handler func()
}

func New(bot *bot.Bot, mainGroupID int64) *Scheduler {
	// moscow timezone (utc+3)
	moscowTZ := time.FixedZone("moscow", 3*60*60)

	return &Scheduler{
		bot:         bot,
		mainGroupID: mainGroupID,
		stopChan:    make(chan struct{}),
		timezone:    moscowTZ,
	}
}

func (s *Scheduler) Start() {
	log.Println("starting cron scheduler")

	s.scheduleWeekly(time.Monday, 12, 00, func() {
		s.scheduledTournamentStart(26, 0, 0, "запись на южный турнир открыта")
	})

	s.scheduleWeekly(time.Monday, 21, 00, func() {
		s.scheduledTournamentEnd()
	})
}

func (s *Scheduler) Stop() {
	log.Println("stopping cron scheduler")
	close(s.stopChan)
}

// scheduleWeekly creates a goroutine that runs a task at the specified weekday and time
func (s *Scheduler) scheduleWeekly(weekday time.Weekday, hour, minute int, handler func()) {
	go func() {
		task := scheduledTask{
			weekday: weekday,
			hour:    hour,
			minute:  minute,
			handler: handler,
		}

		// calculate initial delay
		ticker := time.NewTicker(s.timeUntilNext(task))
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Printf("executing task for %s at %02d:%02d", weekday, hour, minute)
				handler()
				// reset ticker for next week
				ticker.Reset(7 * 24 * time.Hour)
			case <-s.stopChan:
				return
			}
		}
	}()
}

// timeUntilNext calculates duration until the next occurrence of the scheduled task
func (s *Scheduler) timeUntilNext(task scheduledTask) time.Duration {
	now := time.Now().In(s.timezone)

	// calculate days until target weekday
	currentWeekday := int(now.Weekday())
	targetWeekday := int(task.weekday)
	daysUntil := (targetWeekday - currentWeekday + 7) % 7

	// create target time
	targetTime := time.Date(
		now.Year(), now.Month(), now.Day()+daysUntil,
		task.hour, task.minute, 0, 0, s.timezone,
	)

	// if target time is in the past, schedule for next week
	if targetTime.Before(now) || targetTime.Equal(now) {
		targetTime = targetTime.AddDate(0, 0, 7)
	}

	duration := targetTime.Sub(now)
	log.Printf("next %s task in %v (at %s)", task.weekday, duration, targetTime.Format("2006-01-02 15:04:05"))

	return duration
}

func (s *Scheduler) scheduledTournamentStart(limit int, lichessRatingLimit int, chesscomRatingLimit int, announcementIntro string) {
	ctx := context.Background()

	if err := s.bot.Tournament.CreateTournament(ctx, limit, lichessRatingLimit, chesscomRatingLimit, announcementIntro); err != nil {
		log.Printf("failed to create tournament: %v", err)
		return
	}

	announcementMessage := announcementIntro + "\n\nучастники:\nпока никого нет"

	messageID, err := s.bot.SendMessageAndGetID(s.mainGroupID, announcementMessage)
	if err != nil {
		log.Printf("failed to send message: %v", err)
		return
	}

	if err := s.bot.Tournament.SetAnnouncementMessageID(ctx, messageID); err != nil {
		log.Printf("failed to store announcement message ID: %v", err)
	}

	if err := s.bot.PinMessage(s.mainGroupID, messageID); err != nil {
		log.Printf("failed to pin message: %v", err)
	}

	log.Printf("tournament started: limit=%d, lichess_limit=%d, chesscom_limit=%d, intro=%s", limit, lichessRatingLimit, chesscomRatingLimit, announcementIntro)
}

func (s *Scheduler) scheduledTournamentEnd() {
	ctx := context.Background()

	if !s.bot.Tournament.Metadata.Exists {
		log.Printf("no tournament to end")
		return
	}

	announcementMessageID := s.bot.Tournament.Metadata.AnnouncementMessageID
	if announcementMessageID != 0 {
		if err := s.bot.UnpinMessage(s.mainGroupID, announcementMessageID); err != nil {
			log.Printf("failed to unpin message: %v", err)
		}
	}

	if err := s.bot.Tournament.RemoveTournament(ctx); err != nil {
		log.Printf("failed to remove tournament: %v", err)
		return
	}

	log.Printf("tournament ended and removed")
}
