package riotapi

import (
	"testing"
)

func TestActiveGamesBySummonerID(t *testing.T) {
	api := SpectatorAPI{newClient(t)}
	_, err := api.ActiveGamesBySummoner(24749077)
	if err != nil {
		t.Fatalf("unable to get active game: %v", err)
	}
}
