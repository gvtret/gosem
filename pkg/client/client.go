package client

import (
	"fmt"
	"gosem/pkg/dlms"
)

type Client struct {
	settings     dlms.Settings
	transport    dlms.Transport
	isConnected  bool
	isAssociated bool
}

func New(settings dlms.Settings, transport dlms.Transport) (Client, error) {
	c := Client{
		settings:     settings,
		transport:    transport,
		isConnected:  false,
		isAssociated: false,
	}

	return c, nil
}

func (c *Client) Connect() error {
	if c.isConnected {
		return fmt.Errorf("already connected")
	}

	err := c.transport.Connect()
	if err != nil {
		return fmt.Errorf("error connecting: %w", err)
	}

	c.isConnected = true

	return nil
}

func (c *Client) Disconnect() error {
	if !c.isConnected {
		return fmt.Errorf("not connected")
	}

	err := c.transport.Disconnect()
	if err != nil {
		return fmt.Errorf("error disconnecting: %w", err)
	}

	c.isConnected = false
	c.isAssociated = false

	return nil
}

func (c *Client) IsConnected() bool {
	return c.isConnected
}

func (c *Client) Associate() error {
	if !c.isConnected {
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

func (c *Client) IsAssociated() bool {
	return c.isAssociated
}
