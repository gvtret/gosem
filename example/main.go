package main

import (
	"gosem/pkg/client"
	"gosem/pkg/dlms"
	"gosem/pkg/tcp"
	"gosem/pkg/wrapper"
	"time"
)

func main() {
	settings, err := dlms.NewSettingsWithLowAuthentication([]byte("00000001"))
	if err != nil {
		panic(err)
	}

	t := tcp.New(459, "10.0.120.217", 1*time.Second)
	w := wrapper.New(t, 1, 1)
	s := client.New(settings, w)

	err = s.Connect()
	if err != nil {
		panic(err)
	}
	defer s.Disconnect()

	err = s.Associate()
	if err != nil {
		panic(err)
	}
}
