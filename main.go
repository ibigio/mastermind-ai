package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sort"
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

func sum(array []int) int {
	result := 0
	for _, v := range array {
		result += v
	}
	return result
}

func main() {

	// configuration setup
	rand.Seed(time.Now().UnixNano())
	numWorkers := runtime.NumCPU()
	numPegs := 4
	allColors := []CodePeg{red, orange, yellow, green, blue, purple}

	// play an interactive game
	printUsage()
	playInteractive(allColors, numPegs, numWorkers)
	fmt.Println()

	// run evaluation
	numGames := 100
	// runEvaluation(numGames, allColors, numWorkers, numPegs, true)
	runSimulations(numGames, allColors, numWorkers, numPegs, true)
}

func printUsage() {
	fmt.Println("Choose your code. Enter each score like so:")
	fmt.Println("0 red 2 white")
	fmt.Println("Enjoy!")
	fmt.Println()
}

func runEvaluation(numGames int, allColors []CodePeg, numWorkers int, numPegs int, optimize bool) (avgGuesses float32, avgTime time.Duration) {
	fmt.Printf("Evaluation === %v games, %v workers, optimize is %v\n", numGames, numWorkers, optimize)
	totalGuesses := 0
	totalTime := time.Duration(0)
	for i := 0; i < numGames; i++ {
		start := time.Now()
		totalGuesses += selfPlayRandom(allColors, numPegs, numWorkers, false, optimize)
		totalTime += time.Since(start)
	}
	avgGuesses = float32(totalGuesses) / float32(numGames)
	avgTime = totalTime / time.Duration(totalGuesses)
	fmt.Println("Avg Guesses/Game:", avgGuesses)
	fmt.Println("Avg Time/Guess:", avgTime.Round(time.Millisecond))
	fmt.Println()
	return
}

func selfPlayRandom(allColors []CodePeg, numPegs int, numWorkers int, logging bool, optimize bool) (totalGuesses int) {
	// choose random secret
	secret := SecretCode(randomCode(allColors, numPegs))
	return selfPlay(allColors, secret, numPegs, numWorkers, logging, optimize)
}

func selfPlay(allColors []CodePeg, secret SecretCode, numPegs int, numWorkers int, logging bool, optimize bool) (totalGuesses int) {

	// initialize
	secrets, guesses := initializeSecrets(allColors, numPegs)
	perfectScore := Score{red: numPegs}
	isFirstGuess := true

	// guess and discard until finished
	for true {
		var guess GuessCode
		var quality int

		if isFirstGuess && optimize {
			guess = firstGuessStrategey(guesses)
			isFirstGuess = false
		} else {
			guess, quality = calculateBestGuessParallel(guesses, secrets, numWorkers)
		}

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

	fmt.Printf("Guessed after %v guesses!\n", totalGuesses)
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

func randomSecretCodeUnderTest(allColors []CodePeg, numPegs int) SecretCode {
	return SecretCode{red, red, yellow, green}
}

func histogram(guessesPerGame []int, width int) {
	scale := float32(width) / float32(len(guessesPerGame))
	fmt.Println("Histogram")
	freqs := make(map[int]int)
	for _, guess := range guessesPerGame {
		freqs[guess]++
	}
	// sort by frequency
	var keys []int
	for k := range freqs {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		guess := k
		freq := freqs[k]
		bars := int(float32(freq) * scale)
		if bars == 0 {
			bars = 1
		}
		// horizontally print a sequence of '*'s to represent the frequency of the guess
		// print percentage
		percent := fmt.Sprintf("%.1f%%", float32(freq)/float32(len(guessesPerGame))*100)
		fmt.Printf("%v: %v ", guess, percent)
		// print gap to align all bars
		for i := 0; i < 8-len(percent)-len(strconv.Itoa(guess)); i++ {
			fmt.Print(" ")
		}
		for i := 0; i < bars; i++ {
			fmt.Print("*")
		}
		fmt.Println()
	}
}

func runSimulations(numGames int, allColors []CodePeg, numWorkers int, numPegs int, optimize bool) (avgGuesses float32, avgTime time.Duration) {
	fmt.Printf("Simulations === %v games, %v workers, optimize is %v\n", numGames, numWorkers, optimize)
	guessesPerGame := make([]int, numGames)
	totalTime := time.Duration(0)
	for i := 0; i < numGames; i++ {
		secret := randomSecretCodeUnderTest(allColors, numPegs)
		start := time.Now()
		// totalGuesses += selfPlay(allColors, secret, numPegs, numWorkers, false, optimize)
		guessesPerGame[i] = selfPlay(allColors, secret, numPegs, numWorkers, false, optimize)
		totalTime += time.Since(start)
	}
	totalGuesses := sum(guessesPerGame)
	avgGuesses = float32(totalGuesses) / float32(numGames)
	avgTime = totalTime / time.Duration(totalGuesses)
	fmt.Println("Avg Guesses/Game:", avgGuesses)
	fmt.Println("Avg Time/Guess:", avgTime.Round(time.Millisecond))
	fmt.Println()
	histogram(guessesPerGame, 50)
	return
}
