package types

import (
	"time"
)

type Player struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	SavedName string    `json:"saved_name"`
	TimeAdded time.Time `json:"time_added"`
	State     string    `json:"state"`
}

const (
	StateInTournament = "in_tournament"
	StateQueued       = "queued"
)

type TournamentMetadata struct {
	Limit                 int  `json:"limit"`
	RatingLimit           int  `json:"rating_limit"`
	AnnouncementMessageID int  `json:"announcement_message_id"`
	Exists                bool `json:"exists"`
}
