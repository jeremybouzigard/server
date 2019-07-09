package http

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	"github.com/jeremybouzigard/library"
	"github.com/jeremybouzigard/server"
)

// Handler contains an HTTP router, a collection of all services to handle HTTP
// requests, and a logger to log errors.
type Handler struct {
	Router *mux.Router
	Logger *log.Logger

	GenreService  library.GenreService
	AlbumService  library.AlbumService
	ArtistService library.ArtistService
	SongService   library.SongService
}

// NewHandler returns a new instance of a Handler.
func NewHandler() *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
		Logger: log.New(os.Stderr, "", log.LstdFlags)}
	return h
}

// ServeHTTP routes HTTP requests to the appropriate handler function.
func (h *Handler) ServeHTTP() {
	h.Router.HandleFunc("/albums", h.handleGetAlbums).Methods("GET")
	h.Router.HandleFunc("/albums/{id:[0-9]+}", h.handleGetAlbumByID).Methods("GET")
	h.Router.HandleFunc("/genres", h.handleGetGenres).Methods("GET")
	h.Router.HandleFunc("/artists", h.handleGetArtists).Methods("GET")
	h.Router.HandleFunc("/artists/{id:[0-9]+}", h.handleGetArtistByID).Methods("GET")
	h.Router.HandleFunc("/songs", h.handleGetSongs).Methods("GET")
	h.Router.HandleFunc("/songs/{id:[0-9]+}", h.handleGetSongByID).Methods("GET")
	h.Router.PathPrefix("/").HandlerFunc(handleNotFound)
	http.ListenAndServe(":8080", h.Router)
}

// handleGetSongByID handles a request to get a song with the given ID.
func (h *Handler) handleGetSongByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) > 0 {
		a, err := h.SongService.Song(id)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
		} else if a == nil {
			handleNotFound(w, r)
		} else {
			var songs []*library.Song
			songs = append(songs, a)
			response := server.SongResponse{Data: songs}
			encodeJSON(w, response)
		}
	}
}

// handleGetSongs handles a request to get albums.
func (h *Handler) handleGetSongs(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	queries := parseQueries(v)
	songs, err := h.SongService.Songs(queries)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
	} else if songs == nil {
		handleNotFound(w, r)
	} else {
		response := server.SongResponse{Data: songs}
		encodeJSON(w, response)
	}
}

// handleGetAlbums handles a request to get an album with the given ID.
func (h *Handler) handleGetArtistByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) > 0 {
		a, err := h.ArtistService.Artist(id)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
		} else if a == nil {
			handleNotFound(w, r)
		} else {
			var artists []*library.Artist
			artists = append(artists, a)
			response := server.ArtistResponse{Data: artists}
			encodeJSON(w, response)
		}
	}
}

// handleGetArtists handles a request to get albums.
func (h *Handler) handleGetArtists(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	queries := parseQueries(v)
	artists, err := h.ArtistService.Artists(queries)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
	} else if artists == nil {
		handleNotFound(w, r)
	} else {
		response := server.ArtistResponse{Data: artists}
		encodeJSON(w, response)
	}
}

// handleGetGenres handles a request to get all genres.
func (h *Handler) handleGetGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := h.GenreService.Genres()
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
	} else {
		response := server.GenreResponse{Data: genres}
		encodeJSON(w, response)
	}
}

// handleGetAlbums handles a request to get an album with the given ID.
func (h *Handler) handleGetAlbumByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) > 0 {
		a, err := h.AlbumService.Album(id)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
		} else {
			var albums []*library.Album
			albums = append(albums, a)
			response := server.AlbumResponse{Data: albums}
			encodeJSON(w, response)
		}
	}
}

// handleGetAlbums handles a request to get albums.
func (h *Handler) handleGetAlbums(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	queries := parseQueries(v)
	albums, err := h.AlbumService.Albums(queries)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
	} else {
		response := server.AlbumResponse{Data: albums}
		encodeJSON(w, response)
	}
}

// parseQueries parses URL values for known possible queries.
func parseQueries(v url.Values) map[string]string {
	queries := make(map[string]string, 3)
	queries["albumID"] = v.Get("album-id")
	queries["artistID"] = v.Get("artist-id")
	queries["genreID"] = v.Get("genre-id")
	return queries
}

// encodeJSON writes the JSON-encoded response.
func encodeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
	}
}

// handleNotFound writes the API error message when a fetched resource object
// is not found.
func handleNotFound(w http.ResponseWriter, r *http.Request) {
	handleError(w, nil, http.StatusNotFound)
}

// handleError writes an API error message to the response.
func handleError(w http.ResponseWriter, err error, code int) {
	var er server.ErrorResponse
	var e *server.Error

	if code == http.StatusInternalServerError {
		e = server.NewInternalServerError()
	} else if code == http.StatusNotFound {
		e = server.NewStatusNotFoundError()
	} else {
		e = &server.Error{Status: string(code),
			Detail: err.Error()}
	}
	er.Errors = append(er.Errors, *e)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(er)
}
