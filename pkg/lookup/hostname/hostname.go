package hostname

import (
	"net"
)

// Get issue a reverse DNS lookup using net.LookupAddr and return hostnames.
func Get(addr string) ([]string, error) {
	names, err := net.LookupAddr(addr)
	if err != nil {
		return nil, err
	}

	if len(names) < 1 {
		return nil, NoHostnameFound
	}

	return names, nil
}
