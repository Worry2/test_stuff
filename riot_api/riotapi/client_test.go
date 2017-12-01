package riotapi

import "testing"

const aPIKey = "RGAPI-abcf80c0-6f71-4a65-acf5-2f03e087dd07"

func newClient(t *testing.T) *Client {
	c, err := New(aPIKey, "eune", 50, 20)
	if err != nil {
		t.Fatalf("unable to create client: %v", err)
	}
	return c
}
