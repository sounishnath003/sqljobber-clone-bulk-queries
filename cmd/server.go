package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sounishnath003/jobprocessor/internal/core"
	"github.com/sounishnath003/jobprocessor/internal/core/handlers"
)

type Server struct {
	Core *core.Core
}

func NewServer(co *core.Core) *Server {
	return &Server{
		Core: co,
	}
}

// MustStart - Initiates the server with panic. sets up with the provided Conf parameters from the arguments.
// Implements the handlers as well and Middlewares for the echo.Echo server.
func (s *Server) MustStart() {
	// Create the API version and path prefix and group handler with /api/v1
	srv := http.NewServeMux()

	// Define route endpoints
	srv.HandleFunc("GET /healthy", handlers.HealthyHandler)
	srv.HandleFunc("GET /api/v1/tasks", handlers.HandleGetTasksList)

	// Go routine
	go func() {
		log.Printf("API is up and running on http://localhost:%d\n", s.Core.Conf.PORT)
	}()

	// Starting the server
	http.ListenAndServe(fmt.Sprintf(":%d", s.Core.Conf.PORT), ConfigureMiddlewares(srv))
}

func ConfigureMiddlewares(srv *http.ServeMux) http.Handler {
	return LoggerMiddleware(TraceIdMiddleware(srv))
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("Method=%s Path=%s RemoteIP=%s, Time=%s", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}
func TraceIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceId := uuid.New().String()
		w.Header().Add("X-Trace-Id", traceId)
		next.ServeHTTP(w, r)
	})
}

func TimeTakenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("Time taken: %s", time.Since(start))
	})
}
