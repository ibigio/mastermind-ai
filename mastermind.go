package main

import (
	"math"
	"math/rand"
	"strconv"
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

type QualifiedGuess struct {
	guess   GuessCode
	quality int
}

const (
	red CodePeg = iota
	orange
	yellow
	green
	blue
	purple
)

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

func initializeSecrets(allColors []CodePeg, numPegs int) (secrets []SecretCode, guesses []GuessCode) {
	allCodes := generateAllPossibleCodes(allColors, numPegs)
	secrets = make([]SecretCode, len(allCodes))
	guesses = make([]GuessCode, len(allCodes))
	for i := range allCodes {
		secrets[i] = SecretCode(allCodes[i])
		guesses[i] = GuessCode(allCodes[i])
	}
	return
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

	// if few choices left, take a gander
	guessThreshold := 3
	if len(secrets) <= guessThreshold {
		return GuessCode(secrets[rand.Intn(len(secrets))]), 1
	}

	bestGuessesChan := make(chan QualifiedGuess)

	chunks := generateChunks(guesses, numWorkers)
	for i := 0; i < numWorkers; i++ {
		go calculateBestGuessWorkerAsync(chunks[i], secrets, bestGuessesChan)
	}

	// accumulate guesses and determine best
	bestGuessQuality = math.MinInt64
	for i := 0; i < numWorkers; i++ {
		qualifiedGuess := <-bestGuessesChan
		guess := qualifiedGuess.guess
		quality := qualifiedGuess.quality
		if quality > bestGuessQuality {
			bestGuess = guess
			bestGuessQuality = quality
		}
	}

	return
}

func calculateBestGuessWorkerAsync(guesses []GuessCode, secrets []SecretCode, guessChan chan QualifiedGuess) {
	guess, quality := calculateBestGuess(guesses, secrets)
	guessChan <- QualifiedGuess{guess, quality}
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
