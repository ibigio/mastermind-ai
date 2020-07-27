package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type CodePeg int
type ScorePeg int
type Code []CodePeg
type SecretCode Code
type GuessCode Code

type Score struct {
	red   int
	white int
}

const (
	red CodePeg = iota
	orange
	yellow
	green
	blue
	purple
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

// func (GuessCode) String() string {
// 	return "code!"
// }

func codesEqual(a, b Code) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func calculateScore(secret SecretCode, guess GuessCode) Score {
	if len(secret) != len(guess) {
		panic("secret and guess different lengths")
	}

	usedSecretPegs := make([]bool, len(guess))
	usedGuessPegs := make([]bool, len(guess))
	score := Score{}

	// determine reds
	for i, s := range secret {
		if guess[i] == s {
			score.red++
			usedSecretPegs[i] = true
			usedGuessPegs[i] = true
		}
	}

	// determine whites
	for i, s := range secret {
		// skip used secret pegs
		if usedSecretPegs[i] {
			continue
		}
		// check if guess peg exists
		for j, g := range guess {
			// skip used guess pegs
			if usedGuessPegs[j] {
				continue
			}
			if s == g {
				usedSecretPegs[i] = true
				usedGuessPegs[j] = true
				score.white++
				break
			}
		}
	}

	return score
}

func generateAllPossibleCodes(colors []CodePeg, length int) []Code {
	base := len(colors)
	numPossibilities := int(math.Pow(float64(base), float64(length)))
	possibileCodes := make([]Code, numPossibilities)

	for i := range possibileCodes {
		perm := strconv.FormatInt(int64(i), base)
		newCode := make(Code, length)
		for j, c := range perm {
			index, _ := strconv.Atoi(string(c))
			newCode[(length-len(perm))+j] = colors[index]
			possibileCodes[i] = newCode
		}
	}
	return possibileCodes
}

func countImplausibleSecrets(guess GuessCode, score Score, secrets []SecretCode) int {
	i := 0
	for _, secret := range secrets {
		if calculateScore(secret, guess) != score {
			i++
		}
	}
	return i
}

func discardImplausibleSecrets(guess GuessCode, score Score, secrets []SecretCode) []SecretCode {
	plausibleSecrets := make([]SecretCode, 0)
	for _, secret := range secrets {
		if calculateScore(secret, guess) == score {
			plausibleSecrets = append(plausibleSecrets, secret)
		}
	}
	return plausibleSecrets
}

func determineGuessQuality(guess GuessCode, secrets []SecretCode) int {
	return getExpectedDiscard(guess, secrets)
}

func getExpectedDiscard(guess GuessCode, secrets []SecretCode) int {
	totalDiscard := 0
	// determine possible score frequencies
	scoreFreqs := make(map[Score]int)
	for _, secret := range secrets {
		score := calculateScore(secret, guess)
		if _, ok := scoreFreqs[score]; !ok {
			scoreFreqs[score] = 0
		}
		scoreFreqs[score]++
	}

	// determine discarded per score
	for score, freq := range scoreFreqs {
		discarded := countImplausibleSecrets(guess, score, secrets)
		totalDiscard += discarded * freq
	}
	return totalDiscard / len(secrets)
}

func generateChunks(guesses []GuessCode, numChunks int) [][]GuessCode {
	minChunkSize := len(guesses) / numChunks
	leftovers := len(guesses) % numChunks

	chunks := make([][]GuessCode, numChunks)

	index := 0
	for i := 0; i < numChunks; i++ {
		chunkSize := minChunkSize
		// evenly distribute leftover work
		if leftovers > 0 {
			chunkSize++
			leftovers--
		}
		chunks[i] = guesses[index : index+chunkSize]
		index = index + chunkSize // set next chunk start
	}

	return chunks
}

func calculateBestGuessParallel(guesses []GuessCode, secrets []SecretCode, numWorkers int) (bestGuess GuessCode, bestGuessQuality int) {

	bestGuessesChan := make(chan GuessCode)

	chunks := generateChunks(guesses, numWorkers)
	for i := 0; i < numWorkers; i++ {
		go calculateBestGuessWorkerAsync(chunks[i], secrets, bestGuessesChan)
	}

	// accumulate guesses and determine best
	bestGuessQuality = math.MinInt64
	for i := 0; i < numWorkers; i++ {
		guess := <-bestGuessesChan
		quality := determineGuessQuality(guess, secrets)
		if quality > bestGuessQuality {
			bestGuess = guess
			bestGuessQuality = quality
		}
	}

	return
}

func calculateBestGuessWorkerAsync(guesses []GuessCode, secrets []SecretCode, guessChan chan GuessCode) {
	guess, _ := calculateBestGuess(guesses, secrets)
	guessChan <- guess
}

func calculateBestGuess(guesses []GuessCode, secrets []SecretCode) (bestGuess GuessCode, bestGuessQuality int) {
	// currently could make a non-plausible guess
	// to only guess plausible guesses, switch "guesses" to "secrets"
	bestGuesses := make([]GuessCode, 0)
	bestGuessQuality = math.MinInt64
	for _, guess := range guesses {
		quality := determineGuessQuality(guess, secrets)
		if quality >= bestGuessQuality {
			// if strictly better, create new set
			if quality > bestGuessQuality {
				bestGuesses = make([]GuessCode, 0)
			}
			bestGuessQuality = quality
			bestGuesses = append(bestGuesses, guess)
		}
	}
	index := rand.Intn(len(bestGuesses))
	bestGuess = bestGuesses[index]
	return
}
