package dlms

import "log"

type DataChannel chan []byte

//go:generate mockery --name Transport --structname TransportMock --filename transportMock.go

// Transport specifies the transport layer.
type Transport interface {
	Close()
	Connect() (err error)
	Disconnect() (err error)
	IsConnected() bool
	SetAddress(client int, server int)
	SetReception(dc DataChannel)
	Send(src []byte) error
	SetLogger(logger *log.Logger)
}

// TransportWithBroadcast are optional methods for transport layers with broadcast capabilities
type TransportWithBroadcast interface {
	SendBroadcast(src []byte) error
}
