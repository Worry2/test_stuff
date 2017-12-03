package riotapi

import "testing"

const aPIKey = "RGAPI-8b46dbb2-825e-47d3-8b42-6065baaf018f"

func newClient(t *testing.T) *Client {
	c, err := New(aPIKey, "eune", 50, 20)
	if err != nil {
		t.Fatalf("unable to create client: %v", err)
	}
	return c
}
