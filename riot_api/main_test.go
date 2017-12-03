package main

import (
	"testing"

	"github.com/tahkapaa/test_stuff/riot_api/riotapi"
)

func TestFindPlayerRank(t *testing.T) {
	c, err := riotapi.New(apiKey, "eune", 50, 20)
	if err != nil {
		t.Fatal(err)
	}
	s, err := findPlayerRank(c, 29652836)
	if err != nil {
		t.Fatal(err)
	}

	if s == "Not found" {
		t.Errorf("invalid response: %s", s)
	}
}
