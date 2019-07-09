package server

import (
	"github.com/jeremybouzigard/library"
)

// AlbumResponse represents the primary data provided in the response to a
// successful request to fetch an album resource object.
type AlbumResponse struct {
	Data []*library.Album `json:"data,omitempty"`
}
