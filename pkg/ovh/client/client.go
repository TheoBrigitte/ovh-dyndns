package client

import (
	"github.com/ovh/go-ovh/ovh"
)

// Client extends ovh.Client.
type Client struct {
	client ovh.Client
}

// New instantiate a new Client.
func New(client ovh.Client) Client {
	return Client{
		client: client,
	}
}
