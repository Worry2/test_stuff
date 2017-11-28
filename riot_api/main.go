package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	aPIKey            = "?api_key=RGAPI-5f81076a-6d93-46c4-9665-xxxxxxx"
	summonerByNameURL = "https://eun1.api.riotgames.com/lol/summoner/v3/summoners/by-name/"
	spectatorURL      = "https://eun1.api.riotgames.com/lol/spectator/v3/active-games/by-summoner/"
	championURL       = "https://eun1.api.riotgames.com/lol/static-data/v3/champions/"

	discordHook = "https://discordapp.com/api/webhooks/384835834815447052/K6amIwt30YVjWBJFivZIR8UIBB8Qh-mUGcleVUQ0oTSTt5BJuR0eXRKZ1xJyqEmEzscF"

	avatarURL = "http://ddragon.leagueoflegends.com/cdn/img/champion/tiles/gangplank_0.jpg"
)

var players = []*Player{
	{Name: "Uxipaxa", ID: 24749077},
	{Name: "Invataxi", ID: 31507600},
	{Name: "Ignusnus", ID: 25251553},
	{Name: "Opettaja", ID: 28490422},
}

var urlBusy = make(map[string]time.Time)

// Player is a lol player
type Player struct {
	Name     string
	ID       int
	InGame   bool
	Champion string
}

type discordMsg struct {
	Content   string `json:"content"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type discordEmbed struct {
}

func main() {
	//readPlayerIDs()
	// monitorPlayers()
	imageToDiscord("Teemo")
}

func monitorPlayers() {
	for {
		for _, p := range players {
			handleMonitorPlayer(p)
		}
		time.Sleep(time.Second * 60)
	}
}

func handleMonitorPlayer(p *Player) {
	r, err := getActiveGames(p.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if r["status"] != http.StatusOK {
		if p.InGame {
			sendToDiscord(p.Name + " lopetti pelin")
		}
		p.InGame = false
		return
	}

	if p.InGame {
		return
	}

	participants := r["participants"].([]interface{})
	for _, participant := range participants {
		pmap := participant.(map[string]interface{})
		summID, err := pmap["summonerId"].(json.Number).Int64()
		if err != nil {
			fmt.Println(err)
			return
		}
		if int(summID) == p.ID {
			champID, err := pmap["championId"].(json.Number).Int64()
			champ, err := getChampionData(int(champID))
			if err != nil {
				fmt.Println("unable to get champion data: ", err)
				return
			}
			p.Champion = fmt.Sprintf("%v, %v\n", champ["name"], champ["title"])
		}
	}

	if !p.InGame {
		sendToDiscord(fmt.Sprintf("%v meni peliin hahmolla %v", p.Name, p.Champion))
	}
	p.InGame = true

}

func requestRIOT(url string) (map[string]interface{}, error) {
	if !urlBusy[url].IsZero() && time.Now().After(urlBusy[url]) {
		return nil, errors.New("too many requests to " + url)
	}
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var respMap map[string]interface{}
	d := json.NewDecoder(r.Body)
	d.UseNumber()
	if err := d.Decode(&respMap); err != nil {
		return nil, err
	}
	respMap["status"] = r.StatusCode

	if r.StatusCode == http.StatusTooManyRequests {
		rafTime, err := strconv.Atoi(r.Header.Get("retry-after"))
		if err != nil {
			return nil, fmt.Errorf("unable to read retry-after: %v", err)
		}
		fmt.Printf("Rate limit exceeded! set endpoint unusable for %v seconds\n", rafTime)
		urlBusy[url] = time.Now().Add(time.Second * time.Duration(rafTime))
		return nil, errors.New("too many requests")
	}
	return respMap, nil
}

func getActiveGames(id int) (map[string]interface{}, error) {
	return getIDData(spectatorURL, id)
}

func getChampionData(id int) (map[string]interface{}, error) {
	return getIDData(championURL, id)
}

func getIDData(url string, id int) (map[string]interface{}, error) {
	reqS := fmt.Sprintf("%s%d%s", url, id, aPIKey)
	fmt.Println("requesting: ", reqS)
	resp, err := requestRIOT(reqS)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func readPlayerIDs() {
	for i, p := range players {
		id, err := getPlayerID(p.Name)
		if err != nil {
			panic(err)
		}
		players[i].ID = id
	}
	fmt.Println(players)
	time.Sleep(time.Second * 10)
}

func getPlayerID(name string) (int, error) {
	respMap, err := requestRIOT(summonerByNameURL + name + aPIKey)
	if err != nil {
		return -1, err
	}
	id, err := respMap["id"].(json.Number).Int64()
	if err != nil {
		return -1, err
	}
	fmt.Printf("%s : %d\n", name, id)
	return int(id), nil
}

func sendToDiscord(s string) {
	fmt.Println("Lähetys: ", s)
	fmt.Println(s)
	dm := discordMsg{
		Content:   fmt.Sprintf("%s", s),
		AvatarURL: avatarURL,
	}

	b, err := json.Marshal(dm)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(discordHook, "application/json", bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func printAndReadResponse(r *http.Response) string {
	s, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("status: %d, body: %s\n", r.StatusCode, s)
	return string(s)
}

func imageToDiscord(s string) {
	dm := discordgo.WebhookParams{
		AvatarURL: avatarURL,
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: s,
				Image: &discordgo.MessageEmbedImage{URL: "http://ddragon.leagueoflegends.com/cdn/img/champion/tiles/" + strings.ToLower(s) + "_0.jpg"},
			},
		},
	}

	b, err := json.Marshal(dm)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(discordHook, "application/json", bytes.NewReader(b))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}
