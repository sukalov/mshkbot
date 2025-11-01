package utils

import (
	"math/rand"
)

func ApproveEmoji() string {
	n := rand.Intn(100)

	if n < 2 {
		return "🎉"
	}

	remaining := (n - 2) % 5

	switch remaining {
	case 0:
		return "👍"
	case 1:
		return "👌"
	case 2:
		return "🫡"
	case 3:
		return "✍️"
	default:
		return "🤝"
	}
}

func AlreadyCheckedInMessage() string {
	n := rand.Intn(4)

	switch n {
	case 0:
		return "вы уже записаны на турнир"
	case 1:
		return "хватит тыкать, вы уже записаны"
	case 2:
		return "второй раз записаться нельзя"
	default:
		return "достаточно записаться один раз"
	}
}

func CheckinUnavailibleMessage() string {
	n := rand.Intn(5)

	switch n {
	case 0:
		return "сейчас нелья записаться"
	case 1:
		return "турнир ещё не начался"
	case 2:
		return "попробуйте позже"
	case 3:
		return "ceйчас никуда не могу записать"
	default:
		return "сначала дождитесь объявления"
	}
}

func NoTournamentMessage() string {
	n := rand.Intn(5)

	switch n {
	case 0:
		return "турнира нет пока"
	case 1:
		return "турнир ещё не начался"
	case 2:
		return "попробуйте позже"
	case 3:
		return "ceйчас никуда не могу записать"
	default:
		return "дождитесь объявления"
	}
}

func SadEmoji() string {
	n := rand.Intn(4)

	switch n {
	case 0, 1:
		return "😢"
	case 2:
		return "💔"
	default:
		return "😭"
	}
}
