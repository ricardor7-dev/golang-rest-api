package server

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	//"github.com/rs/zerolog/pkgerrors"

	"github.com/go-openapi/runtime/middleware"

	"golang-rest-api/repositories"
	"golang-rest-api/server/driver"
	"golang-rest-api/errs"
)

const pathPrefix string = "/audios"

//Repositories
type Repositories struct{
	AudiobookRepository			repositories.AudiobookRepository
	LanguageRepository			repositories.LanguagesRepository
	TagRepository				repositories.TagRepository
}

type Server struct {
	router *mux.Router
	Driver driver.Server
	//DB *db.BDData
	Log *zerolog.Logger
	Addr string
	Repositories 
}


// NewMuxRouter initializes a gorilla/mux router and
// adds the /api subroute to it
func NewMuxRouter() *mux.Router {
	// initializer gorilla/mux router
	r := mux.NewRouter()

	//TODO: Put here or in another file?
	r.Handle("/swagger.yaml", http.FileServer(http.Dir("../swagger/api")))
	opts := middleware.SwaggerUIOpts{SpecURL: "/swagger.yaml"}
	sh := middleware.SwaggerUI(opts, nil)
	r.Handle("/docs", sh)

	opts1 := middleware.RedocOpts{SpecURL: "/swagger.yaml", Path: "docs1"}
	sh1 := middleware.Redoc(opts1, nil)
	r.Handle("/docs1", sh1)

	// send Router through PathPrefix method to validate any standard
	// subroutes you may want for your APIs. e.g. I always want to be
	// sure that every request has "/api" as part of its path prefix
	// without having to put it into every handle path in my various
	// routing functions
	s := r.PathPrefix(pathPrefix).Subrouter()

	return s
}

// New initializes a new Server and registers
// routes to the given router
//func New(rtr *mux.Router, serverDriver driver.Server, lgr zerolog.Logger) *Server {
func New(rtr *mux.Router, serverDriver driver.Server, lgr zerolog.Logger) *Server {

	s := &Server{router: rtr}
	s.Log = &lgr
	s.Driver = serverDriver

	// register routes to the router
	s.registerRoutes()

	return s
}

// ListenAndServe is a wrapper to use wherever http.ListenAndServe is used.
func (s *Server) ListenAndServe() error{
	const op errs.Op = "server/Server.ListenAndServe"
	if s.Addr == "" {
		return errs.E(op, "Server Addr is empty")
	}
	if s.router == nil {
		return errs.E(op, "Server router is nil")
	}
	if s.Driver == nil {
		return errs.E(op, "Server driver is nil")
	}
	return s.Driver.ListenAndServe(s.Addr, s.router)
}


// Driver implements the driver.Server interface. The zero value is a valid http.Server.
type Driver struct {
	Server http.Server
}
// NewDriver creates a Driver enfolding a http.Server with default timeouts.
func NewDriver() *Driver {
	return &Driver{
		Server: http.Server{
			ReadTimeout:  40 * time.Second,
			WriteTimeout: 40 * time.Second,
			IdleTimeout:  140 * time.Second,
		},
	}
}

// ListenAndServe sets the address and handler on Driver's http.Server,
// then calls ListenAndServe on it.
func (d *Driver) ListenAndServe(addr string, h http.Handler) error {
	d.Server.Addr = addr
	d.Server.Handler = h
	return d.Server.ListenAndServe()
}

// Shutdown gracefully shuts down the server without interrupting any active connections,
// by calling Shutdown on Driver's http.Server
func (d *Driver) Shutdown(ctx context.Context) error {
	return d.Server.Shutdown(ctx)
}

// decoderErr is a convenience function to handle errors returned by
// json.NewDecoder(r.Body).Decode(&data) and return the appropriate
// error response
func decoderErr(err error) error {
	const op errs.Op = "server/decoderErr"

	switch {
	// If the request body is empty (io.EOF)
	// return an error
	case err == io.EOF:
		return errs.E(op, errs.ERR_DECODEJSON, "request body cannot be empty")
	// If the request body has malformed JSON (io.ErrUnexpectedEOF)
	// return an error
	case err == io.ErrUnexpectedEOF:
		return errs.E(op, errs.ERR_DECODEJSON, "malformed JSON")
	// return other errors
	case err != nil:
		return errs.E(op, errs.ERR_DECODEJSON, err)
	}
	return nil
}
