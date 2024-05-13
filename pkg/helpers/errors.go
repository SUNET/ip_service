package helpers

import "errors"

var (
	// ErrNotValidEndpoint is returned when the endpoint is not valid
	ErrNotValidEndpoint = errors.New("not a valid endpoint")

	// ErrMissingDBFile is returned when the DB file is missing
	ErrMissingDBFile = errors.New("missing DB file")
)
