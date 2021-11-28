package tcp

import (
	"fmt"
	"net"
	"strconv"
	"time"
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
}

func New(port int, host string, timeout time.Duration) *TCP {
	t := &TCP{
		port:        port,
		host:        host,
		timeout:     timeout,
		isConnected: false,
	}

	return t
}

func (t *TCP) Connect() error {
	if !t.isConnected {
		address := net.JoinHostPort(t.host, strconv.Itoa(t.port))

		conn, err := net.DialTimeout("tcp", address, t.timeout)
		if err != nil {
			return fmt.Errorf("connect failed: %w", err)
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
			return fmt.Errorf("disconnect failed: %w", err)
		}

		t.isConnected = false
	}

	return nil
}

func (t *TCP) IsConnected() bool {
	return t.isConnected
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

	out := make([]byte, maxLength)

	n, err := t.conn.Read(out)
	if err != nil {
		t.Disconnect()
		return nil, fmt.Errorf("read failed: %w", err)
	}

	return out[:n], nil
}
