package main

import (
	"log"
	"time"

	"github.com/Circutor/gosem/pkg/client"
	"github.com/Circutor/gosem/pkg/dlms"
	"github.com/Circutor/gosem/pkg/tcp"
	"github.com/Circutor/gosem/pkg/wrapper"
)

type Instantaneous struct {
	Clock time.Time

	VoltageR        uint
	CurrentR        int32
	PowerFactorR    uint
	ActiveQuadrantR uint8
	ActivePowerR    uint32
	ReactivePowerR  uint32

	VoltageS        uint16
	CurrentS        int32
	PowerFactorS    uint16
	ActiveQuadrantS uint8
	ActivePowerS    uint32
	ReactivePowerS  uint32

	VoltageT        uint16
	CurrentT        int32
	PowerFactorT    uint16
	ActiveQuadrantT uint8
	ActivePowerT    uint32
	ReactivePowerT  uint32

	CurrentTotal     uint16
	PowerFactorTotal int16
	PhasePresence    uint8
	ActiveQuadrant   uint8

	Temperature int16
	RefVoltage  uint16
}

func main() {
	settings, err := dlms.NewSettingsWithLowAuthentication([]byte("TSCLBT01"))
	if err != nil {
		panic(err)
	}

	t := tcp.New(4059, "10.0.120.57", 1*time.Second)
	w := wrapper.New(t, 1, 1)
	c := client.New(settings, w)

	err = c.Connect()
	if err != nil {
		panic(err)
	}
	defer c.Disconnect()

	err = c.Associate()
	if err != nil {
		panic(err)
	}

	var timeZone int16

	attTimeZone := dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3)
	err = c.GetWithUnmarshal(attTimeZone, &timeZone)
	if err != nil {
		panic(err)
	}

	log.Printf("Time zone: %d\n", timeZone)

	var instantaneous []Instantaneous

	attInstantaneous := dlms.CreateAttributeDescriptor(7, "0-0:21.0.5.255", 2)
	err = c.GetWithUnmarshal(attInstantaneous, &instantaneous)
	if err != nil {
		panic(err)
	}

	log.Printf("Instantaneous: %v\n", instantaneous)
}
