package types

import (
	"time"
)

type Player struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	TimeAdded time.Time `json:"time_added"`
	State     string    `json:"state"`
}

const (
	StateInTournament = "in_tournament"
	StateQueued       = "queued"
)
