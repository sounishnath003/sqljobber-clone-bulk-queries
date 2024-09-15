package main

import (
	"fmt"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sounishnath003/jobprocessor/internal/core"
	"github.com/sounishnath003/jobprocessor/internal/handlers"
)

type Server struct {
	PORT int
}

func NewServer(co *core.Core) *Server {
	return &Server{
		PORT: co.Conf.PORT,
	}
}

// MustStart - Initiates the server with panic. sets up with the provided Conf parameters from the arguments.
// Implements the handlers as well and Middlewares for the echo.Echo server.
func (s *Server) MustStart() {
	e := ConfigureMiddlewares(echo.New())

	apiRoutes := e.Group("/api")
	apiRoutes.Add("GET", "/healthy", handlers.HealthyHandler)

	// Versioning V1 routes
	v1Routes := e.Group("api/v1")
	v1Routes.Add("GET", "/tasks", handlers.GetJobs)

	// Go routine
	go func() {
		log.Printf("Api is up and running on http://localhost:%d\n", s.PORT)
	}()

	// Starting the server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", s.PORT)))
}

// Implement the following middle for Echo server
//
// 1) Logging every request and response with Timing
//
// 2) Throws error if the Response of the output is More than 2 Seconds
//
// 3) Sends the Application/JSON as Content-Type Header
//
// 4) Adds TraceID in the Response Data for the Future Tracking
func ConfigureMiddlewares(e *echo.Echo) *echo.Echo {

	// Inbuilt - Middlewares groupping addition
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:4200", "http://localhost:3000"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderContentSecurityPolicy, echo.HeaderAuthorization, echo.HeaderXForwardedFor},
		AllowMethods: []string{"GET", "POST", "DELETE"},
	}))

	e.Use(middleware.AddTrailingSlash())
	e.Use(middleware.Decompress())
	e.Use(middleware.Logger())
	e.Use(middleware.RequestID())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "Response is taking too long to execute. Please try again.",
		OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
			log.Println(c.Path())
		},
		Timeout: 5 * time.Second,
	}))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))

	// Custom Middlewares
	e.Use(TimingMiddleware)

	return e
}

func TimingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		latency := time.Since(start).Seconds()
		c.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Response().Header().Set("X-Response-Time", fmt.Sprintf("%f seconds", latency))
		return err
	}
}
