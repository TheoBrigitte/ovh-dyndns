package ip

import (
	externalip "github.com/glendc/go-external-ip"
)

const errInvalidIP = "invalid IP address format"

// GetPublic retrieves the public facing IP address
func GetPublic() (string, error) {
	consensus := externalip.DefaultConsensus(nil, nil)

	ip, err := consensus.ExternalIP()
	if err != nil {
		return "", err
	}

	return ip.String(), nil
}

func GetPublicIPv4() (string, error) {
	consensus := externalip.DefaultConsensus(nil, nil)
	err := consensus.UseIPProtocol(4)
	if err != nil {
		return "", err
	}

	ip, err := consensus.ExternalIP()
	if err != nil {
		return "", err
	}

	return ip.String(), nil
}

func GetPublicIPv6() (string, error) {
	consensus := externalip.DefaultConsensus(nil, nil)
	err := consensus.UseIPProtocol(6)
	if err != nil {
		return "", err
	}

	ip, err := consensus.ExternalIP()
	if err != nil {
		return "", err
	}

	return ip.String(), nil
}
