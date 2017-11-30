package riotapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// apiHosts is a map from regions to Riot hosts
var apiHosts = map[string]string{
	"br":   "br1.api.riotgames.com",
	"eune": "eun1.api.riotgames.com",
	"euw":  "euw1.api.riotgames.com",
	"jp":   "jp1.api.riotgames.com",
	"kr":   "kr.api.riotgames.com",
	"lan":  "la1.api.riotgames.com",
	"las":  "la2.api.riotgames.com",
	"na":   "na1.api.riotgames.com",
	"oce":  "oc1.api.riotgames.com",
	"tr":   "tr1.api.riotgames.com",
	"ru":   "ru.api.riotgames.com",
	"pbe":  "pbe1.api.riotgames.com",
}

// Client is the http client used for sending the requests
type Client struct {
	c         *http.Client
	host      string
	apiKey    string
	Summoner  SummonerAPI
	Spectator SpectatorAPI
	Static    LoLStaticDataAPI
}

// New creates a new riot API client
func New(apiKey, region string) (*Client, error) {
	host, ok := apiHosts[region]
	if !ok {
		return nil, errors.New("invalid region")
	}
	c := &Client{
		c:      &http.Client{Timeout: time.Second * 20},
		host:   host,
		apiKey: apiKey,
	}

	c.Summoner = SummonerAPI{c: c}
	c.Spectator = SpectatorAPI{c: c}
	c.Static = LoLStaticDataAPI{c: c}
	return c, nil
}

// APIError wraps http error messages
type APIError struct {
	StatusCode int
	Msg        string
}

func (err APIError) Error() string {
	return fmt.Sprintf("status: %d, msg: %s", err.StatusCode, err.Msg)
}

// Request sends a new request to the given api endpoint and unmarshalls the response to given data
func (c *Client) Request(api, apiMethod string, data interface{}) error {
	q := url.Values{}
	q.Add("api_key", c.apiKey)

	u := url.URL{
		Host:     c.host,
		Scheme:   "https",
		Path:     fmt.Sprintf("lol/%s/v3/%s", api, apiMethod),
		RawQuery: q.Encode(),
	}

	fmt.Println(u.String())

	resp, err := c.c.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleErrorStatus(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return err
	}
	return nil
}

func handleErrorStatus(resp *http.Response) error {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("unable to read response body")
	}
	return APIError{StatusCode: resp.StatusCode, Msg: string(b)}
}
