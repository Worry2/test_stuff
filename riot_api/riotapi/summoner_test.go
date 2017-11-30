package riotapi

import (
	"testing"
)

func TestSummonerByName(t *testing.T) {
	api := SummonerAPI{newClient(t)}
	uxi, err := api.SummonerByName("uxipaxa")
	if err != nil {
		t.Fatalf("unable to get summoner: %v", err)
	}
	if uxi.ID != 24749077 {
		t.Fatalf("invalid summoner id: %d", uxi.AccountID)
	}
}
