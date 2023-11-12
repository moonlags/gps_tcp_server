package tcp

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

func NewTcpServer(port int) (net.Listener, error) {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	log.WithField("listener", listener).Info("initializated new tcp listener")

	return listener, nil
}
