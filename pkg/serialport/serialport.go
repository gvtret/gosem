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
	wg          sync.WaitGroup
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
	sp.disconnect() // This already acquires mutex, sets isConnected=false, closes port
	sp.wg.Wait()   // Wait for manager goroutine to exit

	sp.mutex.Lock() // Lock specifically for dc manipulation
	defer sp.mutex.Unlock()
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

	sp.wg.Add(1)
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
	defer sp.wg.Done()
	for {
		// Early exit if not connected, reduces lock contention.
		// The critical check is before sending to sp.dc.
		sp.mutex.Lock()
		isConnected := sp.isConnected
		sp.mutex.Unlock()
		if !isConnected {
			return
		}

		data, err := sp.read() // This can block.
		if err != nil {
			// disconnect acquires its own lock
			sp.disconnect()
			return // Exit manager if read fails or port is closed.
		}

		sp.mutex.Lock()
		if sp.isConnected && len(data) > 0 && sp.dc != nil {
			sp.dc <- data
		}
		sp.mutex.Unlock()
	}
}

func (sp *serialport) disconnect() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	if sp.isConnected {
		sp.isConnected = false
		if sp.port != nil {
			sp.port.Close() // This helps unblock manager's read
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
