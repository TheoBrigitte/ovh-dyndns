package client

import (
	"fmt"
)

// UpdateRecord update DNS record.
func (c Client) UpdateRecord(zone string, id int64, record DNSRecord) error {
	err := c.client.Put(fmt.Sprintf("/domain/zone/%s/record/%d", zone, id), record, nil)
	if err != nil {
		return err
	}

	return c.RefreshZone(zone)
}

// RefreshZone apply changes to DNS zone.
func (c Client) RefreshZone(zone string) error {
	err := c.client.Post(fmt.Sprintf("/domain/zone/%s/refresh", zone), nil, nil)
	if err != nil {
		return err
	}

	return nil
}
