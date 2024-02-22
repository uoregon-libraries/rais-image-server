package servers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/uoregon-libraries/gopkg/logger"
)

var servers = make(map[string]*Server)
var running sync.WaitGroup

// Server wraps an http.Server with some helpers for running in the background,
// setting up sane defaults (no global ServeMux), and shutdown of all
// registered servers
type Server struct {
	*http.Server
	Name       string
	Mux        *mux.Router
	middleware []func(http.Handler) http.Handler
}

// New registers a named server at the given bind address.  If the address is
// already in use, the "new" server will instead merge with the existing
// server.
func New(name, addr string) *Server {
	if servers[addr] != nil {
		servers[addr].Name += ", " + name
		return servers[addr]
	}

	var m = mux.NewRouter()
	m.SkipClean(true)
	var s = &Server{
		Name: name,
		Mux:  m,
		Server: &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 30 * time.Second,
			Addr:         addr,
			Handler:      m,
		},
	}

	servers[addr] = s
	return s
}

// AddMiddleware appends to the list of middleware handlers - these wrap *all*
// handlers in the given middleware
//
// Middleware is any function that takes a handler and returns another handler
func (s *Server) AddMiddleware(mw func(http.Handler) http.Handler) {
	s.middleware = append(s.middleware, mw)
}

func (s *Server) wrapMiddleware(handler http.Handler) http.Handler {
	for _, m := range s.middleware {
		handler = m(handler)
	}

	return handler
}

// HandleExact sets up a gorilla/mux handler that response only to the exact
// path given
func (s *Server) HandleExact(pth string, handler http.Handler) {
	s.Mux.Path(pth).Handler(s.wrapMiddleware(handler))
}

// HandlePrefix sets up a gorilla/mux handler for any request where the
// beginning of the path matches the given prefix
func (s *Server) HandlePrefix(prefix string, handler http.Handler) {
	s.Mux.PathPrefix(prefix).Handler(s.wrapMiddleware(handler))
}

// run wraps http.Server's ListenAndServe in a background-friendly way, sending
// any errors to the "done" callback when the server closes
func (s *Server) run(done func(*Server, error)) {
	var err = s.Server.ListenAndServe()
	if err == http.ErrServerClosed {
		err = nil
	}
	done(s, err)
}

// Shutdown stops all registered servers
func Shutdown(ctx context.Context, l *logger.Logger) {
	for _, s := range servers {
		var err = s.Shutdown(ctx)
		if err != nil {
			l.Errorf("Error shutting down server %q: %s", s.Name, err)
		}
	}
}

// ListenAndServe runs all servers and waits for them to shut down, running onErr
// when a server returns an error (other than http.ErrServerClosed) occurs
func ListenAndServe(onErr func(*Server, error)) {
	var done = func(s *Server, err error) {
		running.Done()
		if err != nil {
			onErr(s, err)
		}
	}

	for _, s := range servers {
		running.Add(1)
		go s.run(done)
	}

	running.Wait()
}
