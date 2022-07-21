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
	settings     dlms.Settings
	transport    dlms.Transport
	timeout      time.Duration
	isAssociated bool
	timeoutTimer *time.Timer
	mutex        sync.Mutex
}

func New(settings dlms.Settings, transport dlms.Transport, timeout time.Duration) dlms.Client {
	c := &client{
		settings:     settings,
		transport:    transport,
		timeout:      timeout,
		isAssociated: false,
		timeoutTimer: nil,
		mutex:        sync.Mutex{},
	}

	return c
}

func (c *client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.transport.Connect()
	if err != nil {
		return dlms.NewError(dlms.ErrorCommunicationFailed, fmt.Sprintf("error connecting: %v", err))
	}

	if c.timeout != 0 {
		c.timeoutTimer = time.AfterFunc(c.timeout, func() {
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

	out, err := c.transport.Send(src)
	if err != nil {
		return dlms.NewError(dlms.ErrorCommunicationFailed, fmt.Sprintf("error sending AARQ: %v", err))
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

	out, err := c.transport.Send(src)
	if err != nil {
		return dlms.NewError(dlms.ErrorCommunicationFailed, fmt.Sprintf("error sending RLRQ: %v", err))
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

	out, err := c.transport.Send(src)
	if err != nil {
		if !c.transport.IsConnected() {
			c.closeAssociation()
		}

		err = dlms.NewError(dlms.ErrorCommunicationFailed, fmt.Sprintf("error sending PDU: %v", err))
		return
	}

	pdu, err = dlms.DecodeCosem(&out)
	if err != nil {
		err = dlms.NewError(dlms.ErrorInvalidResponse, fmt.Sprintf("error decoding PDU: %v", err))
		return
	}

	if c.timeoutTimer != nil {
		c.timeoutTimer.Reset(c.timeout)
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
