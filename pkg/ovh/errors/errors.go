package errors

import (
	"errors"
)

var (
	// ErrNoZone is returned when the zone name argument is missing.
	ErrNoZone = errors.New("missing zone name argument")
	// ErrTooManyRecords is returned when more than one DNS record is found.
	ErrTooManyRecords = errors.New("too many dns records found")
	// ErrRecordNotFound is returned when no DNS record is found.
	ErrRecordNotFound = errors.New("dns record not found")
)
