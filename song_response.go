package server

import (
	"github.com/jeremybouzigard/library"
)

// SongResponse represents the primary data provided in the response to a
// successful request to fetch a song resource object.
type SongResponse struct {
	Data []*library.Song `json:"data,omitempty"`
}
