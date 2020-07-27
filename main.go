package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func (c CodePeg) String() string {
	switch c {
	case red:
		return "red"
	case orange:
		return "orange"
	case yellow:
		return "yellow"
	case green:
		return "green"
	case blue:
		return "blue"
	case purple:
		return "purple"
	}
	return "invalid"
}

func main() {

	rand.Seed(time.Now().UnixNano())

	allColors := []CodePeg{red, orange, yellow, green, blue, purple}
	playInteractive(allColors, 4, 4)
}

func playInteractive(allColors []CodePeg, numPegs int, numWorkers int) {

	// generate initial guesses and secrets
	allCodes := generateAllPossibleCodes(allColors, 4)
	secrets := make([]SecretCode, len(allCodes))
	guesses := make([]GuessCode, len(allCodes))
	for i := range allCodes {
		secrets[i] = SecretCode(allCodes[i])
		guesses[i] = GuessCode(allCodes[i])
	}

	perfectScore := Score{red: numPegs}

	// guess and discard until finished
	totalGuesses := 0
	for len(secrets) != 1 {
		guess, _ := calculateBestGuessParallel(guesses, secrets, numWorkers)
		fmt.Println(guess)
		totalGuesses++

		score := getScore()
		if score == perfectScore {
			break
		}
		secrets = discardImplausibleSecrets(guess, score, secrets)

		if len(secrets) == 0 {
			panic("All secrets discarded. Double check scores.")
		}
	}

	fmt.Printf("Guess after %v guesses!\n", totalGuesses)
}

func getScore() Score {
	score, err := readScore()
	for err != nil {
		fmt.Println(err.Error())
		score, err = readScore()
	}
	fmt.Println(score)
	return score
}

func readScore() (Score, error) {
	reader := bufio.NewReader(os.Stdin)
	defaultError := errors.New("Please enter a valid score. (e.g. 0 red 2 white)")

	// read score
	fmt.Printf("Enter score: ")
	scoreText, err := reader.ReadString('\n')
	scoreText = strings.Replace(scoreText, "\n", "", -1)
	if err != nil {
		return Score{}, defaultError
	}
	// parse score
	scoreWords := strings.Split(scoreText, " ")
	if len(scoreWords) != 4 {
		return Score{}, defaultError
	}
	if strings.ToLower(scoreWords[1]) != "red" || strings.ToLower(scoreWords[3]) != "white" {
		return Score{}, defaultError
	}
	numRed, err := strconv.Atoi(scoreWords[0])
	if err != nil {
		return Score{}, defaultError
	}
	numWhite, err := strconv.Atoi(scoreWords[2])
	if err != nil {
		return Score{}, defaultError
	}
	if numRed < 0 || numWhite < 0 || numRed+numWhite > 4 { // TODO: remove hard-coded max amount
		return Score{}, defaultError
	}

	return Score{numRed, numWhite}, nil
}
