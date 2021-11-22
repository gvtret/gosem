package wrapper

import (
	"encoding/binary"
	"fmt"
	"gosem/pkg/dlms"
)

const (
	version      = 1
	headerLength = 8
	maxLength    = 2048
)

type Wrapper struct {
	transport   dlms.Transport
	source      uint16
	destination uint16
	isConnected bool
}

func New(transport dlms.Transport, client int, server int) (Wrapper, error) {
	w := Wrapper{
		transport:   transport,
		source:      uint16(client),
		destination: uint16(server),
		isConnected: false,
	}

	return w, nil
}

func (w *Wrapper) Connect() error {
	if w.isConnected {
		return fmt.Errorf("already connected")
	}

	err := w.transport.Connect()
	if err != nil {
		return fmt.Errorf("error connecting: %w", err)
	}

	w.isConnected = true

	return nil
}

func (w *Wrapper) Disconnect() error {
	if !w.isConnected {
		return fmt.Errorf("not connected")
	}

	err := w.transport.Disconnect()
	if err != nil {
		return fmt.Errorf("error disconnecting: %w", err)
	}

	w.isConnected = false

	return nil
}

func (w *Wrapper) IsConnected() bool {
	return w.isConnected
}

func (w *Wrapper) Send(src []byte) ([]byte, error) {
	if !w.isConnected {
		return nil, fmt.Errorf("not connected")
	}

	if len(src) > (maxLength - headerLength) {
		return nil, fmt.Errorf("message too long")
	}

	uri := make([]byte, headerLength+len(src))

	binary.BigEndian.PutUint16(uri[0:2], uint16(version))
	binary.BigEndian.PutUint16(uri[2:4], w.source)
	binary.BigEndian.PutUint16(uri[4:6], w.destination)
	binary.BigEndian.PutUint16(uri[6:8], uint16(len(src)))

	copy(uri[headerLength:], src)

	out, err := w.transport.Send(uri)
	if err != nil {
		return nil, fmt.Errorf("error sending: %w", err)
	}

	err = w.parseHeader(out)
	if err != nil {
		return nil, fmt.Errorf("error parsing header: %w", err)
	}

	return out[headerLength:], nil
}

func (w *Wrapper) parseHeader(src []byte) error {
	if len(src) < headerLength {
		return fmt.Errorf("message too short")
	}

	if binary.BigEndian.Uint16(src[0:2]) != uint16(version) {
		return fmt.Errorf("invalid version")
	}

	if binary.BigEndian.Uint16(src[2:4]) != w.destination {
		return fmt.Errorf("invalid destination")
	}

	if binary.BigEndian.Uint16(src[4:6]) != w.source {
		return fmt.Errorf("invalid source")
	}

	length := int(binary.BigEndian.Uint16(src[6:8]))
	if length > maxLength {
		return fmt.Errorf("message too long")
	}

	if len(src) != (headerLength + length) {
		return fmt.Errorf("message length mismatch")
	}

	return nil
}
