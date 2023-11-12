package main

import (
	"gps_tcp_server/internal/globals"
	"gps_tcp_server/internal/server"
	"gps_tcp_server/internal/tcp"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
}

func main() {
	go tcp.TcpHandler(globals.DeviceSessions(), globals.Mutex())

	server := server.NewServer(50731)

	if err := server.ListenAndServe(); err != nil {
		log.WithError(err).Fatal("failed to start the server")
	}
}
