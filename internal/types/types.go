package types

import (
	"time"
)

type Player struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	SavedName      string    `json:"saved_name"`
	TimeAdded      time.Time `json:"time_added"`
	State          string    `json:"state"`
	CheckedOutTime time.Time `json:"checked_out_time,omitempty"`
}

const (
	StateInTournament = "in_tournament"
	StateQueued       = "queued"
	StateCheckedOut   = "checked_out"
)

type TournamentMetadata struct {
	Limit                 int  `json:"limit"`
	LichessRatingLimit    int  `json:"lichess_rating_limit"`
	ChesscomRatingLimit   int  `json:"chesscom_rating_limit"`
	AnnouncementMessageID int  `json:"announcement_message_id"`
	Exists                bool `json:"exists"`
}
