package main

import (
	"fmt"
	"testing"

	"github.com/RobWC/riotapi"
)

func TestParseChampions(t *testing.T) {
	c, err := readChamps("champions.json")
	if err != nil {
		t.Errorf("readChamps failed: %v", err)
	}

	for k, v := range c.Data {
		if k != v.Key {
			t.Errorf("Unexpected %v, expected %v in %v", k, v.Key, v)
		}
	}
}

func TestFromChampionsDTO(t *testing.T) {
	cdto, err := readChamps("champions.json")
	if err != nil {
		t.Errorf("readChamps failed: %v", err)
	}

	var champs Champions
	champs.fromDTO(cdto)

	for _, v := range champs.Data {
		if v.Name != cdto.Data[v.Key].Name {
			t.Errorf("Unexpected %v, expected %v", v.Name, cdto.Data[v.Key].Name)
		}
	}
}

// https://eun1.api.riotgames.com/lol/static-data/v3/champions/

func TestRiotAPI(t *testing.T) {
	apiKeyEnv := "RGAPI-fd5d6135-b0a1-4099-9d8e-4444af580022"

	if apiKeyEnv == "" {
		t.Fatal("API Key Not Specified")
	}

	c := riotapi.NewAPIClient("eun1", apiKeyEnv)
	sl, err := c.StaticChampions("v3", riotapi.ChampDataAll)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(sl)
	t.Log(sl)

}
