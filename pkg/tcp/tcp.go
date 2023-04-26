package tcp

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Circutor/gosem/pkg/dlms"
)

const (
	maxLength = 2048
)

type tcp struct {
	port        int
	host        string
	timeout     time.Duration
	dc          dlms.DataChannel
	conn        net.Conn
	isConnected bool
	logger      *log.Logger
}

func New(port int, host string, timeout time.Duration) dlms.Transport {
	t := &tcp{
		port:        port,
		host:        host,
		timeout:     timeout,
		dc:          nil,
		isConnected: false,
		logger:      nil,
	}

	return t
}

func (t *tcp) Close() {
	if t.logger != nil {
		t.logger.Printf("Close received, disconnect from %s", t.host)
	}

	t.Disconnect()
	if t.dc != nil {
		close(t.dc)
		t.dc = nil
	}
}

func (t *tcp) Connect() error {
	if !t.isConnected {
		address := net.JoinHostPort(t.host, strconv.Itoa(t.port))

		conn, err := net.DialTimeout("tcp", address, t.timeout)
		if err != nil {
			if t.logger != nil {
				t.logger.Printf("Connect to %s failed: %v", address, err)
			}

			return fmt.Errorf("connect failed: %w", err)
		}

		if t.logger != nil {
			t.logger.Printf("Connected to %s", address)
		}

		go t.manager()

		t.conn = conn
		t.isConnected = true
	}

	return nil
}

func (t *tcp) Disconnect() error {
	if t.isConnected {
		t.isConnected = false

		if t.conn != nil {
			t.conn.Close()
			t.conn = nil
		}

		if t.logger != nil {
			t.logger.Printf("Disconnected from %s", t.host)
		}
	}

	return nil
}

func (t *tcp) IsConnected() bool {
	return t.isConnected
}

func (t *tcp) SetAddress(client int, server int) {
}

func (t *tcp) SetReception(dc dlms.DataChannel) {
	t.dc = dc
}

func (t *tcp) Send(src []byte) error {
	if !t.isConnected {
		return fmt.Errorf("not connected")
	}

	t.conn.SetWriteDeadline(time.Now().Add(t.timeout))

	_, err := t.conn.Write(src)
	if err != nil {
		if t.logger != nil {
			t.logger.Printf("Write failed (%v), disconnect from %s", err, t.host)
		}

		t.Disconnect()
		return fmt.Errorf("write failed: %w", err)
	}

	if t.logger != nil {
		t.logger.Printf("TX (%s): %s", t.host, encodeHexString(src))
	}

	return nil
}

func (t *tcp) SetLogger(logger *log.Logger) {
	t.logger = logger
}

func (t *tcp) manager() {
	for {
		if !t.isConnected {
			return
		}

		data, err := t.read()
		if err != nil {
			if t.logger != nil {
				t.logger.Printf("Read failed (%v), disconnect from %s", err, t.host)
			}

			t.Disconnect()

			return
		}

		if len(data) > 0 && t.dc != nil {
			t.dc <- data
		}
	}
}

func (t *tcp) read() ([]byte, error) {
	rxBuffer := make([]byte, maxLength)

	rxLen, err := t.conn.Read(rxBuffer)
	if err != nil {
		var netErr net.Error
		if !errors.As(err, &netErr) || !netErr.Timeout() {
			return nil, fmt.Errorf("read error: %w", err)
		}
	}

	if t.logger != nil {
		t.logger.Printf("RX (%s): %s", t.host, encodeHexString(rxBuffer[:rxLen]))
	}

	return rxBuffer[:rxLen], nil
}

func encodeHexString(b []byte) string {
	return strings.ToUpper(hex.EncodeToString(b))
}
