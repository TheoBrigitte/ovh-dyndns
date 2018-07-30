package client

// DNSRecord represent OVH API dns record object.
type DNSRecord struct {
	ID        int64  `json:"id,omitempty"`
	FieldType string `json:"fieldType,omitempty"`
	SubDomain string `json:"subDomain,omitempty"`
	Target    string `json:"target,omitempty"`
	TTL       int64  `json:"ttl,omitempty"`
	Zone      string `json:"zone,omitempty"`
}
