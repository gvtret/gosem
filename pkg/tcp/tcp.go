package tcp

import (
	"encoding/hex"
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

type TCP struct {
	port        int
	host        string
	timeout     time.Duration
	conn        net.Conn
	isConnected bool
	logger      *log.Logger
}

func New(port int, host string, timeout time.Duration) dlms.Transport {
	t := &TCP{
		port:        port,
		host:        host,
		timeout:     timeout,
		isConnected: false,
		logger:      nil,
	}

	return t
}

func (t *TCP) Connect() error {
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

		t.conn = conn
		t.isConnected = true
	}

	return nil
}

func (t *TCP) Disconnect() error {
	if t.isConnected {
		err := t.conn.Close()
		if err != nil {
			if t.logger != nil {
				t.logger.Printf("Disconnect from %s failed: %v", t.host, err)
			}

			return fmt.Errorf("disconnect failed: %w", err)
		}

		if t.logger != nil {
			t.logger.Printf("Disconnected from %s", t.host)
		}

		t.isConnected = false
	}

	return nil
}

func (t *TCP) IsConnected() bool {
	return t.isConnected
}

func (t *TCP) SetAddress(client int, server int) {
}

func (t *TCP) Send(src []byte) ([]byte, error) {
	if !t.isConnected {
		return nil, fmt.Errorf("not connected")
	}

	t.conn.SetDeadline(time.Now().Add(t.timeout))

	_, err := t.conn.Write(src)
	if err != nil {
		t.Disconnect()
		return nil, fmt.Errorf("write failed: %w", err)
	}

	if t.logger != nil {
		t.logger.Printf("TX (%s): %s", t.host, encodeHexString(src))
	}

	out := make([]byte, maxLength)

	n, err := t.conn.Read(out)
	if err != nil {
		t.Disconnect()
		return nil, fmt.Errorf("read failed: %w", err)
	}

	if t.logger != nil {
		t.logger.Printf("RX (%s): %s", t.host, encodeHexString(out[:n]))
	}

	return out[:n], nil
}

func (t *TCP) SetLogger(logger *log.Logger) {
	t.logger = logger
}

func encodeHexString(b []byte) string {
	return strings.ToUpper(hex.EncodeToString(b))
}
