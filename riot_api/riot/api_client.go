package riotapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

/*
API End Points

BR	BR1	br.api.pvp.net
EUNE	EUN1	eune.api.pvp.net
EUW	EUW1	euw.api.pvp.net
KR	KR	kr.api.pvp.net
LAN	LA1	lan.api.pvp.net
LAS	LA2	las.api.pvp.net
NA	NA1	na.api.pvp.net
OCE	OC1	oce.api.pvp.net
TR	TR1	tr.api.pvp.net
RU	RU	ru.api.pvp.net
PBE	PBE1	pbe.api.pvp.net
*/

// APIEndpoints mapping to api endpoints
var APIEndpoints = map[string]string{
	"br":   "br.api.pvp.net",
	"eun1": "eun1.api.riotgames.com",
	"euw":  "euw.api.pvp.net",
	"kr":   "kr.api.pvp.net",
	"lan":  "lan.api.pvp.net",
	"las":  "las.api.pvp.net",
	"na":   "na.api.pvp.net",
	"oce":  "oce.api.pvp.net",
	"tr":   "tr.api.pvp.net",
	"ru":   "ru.api.pvp.net",
	"jp":   "jp.api.pvp.net",
}

// ShardName a mapping of regions to shard names
var ShardName = map[string]string{
	"br":   "BR1",
	"eune": "EUN1",
	"euw":  "EUW1",
	"kr":   "KR",
	"lan":  "LA1",
	"las":  "LA2",
	"na":   "NA1",
	"oce":  "OC1",
	"tr":   "TR1",
	"ru":   "RU",
	"jp":   "JP1",
}

// RateLimit the current rate limit of the api
type RateLimit struct {
	LimitType            map[string]string // X-Rate-Limit-Type
	RetryAfter           int               // Retry-After
	AppRateLimitCount    map[string]int    // X-App-Rate-Limit-Count
	MethodRateLimitCount map[string]int    // X-Method-Rate-Limit-Count

	NextAttempt time.Time
	Limits      map[string]string
}

// APIClient Riot API client
type APIClient struct {
	endpoint      string
	game          string
	region        string
	client        *http.Client
	totalRequests int
	key           string // API key
	tokens        chan struct{}
	RateLimit     *RateLimit
	shardName     string
}

// NewAPIClient create an initalized APIClient
func NewAPIClient(region, key string) *APIClient {
	c := &APIClient{
		key:       key,
		game:      "lol",
		region:    strings.ToLower(region),
		shardName: ShardName[strings.ToLower(region)],
		client: &http.Client{
			Jar:     nil,
			Timeout: time.Second * 5,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				Dial: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 10 * time.Second,
			}},
		tokens:    make(chan struct{}, 20),
		RateLimit: &RateLimit{Limits: make(map[string]string)},
	}

	return c
}

func (c *APIClient) genURL(path []string) string {
	return strings.Join(path, "/")
}

// https://eun1.api.riotgames.com/lol/static-data/v3/champions/
func (c *APIClient) genStaticRequest(method, version, api string, query url.Values) (*http.Request, error) {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = APIEndpoints[c.region]
	u.Path = fmt.Sprintf("/%s/static-data/%s/%s", c.game, version, api)
	u.RawQuery = query.Encode()
	return http.NewRequest(method, u.String(), nil)
}

func (c *APIClient) genRequest(method, version, api string, query url.Values) (*http.Request, error) {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = APIEndpoints[c.region]
	u.Path = fmt.Sprintf("/api/%s/%s/%s/%s", c.game, c.region, version, api)
	u.RawQuery = query.Encode()
	return http.NewRequest(method, u.String(), nil)
}

// do execute a request
func (c *APIClient) do(req *http.Request, apiKey bool) ([]byte, error) {
	// add api key

	// manipulate request
	if apiKey {
		query := req.URL.Query()
		query.Add("api_key", c.key)

		//rebuild request
		req.URL.RawQuery = query.Encode()
		fmt.Println(req.URL)
	}
	if c.RateLimit.RetryAfter > 0 {
		time.Sleep(time.Second * time.Duration(c.RateLimit.RetryAfter))
	}

	rc := make(chan struct {
		data []byte
		err  error
	})

	go func(req http.Request) {
		defer func() {
			<-c.tokens
		}()

		// check for rate limiting
		// check retry limit
		// set a timer to wait for acceptable time

		if !time.Now().After(c.RateLimit.NextAttempt) {
			// Wait for the next attempt
			t := time.NewTimer(time.Since(c.RateLimit.NextAttempt))
			<-t.C
		}

		c.tokens <- struct{}{}
		resp, err := c.client.Do(&req)
		if err != nil {
			rc <- struct {
				data []byte
				err  error
			}{nil, err}
		}
		defer resp.Body.Close()

		// Update rate limiting
		// https://developer.riotgames.com/rate-limiting.html
		rl := resp.Header.Get("X-Rate-Limit-Count")
		if rl != "" {
			c.updateRateLimits(rl)
		}

		log.Println(resp.Header.Get("X-App-Rate-Limit-Count"))

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			rc <- struct {
				data []byte
				err  error
			}{nil, err}
		}

		switch resp.StatusCode {
		case http.StatusTooManyRequests:

			//ra := resp.Header.Get("Retry-After")
			//c.RateLimit.NextAttempt = time.Now().Add(time.Second * time.Duration(ra))

		}

		if resp.StatusCode != http.StatusOK {
			rc <- struct {
				data []byte
				err  error
			}{nil, fmt.Errorf("API Error %s", string(data))}
		}

		rc <- struct {
			data []byte
			err  error
		}{data, err}

	}(*req)

	r := <-rc
	return r.data, r.err
}

func (c *APIClient) makeRequest(method, version, api string, query url.Values, auth bool, data interface{}) error {
	req, err := c.genRequest(method, version, api, query)
	if err != nil {
		return err
	}

	respData, err := c.do(req, auth)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respData, &data)
	if err != nil {
		return err
	}

	return err
}

func (c *APIClient) makeStaticRequest(method, version, api string, query url.Values, auth bool, data interface{}) error {
	req, err := c.genStaticRequest(method, version, api, query)
	if err != nil {
		return err
	}

	respData, err := c.do(req, auth)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respData, &data)
	if err != nil {
		return err
	}
	return err
}
