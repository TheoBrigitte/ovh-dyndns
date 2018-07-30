package client

import (
	"github.com/ovh/go-ovh/ovh"
)

// GenerateConsumerKey request a new customer key with following permissions:
// 	 GET /me
// 	 GET, POST, PUT /domain/*
// Make sure to visit response.ValidationURL url to activate the consumer key.
func (c Client) GenerateConsumerKey() (*ovh.CkValidationState, error) {
	ckReq := c.client.NewCkRequest()

	// Allow GET method on /me
	ckReq.AddRules(ovh.ReadOnly, "/me")

	// Allow GET, POST, PUT method on /domain and all its sub routes
	ckReq.AddRecursiveRules(ovh.ReadWriteSafe, "/domain")

	// Run the request
	response, err := ckReq.Do()
	if err != nil {
		return nil, err
	}

	return response, nil
}
