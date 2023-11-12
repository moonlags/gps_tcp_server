package tcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func TcpHandler(device_sessions map[string]*DeviceData, mutex *sync.RWMutex) {
	listener, err := NewTcpServer(55080)
	if err != nil {
		log.WithError(err).Fatal("failed to create new tcp listener")
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.WithError(err).WithField("listener", listener).Error("failed to accept a tcp connection")
			continue
		}

		go handleTcpConnection(conn, device_sessions, mutex)
	}
}

func handleTcpConnection(conn net.Conn, device_sessions map[string]*DeviceData, mutex *sync.RWMutex) {
	defer conn.Close()

	device_data := DeviceData{
		IMEI:                 "",
		Postitions:           []PostitioningPacket{},
		BatteryPower:         100,
		LastStatusPacketTime: time.Now().Unix(),
		StatusCooldown:       1,
		IsLoggedIn:           false,
		Connection:           conn,
	}

	log.WithField("connection", conn).Info("accepted new connection")

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.WithError(err).WithField("connection", conn).Error("failed to read from connection")
			break
		}

		tempLoggedIn := device_data.IsLoggedIn

		response, err := parsePacket(buffer[:n], conn.RemoteAddr(), &device_data)
		if err != nil {
			log.WithError(err).Error("failed to parse a packet")
			break
		}

		if tempLoggedIn != device_data.IsLoggedIn {
			mutex.Lock()
			device_sessions[device_data.IMEI] = &device_data
			mutex.Unlock()
		}

		if _, err := conn.Write(response); err != nil {
			log.WithError(err).WithField("connection", conn).Error("failed to write to the connection")
			break
		}
	}
	device_data.IsLoggedIn = false
}

func parsePacket(packet []byte, address net.Addr, device_data *DeviceData) ([]byte, error) {
	log.WithField("packet", packet).Info("have packet")

	if !bytes.HasSuffix(packet, []byte{0x0d, 0x0a}) || !bytes.HasPrefix(packet, []byte{0x78, 0x78}) {
		return nil, errors.New("invalid packet format")
	}

	packet_length := int(packet[2])
	protocol_number := packet[3]

	if packet_length+4 >= len(packet) {
		return nil, errors.New("invalid packet length")
	}

	switch protocol_number {
	default:
		return nil, errors.New("not supported protocol number")
	case 1:
		if packet_length != 10 {
			return nil, errors.New("invalid packet length")
		}

		packet_struct := LoginPacket{ImeiToString(packet[4:12])}

		return packet_struct.Process(device_data), nil
	case 8:
		if packet_length != 1 {
			return nil, errors.New("invalid packet length")
		}

		return []byte{}, nil
	case 0x10, 0x11:
		if packet_length != 0x13 {
			return nil, errors.New("invalid packet length")
		} else if !device_data.IsLoggedIn {
			return nil, errors.New("device is not logged in")
		}

		var latitude, longitude uint32
		if err := binary.Read(bytes.NewReader(packet[11:15]), binary.BigEndian, &latitude); err != nil {
			return nil, err
		} else if err := binary.Read(bytes.NewReader(packet[15:19]), binary.BigEndian, &longitude); err != nil {
			return nil, err
		}

		packet_struct := PostitioningPacket{
			Latitude:  float32(latitude) / (30000.0 * 60.0),
			Longitude: float32(longitude) / (30000.0 * 60.0),
			Speed:     packet[19],
			Heading:   uint16(packet[21]) | (uint16(packet[20]&3) << 8),
			Timestamp: time.Now().Unix(),
		}

		return packet_struct.Process(device_data, protocol_number), nil
	case 0x13:
		if packet_length != 5 && packet_length != 6 {
			return nil, errors.New("invalid packet length")
		} else if !device_data.IsLoggedIn {
			return nil, errors.New("device is not logged in")
		}

		packet_struct := StatusPacket{
			BatteryPower: packet[4],
		}

		return packet_struct.Process(device_data), nil
	case 0x30:
		if packet_length != 1 {
			return nil, errors.New("invalid packet length")
		} else if !device_data.IsLoggedIn {
			return nil, errors.New("device is not logged in")
		}

		timeNow := time.Now()

		return []byte{0x78, 0x78, 7, protocol_number, byte(timeNow.Year() - 2000), byte(timeNow.Month()), byte(timeNow.Day()), byte(timeNow.Hour()), byte(timeNow.Minute()), byte(timeNow.Second()), 0x0d, 0x0a}, nil
	case 0x81, 0x83:
		if packet_length != 1 {
			return nil, errors.New("invalid packet length")
		} else if !device_data.IsLoggedIn {
			return nil, errors.New("device is not logged in")
		}

		device_data.InChargingState = false

		return []byte{}, nil
	case 0x82:
		if packet_length != 1 {
			return nil, errors.New("invalid packet length")
		} else if !device_data.IsLoggedIn {
			return nil, errors.New("device is not logged in")
		}

		device_data.InChargingState = true

		return []byte{}, nil
	}
}

func ImeiToString(imei []byte) string {
	str := ""
	for _, v := range imei {
		str += strconv.Itoa(int(v))
	}
	return str
}
