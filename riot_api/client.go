package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var urlBusy = make(map[string]time.Time)

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
