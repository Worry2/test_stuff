package randata

import (
	"math/rand"
	"strings"
	"time"

	"github.com/tahkapaa/test_stuff/randata/data"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var words = stringToArray(data.WordsStr)
var countries = stringToArray(data.CountryStr)
var initialized = false

// Initialize opens the file and reads it to memory
func Initialize() {
	rand.Seed(time.Now().UnixNano())
}

func stringToArray(str string) []string {
	return strings.Split(str, "\n")
}

// GetRandomWord returns a random english noun
func GetRandomWord() string {
	return words[rand.Intn(len(words))]
}

// GetThreeWords returns three random nouns separated by a space
func GetThreeWords() string {
	return GetRandomWord() + " " + GetRandomWord() + " " + GetRandomWord()
}

// GetRandomCountry returns a random country
func GetRandomCountry() string {
	return countries[rand.Intn(len(countries))]
}

// RandStringRunes returns a random letter
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
