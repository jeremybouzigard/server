package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
	"github.com/jeremybouzigard/library"
	"github.com/jeremybouzigard/server"
	"github.com/jeremybouzigard/server/pkg/hls"
)

// Handler contains an HTTP router, a collection of all services to handle HTTP
// requests, and a logger to log errors.
type Handler struct {
	Router  *mux.Router
	Logger  *log.Logger
	TempDir string

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

// setTempDir creates a temporary directory to store the files generated for
// HTTP Live Streaming, including index files (playlists) and media stream
// segments.
func (h *Handler) setTempDir() error {
	dir, err := ioutil.TempDir("", "hls")
	if err != nil {
		h.Logger.Fatal(err)
		return err
	}
	h.TempDir = dir
	return nil
}

// StartServer performs an initial setup and then starts the media server.
func (h *Handler) StartServer() {
	// Creates temporary directory for HLS files.
	err := h.setTempDir()
	if err != nil {
		return
	}

	// Routes HTTP requests to the appropriate handler function.
	h.Router.HandleFunc("/albums", h.handleGetAlbums).Methods("GET")
	h.Router.HandleFunc("/albums/{id:[0-9]+}", h.handleGetAlbumByID).Methods("GET")
	h.Router.HandleFunc("/genres", h.handleGetGenres).Methods("GET")
	h.Router.HandleFunc("/artists", h.handleGetArtists).Methods("GET")
	h.Router.HandleFunc("/artists/{id:[0-9]+}", h.handleGetArtistByID).Methods("GET")
	h.Router.HandleFunc("/songs", h.handleGetSongs).Methods("GET")
	h.Router.HandleFunc("/songs/{id:[0-9]+}", h.handleGetSongByID).Methods("GET")
	h.Router.HandleFunc("/songs/{id:[0-9]+}/stream", h.handleGetStreamPlaylist).Methods("GET")
	h.Router.HandleFunc("/songs/{id:[0-9]+}/{seg:fileSequence[0-9]+.aac}", h.handleGetStreamSegment).Methods("GET")
	h.Router.PathPrefix("/").HandlerFunc(handleNotFound)

	// Creates server.
	srv := &http.Server{Addr: ":8080", Handler: h.Router}

	// Defines shutdown behavior.
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// Shuts down when an interrupt signal is received.
		if err := srv.Shutdown(context.Background()); err != nil {
			h.Logger.Printf("HTTP server Shutdown: %v", err)
		}

		// On shutdown, removes temporary directory and closes idle connections.
		h.Logger.Printf("HTTP server Shutdown")
		os.RemoveAll(h.TempDir)
		close(idleConnsClosed)
	}()

	// Begins listening for and serving requests.
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		h.Logger.Printf("HTTP server ListenAndServe: %v", err)
	}
	<-idleConnsClosed
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

// handleGetSongs handles a request to get song data.
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

// handleGetArtists handles a request to get artist data.
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

// handleGetGenres handles a request to get all genre data.
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

// handleGetStreamPlaylist handles a request to get the stream index file for
// the given song ID. An index file, or playlist, provides an ordered list of
// paths of the media segment files.
func (h *Handler) handleGetStreamPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songID := vars["id"]
	if len(songID) > 0 {
		song, err := h.SongService.Song(songID)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
		} else if song == nil {
			handleNotFound(w, r)
		} else {
			h.servePlaylist(w, r, songID, song.Attributes.FilePath)
		}
	}
}

// servePlaylist serves the stream index (playlist) file for the given song ID.
func (h *Handler) servePlaylist(w http.ResponseWriter, r *http.Request,
	songID string, songPath string) {
	playlistPath := fmt.Sprintf("%s/%s/prog_index.m3u8", h.TempDir, songID)
	if _, err := os.Stat(playlistPath); os.IsNotExist(err) {
		playlistDir := fmt.Sprintf("%s/%s", h.TempDir, songID)
		os.Mkdir(playlistDir, 0700)
		hls.Segment(songPath, playlistDir)
	}
	w.Header().Set("Content-Type", "application/x-mpegURL")
	http.ServeFile(w, r, playlistPath)
}

// handleGetStreamSegment handles a request to get a media segment file.
func (h *Handler) handleGetStreamSegment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songID := vars["id"]
	if len(songID) > 0 {
		song, err := h.SongService.Song(songID)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
		} else if song == nil {
			handleNotFound(w, r)
		} else {
			seg := vars["seg"]
			h.serveSegment(w, r, seg, songID)
		}
	}
}

// serveSegment serves a media segment file.
func (h *Handler) serveSegment(w http.ResponseWriter, r *http.Request,
	seg string, songID string) {
	playlistDir := fmt.Sprintf("%s/%s", h.TempDir, songID)
	segPath := fmt.Sprintf("%s/%s", playlistDir, seg)
	w.Header().Set("Content-Type", "audio/aac")
	http.ServeFile(w, r, segPath)
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
