package ip

import (
	"encoding/json"
	"net/http"
)

const ipifyURL = "https://api.ipify.org?format=json"

type IPify struct {
	IP string `json:"ip"`
}

// Get retrieve public facing ip adresse.
func GetPublic() (string, error) {
	var ip IPify

	res, err := http.Get(ipifyURL)
	if err != nil {
		return "", err
	}

	decoder := json.NewDecoder(res.Body)

	err = decoder.Decode(&ip)
	if err != nil {
		return "", err
	}

	return ip.IP, nil
}
