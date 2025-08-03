package secure_pass

import "unicode"

const (
	MIN_LENGTH		= 12
	SPECIAL_CHAR	= "!@#$%&+-/*="
)

func Minimum_length(s string) bool {
	if len([]rune(s)) < MIN_LENGTH {
		return false
	}
	return true
}

func Entropy(s string) bool {
	var (
		has_digit	bool
		has_letter	bool
		has_special	bool
	)
	for _, char := range s {
		switch {
		case unicode.IsDigit(char):
			has_digit = true
		case unicode.IsLetter(char):
			has_letter = true
		case contains_rune(SPECIAL_CHAR, char):
			has_special = true
		}
	}
	if has_digit && has_letter && has_special {
		return true
	}
	return false
}

func contains_rune(s string, r rune) bool {
	for _, char := range s {
		if char == r {
			return true
		}
	}
	return false
}