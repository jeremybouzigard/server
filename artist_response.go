package server

import (
	"github.com/jeremybouzigard/library"
)

// ArtistResponse represents the primary data provided in the response to a
// successful request to fetch an artist resource object.
type ArtistResponse struct {
	Data []*library.Artist `json:"data,omitempty"`
}
