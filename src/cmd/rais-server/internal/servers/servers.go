package servers

import (
	"context"
	"net/http"
	"sync"
	"time"
)

var servers = make(map[string]*Server)
var running sync.WaitGroup

// Server wraps an http.Server with some helpers for running in the background,
// setting up sane defaults (no global ServeMux), and shutdown of all
// registered servers
type Server struct {
	*http.Server
	Name string
	Mux  *http.ServeMux
	middleware []func(http.Handler) http.Handler
}

// NewServer registers a named server at the given bind address.  If the
// address is already in use, the "new" server will instead merge with the
// existing server.
func New(name, addr string) *Server {
	if servers[addr] != nil {
		servers[addr].Name += ", " + name
		return servers[addr]
	}

	var mux = http.NewServeMux()
	var s = &Server{
		Name: name,
		Mux:  mux,
		Server: &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 30 * time.Second,
			Addr:         addr,
			Handler:      mux,
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

// Handle wraps the server's ServeMux Handle method
func (s *Server) Handle(pattern string, handler http.Handler) {
	for _, m := range s.middleware {
		handler = m(handler)
	}
	s.Mux.Handle(pattern, handler)
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
func Shutdown(ctx context.Context) {
	for _, s := range servers {
		s.Shutdown(ctx)
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
