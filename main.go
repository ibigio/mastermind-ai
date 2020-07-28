package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"runtime"
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

	// configuration setup
	rand.Seed(time.Now().UnixNano())
	numWorkers := runtime.NumCPU()
	numPegs := 4
	allColors := []CodePeg{red, orange, yellow, green, blue, purple}

	// play an interactive game
	printUsage()
	playInteractive(allColors, 4, 4)
	fmt.Println()

	// run evaluation
	numGames := 20
	runEvaluation(numGames, allColors, numWorkers, numPegs)
}

func printUsage() {
	fmt.Println("Choose your code. Enter each score like so:")
	fmt.Println("0 red 2 white")
	fmt.Println("Enjoy!")
	fmt.Println()
}

func runEvaluation(numGames int, allColors []CodePeg, numWorkers int, numPegs int) (avgGuesses float32, avgTime time.Duration) {
	fmt.Printf("Evaluation === %v games, %v workers\n", numGames, numWorkers)
	totalGuesses := 0
	totalTime := time.Duration(0)
	for i := 0; i < numGames; i++ {
		start := time.Now()
		totalGuesses += selfPlay(allColors, numPegs, numWorkers, false)
		totalTime += time.Since(start)
	}
	avgGuesses = float32(totalGuesses) / float32(numGames)
	avgTime = totalTime / time.Duration(totalGuesses)
	fmt.Println("Avg Guesses/Game:", avgGuesses)
	fmt.Println("Avg Time/Guess:", avgTime.Round(time.Millisecond))
	fmt.Println()
	return
}

func selfPlay(allColors []CodePeg, numPegs int, numWorkers int, logging bool) (totalGuesses int) {

	// initialize
	secrets, guesses := initializeSecrets(allColors, numPegs)
	perfectScore := Score{red: numPegs}

	// choose secret
	secret := secrets[rand.Intn(len(secrets))]

	// guess and discard until finished
	for true {
		guess, quality := calculateBestGuessParallel(guesses, secrets, numWorkers)
		totalGuesses++

		if logging {
			percent := fmt.Sprintf("%.2f", (float32(quality)/float32(len(secrets)))*100)
			fmt.Printf("%v. %v %v%%\n", totalGuesses, guess, percent)
		}

		score := calculateScore(secret, guess)
		if score == perfectScore {
			break
		}
		secrets = discardImplausibleSecrets(guess, score, secrets)

		if len(secrets) == 0 {
			panic("All secrets discarded. Invalid scoring.")
		}
	}

	return totalGuesses
}

func playInteractive(allColors []CodePeg, numPegs int, numWorkers int) {

	// generate initial guesses and secrets
	secrets, guesses := initializeSecrets(allColors, numPegs)
	perfectScore := Score{red: numPegs}

	// guess and discard until finished
	totalGuesses := 0
	for true {
		guess, _ := calculateBestGuessParallel(guesses, secrets, numWorkers)
		totalGuesses++
		fmt.Printf("%v. %v\n", totalGuesses, guess)

		score := getScore()
		if score == perfectScore {
			break
		}
		secrets = discardImplausibleSecrets(guess, score, secrets)

		if len(secrets) == 0 {
			fmt.Println("No potential secrets remaining. You most likely made a mistake in your scoring.")
			return
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
