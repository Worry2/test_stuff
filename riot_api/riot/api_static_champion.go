package riotapi

import (
	"net/url"
	"strconv"
	"strings"
)

// ChampDataTags tags that provide modifiers of what data is returned on a champion request
type ChampDataTag string

const (
	ChampDataNil         ChampDataTag = ""
	ChampDataAll         ChampDataTag = "all"
	ChampDataAllyTips    ChampDataTag = "allytips"
	ChampDataAltImages   ChampDataTag = "altimages"
	ChampDataBlurb       ChampDataTag = "blurb"
	ChampDataEnemyTips   ChampDataTag = "enemytips"
	ChampDataImage       ChampDataTag = "image"
	ChampDataInfo        ChampDataTag = "info"
	ChampDataLore        ChampDataTag = "lore"
	ChampDataParType     ChampDataTag = "partype"
	ChampDataPassive     ChampDataTag = "passive"
	ChampDataRecommended ChampDataTag = "recommended"
	ChampDataSkins       ChampDataTag = "skins"
	ChampDataSpells      ChampDataTag = "spells"
	ChampDataStats       ChampDataTag = "stats"
	ChampDataTags        ChampDataTag = "tags"
)

// https://developer.riotgames.com/api/methods#!/1055/3633

// StaticChampions get static champion data
// If version is not defined then it uses the current version
func (c *APIClient) StaticChampions(version string, champData ...ChampDataTag) (*ChampionData, error) {
	// setup api query options
	q := url.Values{}
	// if version != "" {
	// 	q.Add("version", version)
	// }
	//
	// if len(champData) > 0 {
	// 	for i := range champData {
	// 		q.Add("champData", string(champData[i]))
	// 	}
	// }
	var cl ChampionData
	err := c.makeStaticRequest("GET", "v3", "champions", q, true, &cl)
	if err != nil {
		return nil, err
	}

	return &cl, err

}

// https://developer.riotgames.com/api/methods#!/1055/3622

// StaticChampionByID get a champion by ID
// If version is not defined then it uses the current version
func (c *APIClient) StaticChampionByID(id int, version string, champData ...ChampDataTag) (cd *Champion, err error) {
	// setup api query options
	q := url.Values{}
	if version != "" {
		q.Add("version", version)
	}

	if len(champData) > 0 {
		for i := range champData {
			q.Add("champData", string(champData[i]))
		}
	}

	err = c.makeStaticRequest("GET", "v1.2", strings.Join([]string{"champion", strconv.Itoa(id)}, "/"), q, true, cd)
	if err != nil {
		return cd, err
	}

	return cd, err
}
