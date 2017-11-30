package riotapi

import "errors"

// SummonerAPI implements the Riot Summoner API methods
type SummonerAPI struct {
	c *Client
}

const summonerAPIPath = "summoner"

// SummonerDTO represents a summoner
type SummonerDTO struct {
	Name          string
	ID            int
	AccountID     int
	SummonerLevel int
	ProfileIconID int
	RevisionDate  int
}

// SummonerByName gets a summoner by summoner name
func (api SummonerAPI) SummonerByName(name string) (*SummonerDTO, error) {
	if name == "" {
		return nil, errors.New("missing summoner name")
	}
	var s SummonerDTO
	if err := api.c.Request(summonerAPIPath, "summoners/by-name/"+name, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
