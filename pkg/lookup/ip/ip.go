package ip

import (
	externalip "github.com/glendc/go-external-ip"
)

const errInvalidIP = "invalid IP address format"

// Get retrieve public facing ip adresse.
func GetPublic() (string, error) {
	consensus := externalip.DefaultConsensus(nil, nil)
	consensus.UseIPProtocol(4)
	ip, err := consensus.ExternalIP()
	if err != nil {
		return "", err
	}

	return ip.String(), nil
}
