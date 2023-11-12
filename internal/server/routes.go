package server

import (
	"encoding/json"
	"gps_tcp_server/internal/globals"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	log "github.com/sirupsen/logrus"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"User-Agent", "Authorization", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "Cache-Control", "Connection", "DNT", "Host", "Origin", "Pragma", "Referer"},
		AllowCredentials: true,
	}))

	r.Get("/", s.helloWorldHandler)
	r.Route("/api", func(r chi.Router) {
		r.Get("/device/{IMEI}", s.DeviceData)
	})

	return r
}

func (s *Server) helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.WithError(err).Fatal("failed to marshal json")
	}

	if _, err = w.Write(jsonResp); err != nil {
		log.WithError(err).Fatal("failed to write json response")
	}
}

func (s *Server) DeviceData(w http.ResponseWriter, r *http.Request) {
	IMEI := chi.URLParam(r, "IMEI")
	mutex := globals.Mutex()

	mutex.RLock()
	device_data, ok := globals.DeviceSessions()[IMEI]
	if !ok {
		log.WithField("IMEI", IMEI).Warn("invalid IMEI")
		return
	}
	mutex.RUnlock()

	jsonResp, err := json.Marshal(device_data)
	if err != nil {
		log.WithError(err).WithField("device_data", device_data).Fatal("failed to marshal json")
	}

	if _, err := w.Write(jsonResp); err != nil {
		log.WithError(err).WithField("jsonResp", jsonResp).Fatal("failed to write json response")
	}
}
