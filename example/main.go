package main

import (
	"log"
	"time"

	"github.com/Circutor/gosem/pkg/dlms"
	"github.com/Circutor/gosem/pkg/dlmsclient"
	"github.com/Circutor/gosem/pkg/tcp"
	"github.com/Circutor/gosem/pkg/wrapper"
)

type Instantaneous struct {
	Clock time.Time

	VoltageR        uint16
	CurrentR        int32
	PowerFactorR    uint16
	ActiveQuadrantR uint8
	ActivePowerR    int32
	ReactivePowerR  int32

	VoltageS        uint16
	CurrentS        int32
	PowerFactorS    uint16
	ActiveQuadrantS uint8
	ActivePowerS    int32
	ReactivePowerS  int32

	VoltageT        uint16
	CurrentT        int32
	PowerFactorT    uint16
	ActiveQuadrantT uint8
	ActivePowerT    int32
	ReactivePowerT  int32

	CurrentTotal     int16
	PowerFactorTotal uint16
	PhasePresence    uint8
	ActiveQuadrant   uint8

	Temperature int16
	RefVoltage  uint16
}

func main() {
	l := log.New(log.Writer(), "", log.Ldate|log.Ltime|log.Lmicroseconds)

	settings, err := dlms.NewSettingsWithLowAuthentication([]byte("TSCLBT01"))
	if err != nil {
		panic(err)
	}

	t := tcp.New(4059, "10.0.120.216", 1*time.Second)
	w := wrapper.New(t, 1, 1)
	c := dlmsclient.New(settings, w, 1*time.Second, 0)
	c.SetLogger(l)

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
	err = c.GetRequest(attTimeZone, &timeZone)
	if err != nil {
		panic(err)
	}

	log.Printf("Time zone: %d\n", timeZone)

	var timeDate time.Time

	attTime := dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 2)
	err = c.GetRequest(attTime, &timeDate)
	if err != nil {
		panic(err)
	}

	log.Printf("Time: %v\n", timeDate)

	var instantaneous []Instantaneous

	attInstantaneous := dlms.CreateAttributeDescriptor(7, "0-0:21.0.5.255", 2)
	err = c.GetRequest(attInstantaneous, &instantaneous)
	if err != nil {
		panic(err)
	}

	log.Printf("Instantaneous: %v\n", instantaneous)

	type ConfigReport struct {
		Time          time.Time       `obis:"8,0-0:1.0.0.255,2"`
		TimeZone      int16           `obis:"8,0-0:1.0.0.255,3"`
		Instantaneous []Instantaneous `obis:"7,0-0:21.0.5.255,2"`
	}

	var configReport ConfigReport

	err = c.GetRequestWithStructOfElements(&configReport)
	if err != nil {
		panic(err)
	}

	log.Printf("Config report: %v\n", configReport)
}
