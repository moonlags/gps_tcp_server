package tcp

import (
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

type DeviceData struct {
	IMEI                 string
	Postitions           []PostitioningPacket
	BatteryPower         uint8
	LastStatusPacketTime int64
	StatusCooldown       int32
	IsLoggedIn           bool
	InChargingState      bool
	Connection           net.Conn
}

type LoginPacket struct {
	IMEI string
}

func (self LoginPacket) Process(device_data *DeviceData) []byte {
	log.WithField("IMEI", self.IMEI).Info("device logged in")

	device_data.IMEI = self.IMEI
	device_data.IsLoggedIn = true

	return []byte{0x78, 0x78, 1, 1, 0x0d, 0x0a}
}

type PostitioningPacket struct {
	Latitude  float32
	Longitude float32
	Speed     uint8
	Heading   uint16
	Timestamp int64
}

func (self PostitioningPacket) Process(device_data *DeviceData, protocol_number byte) []byte {
	log.WithField("positioning packet", self).WithField("IMEI", device_data.IMEI).Info("new position")

	device_data.Postitions = append(device_data.Postitions, self)
	if len(device_data.Postitions) > 10 {
		device_data.Postitions = device_data.Postitions[1:]
	}

	timeNow := time.Now()

	return []byte{0x78, 0x78, 7, protocol_number, byte(timeNow.Year() - 2000), byte(timeNow.Month()), byte(timeNow.Day()), byte(timeNow.Hour()), byte(timeNow.Minute()), byte(timeNow.Second()), 0x0d, 0x0a}
}

type StatusPacket struct {
	BatteryPower uint8
}

func (self StatusPacket) Process(device_data *DeviceData) []byte {
	log.WithField("status packet", self).WithField("IMEI", device_data.IMEI).Info("status packet")

	device_data.BatteryPower = self.BatteryPower

	cooldown := 1

	if device_data.BatteryPower < 15 {
		cooldown = 5
	}

	return []byte{0x78, 0x78, 2, 0x13, byte(cooldown), 0x0d, 0x0a}
}
