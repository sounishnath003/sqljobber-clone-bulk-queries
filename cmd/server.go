package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sounishnath003/jobprocessor/internal/core"
	"github.com/sounishnath003/jobprocessor/internal/core/handlers"
)

// Server - struct for the initlization of the http.Server.
type Server struct {
	Core *core.Core
}

// NewServer - returns the instantianited server construct.
func NewServer(co *core.Core) *Server {
	return &Server{
		Core: co,
	}
}

// MustStart - Initiates the server with panic. sets up with the provided Conf parameters from the arguments.
// Implements the handlers as well and Middlewares for the *http.ServeMux Mux server.
func (s *Server) MustStart() {
	// Create the API version and path prefix and group handler with /api/v1
	srv := http.NewServeMux()

	// Define route endpoints
	srv.HandleFunc("GET /healthy", handlers.HealthyHandler)
	srv.HandleFunc("GET /api/v1/tasks", handlers.HandleGetTasksList)
	srv.HandleFunc("GET /api/v1/jobs/{jobID}", handlers.HandleGetJobStatus)

	// Go routine
	// So that's its not blocking
	// Starting the server
	go func() {
		log.Printf("API is up and running on http://localhost:%d\n", s.Core.Conf.PORT)
		http.ListenAndServe(fmt.Sprintf(":%d", s.Core.Conf.PORT), ConfigureMiddlewares(srv, s.Core))
	}()

}

func ConfigureMiddlewares(srv *http.ServeMux, co *core.Core) http.Handler {
	return TraceIdMiddleware(NecessaryJobProcessorMiddleware(srv, co))
}

func NecessaryJobProcessorMiddleware(next http.Handler, co *core.Core) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Println("[SERVER_LOG]: server received request",
			"method", r.Method,
			"header", r.Header.Clone(),
			"uri", r.URL.Path,
			"remote-address", r.RemoteAddr,
			"content-length", r.ContentLength,
			"form", r.Form,
			"time-taken", time.Since(start),
		)
		// Setting up the *core.Core attrb into the context.Background().
		// So use it whenever needed in my API functionalities.
		ctx := context.WithValue(r.Context(), "core", co)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TraceIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceId := uuid.New().String()
		w.Header().Add("X-Trace-Id", traceId)
		next.ServeHTTP(w, r)
	})
}
