package ip

import "testing"

func TestGetPublic(t *testing.T) {
	ip, err := GetPublic()
	if err != nil {
		t.Error(err)
	}

	t.Logf("found public ip=%s", ip)
}
