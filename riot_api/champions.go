package main

import (
	"encoding/json"
	"io/ioutil"
)

// ChampionsDTO contains the Champions data
type ChampionsDTO struct {
	Type    string
	Version string
	Data    map[string]Champion
}

// Champions contains the Champions indexed by their id's
type Champions struct {
	Data map[int]Champion
}

// Champion is the League of Legends champion
type Champion struct {
	ID    int
	Name  string
	Title string
	Key   string
}

func (champs *Champions) fromDTO(champsDTO *ChampionsDTO) {
	champs.Data = make(map[int]Champion, len(champsDTO.Data))
	for _, v := range champsDTO.Data {
		champs.Data[v.ID] = v
	}
}

func readChamps(fn string) (*ChampionsDTO, error) {
	champsJSON, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	var champs ChampionsDTO
	err = json.Unmarshal(champsJSON, &champs)
	if err != nil {
		return nil, err
	}

	return &champs, nil
}
