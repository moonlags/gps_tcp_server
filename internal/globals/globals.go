package globals

import (
	"gps_tcp_server/internal/tcp"
	"sync"
)

type User struct {
	GithubID int64
	IMEI     string
}

var (
	deviceSessions = make(map[string]*tcp.DeviceData)
	mutex          = new(sync.Mutex)

	users     = make(map[int64]User)
	userMutex = new(sync.Mutex)
)

func DeviceSessions() map[string]*tcp.DeviceData {
	return deviceSessions
}

func Mutex() *sync.Mutex {
	return mutex
}

func Users() map[int64]User {
	return users
}

func UserMutex() *sync.Mutex {
	return userMutex
}
