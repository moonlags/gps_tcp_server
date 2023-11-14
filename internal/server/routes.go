package server

import (
	"encoding/json"
	"fmt"
	"gps_tcp_server/internal/globals"
	"net/http"
	"strconv"

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

	r.Route("/api", func(r chi.Router) {
		r.Get("/user/{ID}/device/{IMEI}", s.DeviceData)
		r.Post("/user/login", s.UserData)
		r.Post("/user/link", s.LinkUser)
	})

	return r
}

func (s *Server) LinkUser(w http.ResponseWriter, r *http.Request) {
	requestBody := new(LinkData)

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	globals.Mutex().Lock()

	if _, ok := globals.DeviceSessions()[requestBody.IMEI]; !ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	globals.Mutex().Unlock()
	globals.UserMutex().Lock()

	user := globals.Users()[requestBody.ID]
	user.IMEI = requestBody.IMEI
	globals.Users()[requestBody.ID] = user

	globals.UserMutex().Unlock()

	if _, err := w.Write([]byte("Linked device")); err != nil {
		log.WithError(err).Fatal("failed to write json response")
	}
}

func (s *Server) UserData(w http.ResponseWriter, r *http.Request) {
	requestBody := new(UserLogin)

	err := json.NewDecoder(r.Body).Decode(requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	globals.UserMutex().Lock()

	if _, ok := globals.Users()[requestBody.ID]; !ok {
		globals.Users()[requestBody.ID] = globals.User{GithubID: requestBody.ID}
	}

	globals.UserMutex().Unlock()

	jsonResp, err := json.Marshal(globals.Users()[requestBody.ID])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := w.Write(jsonResp); err != nil {
		log.WithError(err).WithField("jsonResp", jsonResp).Fatal("failed to write json response")
	}
}

func (s *Server) DeviceData(w http.ResponseWriter, r *http.Request) {
	IMEI := chi.URLParam(r, "IMEI")
	IDstr := chi.URLParam(r, "ID")

	fmt.Println(IDstr)

	ID, err := strconv.ParseInt(IDstr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	globals.UserMutex().Lock()
	if globals.Users()[ID].IMEI != IMEI {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	globals.UserMutex().Unlock()

	globals.Mutex().Lock()
	deviceData, ok := globals.DeviceSessions()[IMEI]
	if !ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	globals.Mutex().Unlock()

	jsonResp, err := json.Marshal(deviceData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := w.Write(jsonResp); err != nil {
		log.WithError(err).WithField("jsonResp", jsonResp).Fatal("failed to write json response")
	}
}
