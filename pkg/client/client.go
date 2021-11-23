package client

import (
	"fmt"
	"gosem/pkg/dlms"
)

type Client struct {
	settings     dlms.Settings
	transport    dlms.Transport
	isAssociated bool
}

func New(settings dlms.Settings, transport dlms.Transport) (*Client, error) {
	c := &Client{
		settings:     settings,
		transport:    transport,
		isAssociated: false,
	}

	return c, nil
}

func (c *Client) Connect() error {
	return c.transport.Connect()
}

func (c *Client) Disconnect() error {
	c.isAssociated = false

	return c.transport.Disconnect()
}

func (c *Client) IsConnected() bool {
	return c.transport.IsConnected()
}

func (c *Client) Associate() error {
	if !c.transport.IsConnected() {
		return fmt.Errorf("not connected")
	}

	if c.IsAssociated() {
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

func (c *Client) IsAssociated() bool {
	if !c.transport.IsConnected() {
		c.isAssociated = false
	}

	return c.isAssociated
}
