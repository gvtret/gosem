package dlms

import "log"

//go:generate mockery --name Transport --structname TransportMock --filename transportMock.go

// Transport specifies the transport layer.
type Transport interface {
	Connect() (err error)
	Disconnect() (err error)
	IsConnected() bool
	SetAddress(client int, server int)
	Send(src []byte) (out []byte, err error)
	SetLogger(logger *log.Logger)
}
