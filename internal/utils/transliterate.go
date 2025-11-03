package utils

import (
	"strings"
	"unicode"
)

var latinToCyrillic = map[rune]rune{
	'a': 'а', 'b': 'б', 'v': 'в', 'g': 'г', 'd': 'д',
	'e': 'е', 'z': 'з', 'i': 'и', 'j': 'й', 'k': 'к',
	'l': 'л', 'm': 'м', 'n': 'н', 'o': 'о', 'p': 'п',
	'r': 'р', 's': 'с', 't': 'т', 'u': 'у', 'f': 'ф',
	'h': 'х', 'c': 'ц', 'y': 'ы',
}

var numberToCyrillic = map[rune]string{
	'0': "ноль",
	'1': "один",
	'2': "два",
	'3': "три",
	'4': "четыре",
	'5': "пять",
	'6': "шесть",
	'7': "семь",
	'8': "восемь",
	'9': "девять",
}

func Transliterate(input string) string {
	var result strings.Builder
	exclamationCount := 0

	for _, r := range input {
		if unicode.IsUpper(r) {
			r = unicode.ToLower(r)
		}

		if r >= '0' && r <= '9' {
			if cyrillicNum, ok := numberToCyrillic[r]; ok {
				result.WriteString(cyrillicNum)
			}
		} else if r >= 'a' && r <= 'z' {
			if cyrillic, ok := latinToCyrillic[r]; ok {
				result.WriteRune(cyrillic)
			}
		} else if (r >= 'а' && r <= 'я') || r == 'ё' {
			result.WriteRune(r)
		} else if r == '!' && exclamationCount == 0 {
			result.WriteRune(r)
			exclamationCount++
		} else if r == ' ' {
			result.WriteRune(r)
		}
	}

	return strings.TrimSpace(result.String())
}
