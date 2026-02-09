package events

import "errors"

var (
	ErrMissingEventID = errors.New("event_id is required")
	ErrMissingSource  = errors.New("source is required")
	ErrMissingTitle   = errors.New("title is required")
	ErrMissingURL      = errors.New("url is required")
)
