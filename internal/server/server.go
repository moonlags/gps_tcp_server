package server

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	port int
}

func NewServer(port int) *http.Server {

	NewServer := &Server{
		port: port,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.WithField("server", server).Info("initializated new server")

	return server
}
