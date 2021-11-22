package wrapper_test

import (
	"bytes"
	"encoding/hex"
	"gosem/pkg/dlms/mocks"
	"gosem/pkg/wrapper"
	"testing"
)

func TestWrapper_Connect(t *testing.T) {
	transportMock := new(mocks.TransportMock)
	transportMock.On("Connect").Return(nil)

	w, err := wrapper.New(transportMock, 1, 3)
	if err != nil {
		t.Errorf("Error creating wrapper: %s", err)
	}

	err = w.Connect()
	if err != nil {
		t.Errorf("Error connecting: %s", err)
	}

	transportMock.AssertNumberOfCalls(t, "Connect", 1)

	if !w.IsConnected() {
		t.Errorf("Wrapper is not connected")
	}

	err = w.Connect()
	if err == nil {
		t.Errorf("Wrapper should be already connected")
	}
}

func TestWrapper_Send(t *testing.T) {
	in := decodeHexString("0001000100030006AABBCCDDEEFF")
	out := decodeHexString("00010003000100050123456789")

	transportMock := new(mocks.TransportMock)
	transportMock.On("Connect").Return(nil)
	transportMock.On("Send", in).Return(out, nil)

	w, _ := wrapper.New(transportMock, 1, 3)
	w.Connect()

	src := decodeHexString("AABBCCDDEEFF")
	out, err := w.Send(src)
	if err != nil {
		t.Errorf("Error sending: %s", err)
	}

	if !bytes.Equal(out, decodeHexString("0123456789")) {
		t.Errorf("Output is not as expected")
	}

	transportMock.AssertNumberOfCalls(t, "Connect", 1)
	transportMock.AssertNumberOfCalls(t, "Send", 1)
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
