package hostname

import (
	"errors"
)

// ErrHostnameNotFound is returned when no hostname is found for the given IP address.
var ErrHostnameNotFound = errors.New("no hostname found")
