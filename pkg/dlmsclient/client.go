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

type Client struct {
	settings     dlms.Settings
	transport    dlms.Transport
	timeout      time.Duration
	isAssociated bool
	timeoutTimer *time.Timer
	mutex        sync.Mutex
}

func New(settings dlms.Settings, transport dlms.Transport, timeout time.Duration) *Client {
	c := &Client{
		settings:     settings,
		transport:    transport,
		timeout:      timeout,
		isAssociated: false,
		timeoutTimer: nil,
		mutex:        sync.Mutex{},
	}

	return c
}

func (c *Client) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.transport.Connect()
	if err != nil {
		return NewError(ErrorCommunicationFailed, fmt.Sprintf("error connecting: %v", err))
	}

	if c.timeout != 0 {
		c.timeoutTimer = time.AfterFunc(c.timeout, func() {
			c.Disconnect()
		})
	}

	return nil
}

func (c *Client) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.closeAssociation()

	err := c.transport.Disconnect()
	if err != nil {
		return NewError(ErrorCommunicationFailed, fmt.Sprintf("error disconnecting: %v", err))
	}

	return nil
}

func (c *Client) IsConnected() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.transport.IsConnected()
}

func (c *Client) SetLogger(logger *log.Logger) {
	c.transport.SetLogger(logger)
}

func (c *Client) Associate() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.transport.IsConnected() {
		return NewError(ErrorInvalidState, "not connected")
	}

	if c.isAssociated {
		return NewError(ErrorInvalidState, "already associated")
	}

	src, err := dlms.EncodeAARQ(&c.settings)
	if err != nil {
		return NewError(ErrorInvalidParameter, fmt.Sprintf("error encoding AARQ: %v", err))
	}

	out, err := c.transport.Send(src)
	if err != nil {
		return NewError(ErrorCommunicationFailed, fmt.Sprintf("error sending AARQ: %v", err))
	}

	aare, err := dlms.DecodeAARE(&c.settings, &out)
	if err != nil {
		return NewError(ErrorInvalidResponse, fmt.Sprintf("error decoding AARE: %v", err))
	}

	if aare.AssociationResult != dlms.AssociationResultAccepted && aare.SourceDiagnostic != dlms.SourceDiagnosticNone && aare.InitiateResponse != nil {
		if aare.ConfirmedServiceError != nil {
			return NewError(ErrorAuthenticationFailed, fmt.Sprintf("association failed: %d - %d (%s)", aare.AssociationResult, aare.SourceDiagnostic, aare.ConfirmedServiceError.String()))
		}
		return NewError(ErrorAuthenticationFailed, fmt.Sprintf("association failed: %d - %d", aare.AssociationResult, aare.SourceDiagnostic))
	}

	c.isAssociated = true
	return nil
}

func (c *Client) CloseAssociation() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.transport.IsConnected() {
		return NewError(ErrorInvalidState, "not connected")
	}

	if !c.isAssociated {
		return NewError(ErrorInvalidState, "not associated")
	}

	src, err := dlms.EncodeRLRQ(&c.settings)
	if err != nil {
		return NewError(ErrorInvalidParameter, fmt.Sprintf("error encoding RLRQ: %v", err))
	}

	out, err := c.transport.Send(src)
	if err != nil {
		return NewError(ErrorCommunicationFailed, fmt.Sprintf("error sending RLRQ: %v", err))
	}

	_, err = dlms.DecodeRLRE(&c.settings, &out)
	if err != nil {
		return NewError(ErrorInvalidResponse, fmt.Sprintf("error decoding RLRE: %v", err))
	}

	c.closeAssociation()

	return nil
}

func (c *Client) IsAssociated() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.transport.IsConnected() {
		c.isAssociated = false
	}

	return c.isAssociated
}

func (c *Client) encodeSendReceiveAndDecode(req dlms.CosemPDU) (pdu dlms.CosemPDU, err error) {
	if !c.isAssociated {
		err = NewError(ErrorInvalidState, "client is not associated")
		return
	}

	src, err := req.Encode()
	if err != nil {
		err = NewError(ErrorInvalidParameter, fmt.Sprintf("error encoding PDU: %v", err))
		return
	}

	out, err := c.transport.Send(src)
	if err != nil {
		if !c.transport.IsConnected() {
			c.closeAssociation()
		}

		err = NewError(ErrorCommunicationFailed, fmt.Sprintf("error sending PDU: %v", err))
		return
	}

	pdu, err = dlms.DecodeCosem(&out)
	if err != nil {
		err = NewError(ErrorInvalidResponse, fmt.Sprintf("error decoding PDU: %v", err))
		return
	}

	if c.timeoutTimer != nil {
		c.timeoutTimer.Reset(c.timeout)
	}

	return
}

func (c *Client) closeAssociation() {
	c.isAssociated = false
	if c.timeoutTimer != nil {
		c.timeoutTimer.Stop()
		c.timeoutTimer = nil
	}
}
