package server

import (
	"github.com/jeremybouzigard/library"
)

// GenreResponse represents the primary data provided in the response to a
// successful request to fetch a genre resource object.
type GenreResponse struct {
	Data []*library.Genre `json:"data,omitempty"`
}
