package serialport

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"

	"gitlab.com/circutor-library/gosem/pkg/dlms"
	"go.bug.st/serial"
)

const (
	maxLength = 2048
)

type serialport struct {
	serialPort  string
	baudRate    int
	dc          dlms.DataChannel
	port        serial.Port
	isConnected bool
	logger      *log.Logger
	mutex       sync.Mutex
}

func New(serialPort string, baudRate int) dlms.Transport {
	sp := &serialport{
		serialPort:  serialPort,
		baudRate:    baudRate,
		dc:          nil,
		port:        nil,
		isConnected: false,
		logger:      nil,
		mutex:       sync.Mutex{},
	}

	return sp
}

func (sp *serialport) Close() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	sp.disconnect()
	if sp.dc != nil {
		close(sp.dc)
		sp.dc = nil
	}
}

func (sp *serialport) Connect() error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	if sp.isConnected {
		return nil
	}

	mode := &serial.Mode{
		BaudRate: sp.baudRate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(sp.serialPort, mode)
	if err != nil {
		return fmt.Errorf("failed to open port %s: %w", sp.serialPort, err)
	}

	sp.port = port
	sp.isConnected = true

	go sp.manager()

	return nil
}

func (sp *serialport) Disconnect() error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	sp.disconnect()

	return nil
}

func (sp *serialport) IsConnected() bool {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	return sp.isConnected
}

func (sp *serialport) SetAddress(_ int, _ int) {
}

func (sp *serialport) SetReception(dc dlms.DataChannel) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	if sp.dc != nil {
		close(sp.dc)
	}

	sp.dc = dc
}

func (sp *serialport) Send(src []byte) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	if !sp.isConnected {
		return fmt.Errorf("not connected")
	}

	_, err := sp.port.Write(src)
	if err != nil {
		sp.disconnect()
		return fmt.Errorf("write failed: %w", err)
	}

	if sp.logger != nil {
		sp.logger.Printf("TX (%s): %s", sp.serialPort, encodeHexString(src))
	}

	return nil
}

func (sp *serialport) SetLogger(logger *log.Logger) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	sp.logger = logger
}

func (sp *serialport) manager() {
	for {
		if !sp.isConnected {
			return
		}

		data, err := sp.read()
		if err != nil {
			sp.mutex.Lock()
			sp.disconnect()
			sp.mutex.Unlock()

			return
		}

		if len(data) > 0 && sp.dc != nil {
			sp.dc <- data
		}
	}
}

func (sp *serialport) disconnect() {
	if sp.isConnected {
		sp.isConnected = false

		if sp.port != nil {
			sp.port.Close()
			sp.port = nil
		}

		if sp.logger != nil {
			sp.logger.Printf("Closed port %s", sp.serialPort)
		}
	}
}

func (sp *serialport) read() ([]byte, error) {
	rxBuffer := make([]byte, maxLength)

	port := sp.port
	if port == nil {
		return nil, fmt.Errorf("connection is nil")
	}

	rxLen, err := port.Read(rxBuffer)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	if sp.logger != nil && rxLen > 0 {
		sp.logger.Printf("RX (%s): %s", sp.serialPort, encodeHexString(rxBuffer[:rxLen]))
	}

	return rxBuffer[:rxLen], nil
}

func encodeHexString(b []byte) string {
	return strings.ToUpper(hex.EncodeToString(b))
}
