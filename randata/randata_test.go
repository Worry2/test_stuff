package randata

import (
	"fmt"
	"strconv"
	"testing"
)

func TestWordsInitialize(t *testing.T) {
	Initialize()
	fmt.Println("Total words: " + strconv.Itoa(len(words)))
	fmt.Println("Total countries: " + strconv.Itoa(len(countries)))

	for i := 0; i < 10; i++ {
		fmt.Printf("Runes: %s\n", RandStringRunes(i))
	}
	for i := 0; i < 10; i++ {
		fmt.Printf("Word: %s\n", GetRandomWord())
	}
	for i := 0; i < 10; i++ {
		fmt.Printf("3 Words: %s\n", GetThreeWords())
	}
	for i := 0; i < 10; i++ {
		fmt.Printf("Country: %s\n", GetRandomCountry())
	}
}
