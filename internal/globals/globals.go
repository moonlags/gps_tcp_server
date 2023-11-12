package globals

import (
	"gps_tcp_server/internal/tcp"
	"sync"
)

var (
	deviceSessions = make(map[string]*tcp.DeviceData)
	mutex          = new(sync.RWMutex)
)

func DeviceSessions() map[string]*tcp.DeviceData {
	return deviceSessions
}

func Mutex() *sync.RWMutex {
	return mutex
}
