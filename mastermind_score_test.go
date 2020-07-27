package main

import (
	"fmt"
	"testing"
)

func TestPerfectScore(t *testing.T) {
	secret := SecretCode{red, red, orange, blue}
	guess := GuessCode{red, red, orange, blue}
	expectedScore := Score{red: 4}

	actualScore := calculateScore(secret, guess)
	if actualScore != expectedScore {
		t.Fail()
	}
}

func TestScoreNone(t *testing.T) {
	secret := SecretCode{red, red, orange, blue}
	guess := GuessCode{green, green, purple, yellow}
	expectedScore := Score{}

	actualScore := calculateScore(secret, guess)
	if actualScore != expectedScore {
		t.Fail()
	}
}

func TestScoreRepeatedGuessPegs1(t *testing.T) {
	secret := SecretCode{red, orange, yellow, green}
	guess := GuessCode{red, red, blue, blue}
	expectedScore := Score{red: 1}

	actualScore := calculateScore(secret, guess)
	if actualScore != expectedScore {
		t.Fail()
	}
}

func TestScoreRepeatedGuessPegs2(t *testing.T) {
	secret := SecretCode{orange, yellow, green, red}
	guess := GuessCode{red, red, blue, blue}
	expectedScore := Score{white: 1}

	actualScore := calculateScore(secret, guess)
	if actualScore != expectedScore {
		fmt.Println(actualScore)
		t.Fail()
	}
}

func TestScoreRepeatedSecretPegs1(t *testing.T) {
	secret := SecretCode{red, red, blue, purple}
	guess := GuessCode{red, orange, yellow, green}
	expectedScore := Score{red: 1}

	actualScore := calculateScore(secret, guess)
	if actualScore != expectedScore {
		t.Fail()
	}
}

func TestScoreRepeatedSecretPegs2(t *testing.T) {
	secret := SecretCode{red, red, blue, purple}
	guess := GuessCode{orange, yellow, green, red}
	expectedScore := Score{white: 1}

	actualScore := calculateScore(secret, guess)
	if actualScore != expectedScore {
		t.Fail()
	}
}
