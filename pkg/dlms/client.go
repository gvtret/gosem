package dlms

import (
	"log"
	"time"
)

//go:generate mockery --name Client --structname ClientMock --filename clientMock.go

// Client specifies the client layer.
type Client interface {
	Connect() error
	Disconnect() error
	IsConnected() bool
	SetLogger(logger *log.Logger)
	Associate() error
	CloseAssociation() error
	IsAssociated() bool
	GetRequest(att *AttributeDescriptor, data interface{}) (err error)
	GetRequestWithSelectiveAccessByDate(att *AttributeDescriptor, start time.Time, end time.Time, data interface{}) (err error)
	GetRequestWithStructOfElements(data interface{}) (err error)
	SetRequest(att *AttributeDescriptor, data interface{}) (err error)
	ActionRequest(mth *MethodDescriptor, data interface{}) (err error)
}
