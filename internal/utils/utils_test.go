package utils

import (
	"testing"
)

func TestGetLichessAllTimeHigh(t *testing.T) {
	ratings, err := GetLichessAllTimeHigh("moscow_chess_club")
	t.Logf("[lichess] moscow_chess_club: %v", ratings)
	if err != nil {
		t.Errorf("failed to get lichess ratings: %v", err)
	}
}

func TestGetChessComAllTimeHigh(t *testing.T) {
	ratings, err := GetChessComAllTimeHigh("sukalov")
	t.Logf("[chesscom] sukalov: %v", ratings)
	if err != nil {
		t.Errorf("failed to get chesssom ratings: %v", err)
	}
}
