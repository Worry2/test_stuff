package riotapi

import (
	"testing"
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

func TestGetChampions(t *testing.T) {
	t.Skip("This API should not be called too often")

	api := LoLStaticDataAPI{newClient(t)}
	champs, err := api.Champions()
	if err != nil {
		t.Fatalf("unable to get champions: %v", err)
	}

	if len(champs.Data) < 100 {
		t.Fatal("not enought champions")
	}
}
