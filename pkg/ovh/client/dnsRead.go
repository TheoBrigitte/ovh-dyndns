package client

import (
	"fmt"
	"net/url"

	"github.com/TheoBrigitte/ovh-dyndns/pkg/ovh/errors"
)

// FindRecord find a DNS record in zone, filter by recordType and subDomain.
// Return DNS record ID.
func (c Client) FindRecord(zone, recordType, subDomain string) (int64, error) {
	IDs, err := c.ListRecords(zone, recordType, subDomain)
	if err != nil {
		return 0, err
	}

	if len(IDs) < 1 {
		return 0, errors.RecordNotFound
	}

	if len(IDs) > 1 {
		return 0, errors.TooManyRecords
	}

	return IDs[0], nil
}

// ListRecords list DNS records in zone, filter by recordType and subDomain.
// Return list of DNS record IDs.
func (c Client) ListRecords(zone, recordType, subDomain string) ([]int64, error) {
	var records []int64

	if zone == "" {
		return nil, errors.NoZoneError
	}

	q := url.Values{}
	if recordType != "" {
		q.Set("fieldType", recordType)
	}

	if subDomain != "" {
		q.Set("subDomain", subDomain)
	}

	err := c.client.Get(fmt.Sprintf("/domain/zone/%s/record?%s", zone, q.Encode()), &records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetRecord read DNS record details.
func (c Client) GetRecord(zone string, id int64) (*DNSRecord, error) {
	var record DNSRecord

	err := c.client.Get(fmt.Sprintf("/domain/zone/%s/record/%d", zone, id), &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}
