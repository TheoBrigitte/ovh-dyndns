package errors

import (
	"errors"
)

var (
	NoZoneError    = errors.New("missing zone name argument")
	TooManyRecords = errors.New("too many dns records found")
	RecordNotFound = errors.New("dns record not found")
)
