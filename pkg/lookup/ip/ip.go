package ip

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

const ipifyURL = "https://api.ipify.org?format=json"

type IPify struct {
	IP string `json:"ip"`
}

const errInvalidIP = "invalid IP address format"

// Get retrieve public facing ip adresse.
func GetPublic() (string, error) {
	var ipify IPify

	res, err := http.Get(ipifyURL)
	if err != nil {
		return "", err
	}

	decoder := json.NewDecoder(res.Body)

	err = decoder.Decode(&ipify)
	if err != nil {
		return "", err
	}

	ip := net.ParseIP(ipify.IP)
	if ip == nil {
		return "", fmt.Errorf("%v : %v", errInvalidIP, ipify.IP)
	}

	return ipify.IP, nil
}
