package ip

import (
	"net"
	"testing"
)

func TestGetPublic(t *testing.T) {
	ip, err := GetPublic()
	if err != nil {
		t.Error(err)
	}

	valid := net.ParseIP(ip)
	if valid == nil {
		t.Errorf("%v : %v", errInvalidIP, ip)
	}

	t.Logf("found public ip=%s", ip)
}
