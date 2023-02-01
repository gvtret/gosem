package dlmsclient

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Circutor/gosem/pkg/dlms"
)

const (
	unicastInvokeID = 0xC1
)

type client struct {
	settings           dlms.Settings
	transport          dlms.Transport
	replyTimeout       time.Duration
	associationTimeout time.Duration
	isAssociated       bool
	timeoutTimer       *time.Timer
	tc                 dlms.DataChannel
	dc                 dlms.DataChannel
	dataNotification   chan dlms.DataNotification
	mutex              sync.Mutex
	subsMutex          sync.Mutex
}

func New(settings dlms.Settings, transport dlms.Transport, replyTimeout time.Duration, associationTimeout time.Duration) dlms.Client {
	c := &client{
		settings:           settings,
		transport:          transport,
		replyTimeout:       replyTimeout,
		associationTimeout: associationTimeout,
		isAssociated:       false,
		timeoutTimer:       nil,
		tc:                 make(dlms.DataChannel, 10),
		dc:                 nil,
		dataNotification:   nil,
		mutex:              sync.Mutex{},
		subsMutex:          sync.Mutex{},
	}

	transport.SetReception(c.tc)

	go c.manager()

	return c
}

func (c *client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.transport.Connect()
	if err != nil {
		return dlms.NewError(dlms.ErrorCommunicationFailed, fmt.Sprintf("error connecting: %v", err))
	}

	if c.associationTimeout != 0 {
		c.timeoutTimer = time.AfterFunc(c.associationTimeout, func() {
			c.Disconnect()
		})
	}

	return nil
}

func (c *client) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.closeAssociation()

	err := c.transport.Disconnect()
	if err != nil {
		return dlms.NewError(dlms.ErrorCommunicationFailed, fmt.Sprintf("error disconnecting: %v", err))
	}

	return nil
}

func (c *client) IsConnected() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.transport.IsConnected()
}

func (c *client) SetDataNotificationChannel(ec chan dlms.DataNotification) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.dataNotification = ec
}

func (c *client) GetSettings() dlms.Settings {
	return c.settings
}

func (c *client) SetSettings(settings dlms.Settings) {
	c.settings = settings
}

func (c *client) SetLogger(logger *log.Logger) {
	c.transport.SetLogger(logger)
}

func (c *client) Associate() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.transport.IsConnected() {
		return dlms.NewError(dlms.ErrorInvalidState, "not connected")
	}

	if c.isAssociated {
		return dlms.NewError(dlms.ErrorInvalidState, "already associated")
	}

	src, err := dlms.EncodeAARQ(&c.settings)
	if err != nil {
		return dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("error encoding AARQ: %v", err))
	}

	out, err := c.sendReceive(src)
	if err != nil {
		return err
	}

	aare, err := dlms.DecodeAARE(&c.settings, &out)
	if err != nil {
		return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("error decoding AARE: %v", err))
	}

	if aare.AssociationResult != dlms.AssociationResultAccepted || aare.SourceDiagnostic != dlms.SourceDiagnosticNone || aare.InitiateResponse == nil {
		if aare.SourceDiagnostic == dlms.SourceDiagnosticAuthenticationFailure {
			return dlms.NewError(dlms.ErrorInvalidPassword, fmt.Sprintf("association failed (invalid password): %d - %d", aare.AssociationResult, aare.SourceDiagnostic))
		}

		if aare.ConfirmedServiceError != nil {
			return dlms.NewError(dlms.ErrorAuthenticationFailed, fmt.Sprintf("association failed: %d - %d (%s)", aare.AssociationResult, aare.SourceDiagnostic, aare.ConfirmedServiceError.String()))
		}
		return dlms.NewError(dlms.ErrorAuthenticationFailed, fmt.Sprintf("association failed: %d - %d", aare.AssociationResult, aare.SourceDiagnostic))
	}

	c.isAssociated = true
	return nil
}

func (c *client) CloseAssociation() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.transport.IsConnected() {
		return dlms.NewError(dlms.ErrorInvalidState, "not connected")
	}

	if !c.isAssociated {
		return dlms.NewError(dlms.ErrorInvalidState, "not associated")
	}

	src, err := dlms.EncodeRLRQ(&c.settings)
	if err != nil {
		return dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("error encoding RLRQ: %v", err))
	}

	out, err := c.sendReceive(src)
	if err != nil {
		return err
	}

	_, err = dlms.DecodeRLRE(&c.settings, &out)
	if err != nil {
		return dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("error decoding RLRE: %v", err))
	}

	c.closeAssociation()

	return nil
}

func (c *client) IsAssociated() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.transport.IsConnected() {
		c.isAssociated = false
	}

	return c.isAssociated
}

func (c *client) manager() {
	for {
		data := <-c.tc

		dn, err := dlms.DecodeDataNotification(&data)
		if err == nil {
			c.mutex.Lock()
			if c.dataNotification != nil {
				c.dataNotification <- dn
			}
			c.mutex.Unlock()
		} else {
			c.subsMutex.Lock()
			if c.dc != nil {
				c.dc <- data
			}
			c.subsMutex.Unlock()
		}
	}
}

func (c *client) sendReceive(src []byte) ([]byte, error) {
	c.subscribe()
	defer c.unsubscribe()

	err := c.transport.Send(src)
	if err != nil {
		return nil, dlms.NewError(dlms.ErrorCommunicationFailed, fmt.Sprintf("error sending AARQ: %v", err))
	}

	// Wait for the device response
	timeout := time.NewTimer(c.replyTimeout)
	defer timeout.Stop()

	select {
	case data := <-c.dc:
		return data, nil
	case <-timeout.C:
		return nil, dlms.NewError(dlms.ErrorCommunicationFailed, "timeout reached")
	}
}

func (c *client) subscribe() {
	c.subsMutex.Lock()
	defer c.subsMutex.Unlock()

	c.dc = make(dlms.DataChannel)
}

func (c *client) unsubscribe() {
	c.subsMutex.Lock()
	defer c.subsMutex.Unlock()

	c.dc = nil
}

func (c *client) encodeSendReceiveAndDecode(req dlms.CosemPDU) (pdu dlms.CosemPDU, err error) {
	if !c.isAssociated {
		err = dlms.NewError(dlms.ErrorInvalidState, "client is not associated")
		return
	}

	src, err := req.Encode()
	if err != nil {
		err = dlms.NewError(dlms.ErrorInvalidParameter, fmt.Sprintf("error encoding PDU: %v", err))
		return
	}

	out, err := c.sendReceive(src)
	if err != nil {
		if !c.transport.IsConnected() {
			c.closeAssociation()
		}

		return
	}

	pdu, err = dlms.DecodeCosem(&out)
	if err != nil {
		err = dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("error decoding PDU: %v", err))
		return
	}

	if c.timeoutTimer != nil {
		c.timeoutTimer.Reset(c.associationTimeout)
	}

	return
}

func (c *client) closeAssociation() {
	c.isAssociated = false
	if c.timeoutTimer != nil {
		c.timeoutTimer.Stop()
		c.timeoutTimer = nil
	}
}
