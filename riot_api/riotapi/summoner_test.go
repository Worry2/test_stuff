package riotapi

import (
	"testing"
)

func TestSummonerByName(t *testing.T) {
	c, err := New(aPIKey, "eune")
	if err != nil {
		t.Fatalf("unable to create client: %v", err)
	}
	api := SummonerAPI{c}
	uxi, err := api.SummonerByName("uxipaxa")
	if err != nil {
		t.Fatalf("unable to get champions: %v", err)
	}
	if uxi.ID != 24749077 {
		t.Fatalf("invalid summoner id: %d", uxi.AccountID)
	}
}
