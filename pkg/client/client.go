package client

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
		return fmt.Errorf("error connecting: %w", err)
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

	return c.transport.Disconnect()
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
		return fmt.Errorf("not connected")
	}

	if c.isAssociated {
		return fmt.Errorf("already associated")
	}

	src, err := dlms.EncodeAARQ(&c.settings)
	if err != nil {
		return fmt.Errorf("error encoding AARQ: %w", err)
	}

	out, err := c.transport.Send(src)
	if err != nil {
		return fmt.Errorf("error sending AARQ: %w", err)
	}

	aare, err := dlms.DecodeAARE(&c.settings, &out)
	if err != nil {
		return fmt.Errorf("error decoding AARE: %w", err)
	}

	if aare.AssociationResult != dlms.AssociationResultAccepted && aare.SourceDiagnostic != dlms.SourceDiagnosticNone && aare.InitiateResponse != nil {
		if aare.ConfirmedServiceError != nil {
			return fmt.Errorf("association failed: %d - %d (%s)", aare.AssociationResult, aare.SourceDiagnostic, aare.ConfirmedServiceError.String())
		}
		return fmt.Errorf("association failed: %d - %d", aare.AssociationResult, aare.SourceDiagnostic)
	}

	c.isAssociated = true
	return nil
}

func (c *Client) CloseAssociation() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.transport.IsConnected() {
		return fmt.Errorf("not connected")
	}

	if !c.isAssociated {
		return fmt.Errorf("not associated")
	}

	src, err := dlms.EncodeRLRQ(&c.settings)
	if err != nil {
		return fmt.Errorf("error encoding RLRQ: %w", err)
	}

	out, err := c.transport.Send(src)
	if err != nil {
		return fmt.Errorf("error sending RLRQ: %w", err)
	}

	_, err = dlms.DecodeRLRE(&c.settings, &out)
	if err != nil {
		return fmt.Errorf("error decoding RLRE: %w", err)
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
		err = fmt.Errorf("client is not associated")
		return
	}

	src, err := req.Encode()
	if err != nil {
		err = fmt.Errorf("error encoding CosemPDU: %w", err)
		return
	}

	out, err := c.transport.Send(src)
	if err != nil {
		if !c.transport.IsConnected() {
			c.closeAssociation()
		}

		err = fmt.Errorf("error sending CosemPDU: %w", err)
		return
	}

	pdu, err = dlms.DecodeCosem(&out)
	if err != nil {
		err = fmt.Errorf("error decoding CosemPDU: %w", err)
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
