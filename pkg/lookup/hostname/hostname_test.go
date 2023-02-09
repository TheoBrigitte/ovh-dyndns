package hostname

import (
	"testing"
)

const testIP = "1.1.1.1"

func TestGet(t *testing.T) {
	hostnames, err := Get(testIP)
	if err != nil {
		t.Error(err)
	}

	if len(hostnames) <= 0 {
		t.Errorf("found 0 hostnames for ip=%s", testIP)
	}

	t.Logf("found %d hostnames for ip=%s\n%v\n", len(hostnames), testIP, hostnames)
}
