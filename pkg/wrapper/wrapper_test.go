package wrapper_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"gosem/pkg/dlms/mocks"
	"gosem/pkg/wrapper"
	"testing"
)

func TestWrapper_Connect(t *testing.T) {
	transportMock := new(mocks.TransportMock)
	transportMock.On("Connect").Return(nil)
	transportMock.On("Disconnect").Return(nil)

	w := wrapper.New(transportMock, 1, 3)

	err := w.Connect()
	if err != nil {
		t.Errorf("Error connecting: %s", err)
	}

	transportMock.On("IsConnected").Return(true)

	if !w.IsConnected() {
		t.Errorf("Wrapper is not connected")
	}

	err = w.Disconnect()
	if err != nil {
		t.Errorf("Error disconnecting: %s", err)
	}

	transportMock.AssertNumberOfCalls(t, "Connect", 1)
	transportMock.AssertNumberOfCalls(t, "IsConnected", 1)
}

func TestWrapper_Send(t *testing.T) {
	transportMock := new(mocks.TransportMock)
	transportMock.On("Connect").Return(nil)

	w := wrapper.New(transportMock, 1, 3)
	w.Connect()

	transportMock.On("IsConnected").Return(true)

	in := decodeHexString("0001000100030006AABBCCDDEEFF")
	out := decodeHexString("00010003000100050123456789")
	transportMock.On("Send", in).Return(out, nil)

	src := decodeHexString("AABBCCDDEEFF")
	out, err := w.Send(src)
	if err != nil {
		t.Errorf("Error sending: %s", err)
	}

	if !bytes.Equal(out, decodeHexString("0123456789")) {
		t.Errorf("Output is not as expected")
	}

	transportMock.AssertNumberOfCalls(t, "Connect", 1)
	transportMock.AssertNumberOfCalls(t, "IsConnected", 1)
	transportMock.AssertNumberOfCalls(t, "Send", 1)
}

func TestWrapper_SendFailed(t *testing.T) {
	transportMock := new(mocks.TransportMock)

	w := wrapper.New(transportMock, 1, 3)

	// Send failed
	in := decodeHexString("0001000100030006AABBCCDDEEFF")
	out := decodeHexString("00010003000100050123456789")
	transportMock.On("Send", in).Return(out, fmt.Errorf("send failed")).Once()
	transportMock.On("IsConnected").Return(true).Once()

	src := decodeHexString("AABBCCDDEEFF")
	_, err := w.Send(src)
	if err == nil {
		t.Errorf("Error expected")
	}

	// Not connected
	transportMock.On("Send", in).Return(out, nil).Once()
	transportMock.On("IsConnected").Return(false).Once()

	src = decodeHexString("AABBCCDDEEFF")
	_, err = w.Send(src)
	if err == nil {
		t.Errorf("Error expected")
	}

	// Too long
	transportMock.On("Send", in).Return(out, nil).Once()
	transportMock.On("IsConnected").Return(true).Once()

	src = make([]byte, 3000)
	_, err = w.Send(src)
	if err == nil {
		t.Errorf("Error expected")
	}
}

func TestWrapper_InvalidReply(t *testing.T) {
	transportMock := new(mocks.TransportMock)
	transportMock.On("IsConnected").Return(true)

	w := wrapper.New(transportMock, 1, 3)

	// Invalid version
	in := decodeHexString("0001000100030006AABBCCDDEEFF")
	out := decodeHexString("00020003000100050123456789")
	transportMock.On("Send", in).Return(out, nil).Once()

	src := decodeHexString("AABBCCDDEEFF")
	_, err := w.Send(src)
	if err == nil {
		t.Errorf("Error expected")
	}

	// Invalid destination
	out = decodeHexString("00010001000100050123456789")
	transportMock.On("Send", in).Return(out, nil).Once()

	_, err = w.Send(src)
	if err == nil {
		t.Errorf("Error expected")
	}

	// Invalid source
	out = decodeHexString("00010003000300050123456789")
	transportMock.On("Send", in).Return(out, nil).Once()

	_, err = w.Send(src)
	if err == nil {
		t.Errorf("Error expected")
	}

	// Too short
	out = decodeHexString("0001")
	transportMock.On("Send", in).Return(out, nil).Once()

	_, err = w.Send(src)
	if err == nil {
		t.Errorf("Error expected")
	}

	// Too long
	out = decodeHexString("00010003000110000123456789")
	transportMock.On("Send", in).Return(out, nil).Once()

	_, err = w.Send(src)
	if err == nil {
		t.Errorf("Error expected")
	}

	// Length mismatch
	out = decodeHexString("00010003000100040123456789")
	transportMock.On("Send", in).Return(out, nil).Once()

	_, err = w.Send(src)
	if err == nil {
		t.Errorf("Error expected")
	}
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
