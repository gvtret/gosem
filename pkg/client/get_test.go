package client_test

import (
	"fmt"
	"gosem/pkg/axdr"
	"gosem/pkg/client"
	"gosem/pkg/dlms"
	"gosem/pkg/dlms/mocks"
	"testing"
)

func TestClient_Get(t *testing.T) {
	c, tm, err := associate()
	if err != nil {
		t.Error(err)
	}

	in := decodeHexString("C001C100080000010000FF0300")
	out := decodeHexString("C401C10010003C")
	tm.On("Send", in).Return(out, nil).Once()

	a, err := c.Get(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3))
	if err != nil {
		t.Error(err)
	}

	b := axdr.CreateAxdrLong(0x003C)
	if a.Tag != b.Tag {
		t.Errorf("Expected %v, got %v", b.Tag, a.Tag)
	}
	if a.Value != b.Value {
		t.Errorf("Expected %v, got %v", b.Value, a.Value)
	}

	tm.AssertExpectations(t)
}

func TestClient_GetFail(t *testing.T) {
	c, tm, err := associate()
	if err != nil {
		t.Error(err)
	}

	// Get failed
	in := decodeHexString("C001C100080000010000FF0300")
	out := decodeHexString("C401C10102")
	tm.On("Send", in).Return(out, nil).Once()

	_, err = c.Get(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3))
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Unexpected response
	in = decodeHexString("C001C100080000010000FF0300")
	out = decodeHexString("0E010203")
	tm.On("Send", in).Return(out, nil).Once()

	_, err = c.Get(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3))
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Invalid response
	in = decodeHexString("C001C100080000010000FF0300")
	out = decodeHexString("AE12")
	tm.On("Send", in).Return(out, nil).Once()

	_, err = c.Get(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3))
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Send failed
	in = decodeHexString("C001C100080000010000FF0300")
	out = decodeHexString("C401C10102")
	tm.On("Send", in).Return(out, fmt.Errorf("error")).Once()

	_, err = c.Get(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3))
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Not associated
	tm.On("Disconnect").Return(nil).Once()
	c.Disconnect()

	_, err = c.Get(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3))
	if err == nil {
		t.Error("Expected error, got nil")
	}

	tm.AssertExpectations(t)
}

func TestClient_GetWithUnmarshal(t *testing.T) {
	c, tm, err := associate()
	if err != nil {
		t.Error(err)
	}

	in := decodeHexString("C001C100080000010000FF0300")
	out := decodeHexString("C401C10010003C")
	tm.On("Send", in).Return(out, nil).Once()

	var data int16

	err = c.GetWithUnmarshal(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3), &data)
	if err != nil {
		t.Error(err)
	}

	if data != 0x003C {
		t.Errorf("Expected %v, got %v", 0x003C, data)
	}

	tm.AssertExpectations(t)
}

func TestClient_GetWithUnmarshalFail(t *testing.T) {
	c, tm, err := associate()
	if err != nil {
		t.Error(err)
	}

	// Get failed
	in := decodeHexString("C001C100080000010000FF0300")
	out := decodeHexString("C401C10102")
	tm.On("Send", in).Return(out, nil).Once()

	var data int32

	err = c.GetWithUnmarshal(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3), &data)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Unexpected response
	in = decodeHexString("C001C100080000010000FF0300")
	out = decodeHexString("C401C10010003C")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.GetWithUnmarshal(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3), &data)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func associate() (*client.Client, *mocks.TransportMock, error) {
	in := decodeHexString("601DA109060760857405080101BE10040E01000000065F1F040000181F0100")
	out := decodeHexString("6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")

	transportMock := new(mocks.TransportMock)
	transportMock.On("Connect").Return(nil).Once()
	transportMock.On("Send", in).Return(out, nil).Once()
	transportMock.On("IsConnected").Return(true).Twice()

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	client := client.New(settings, transportMock)

	client.Connect()

	err := client.Associate()
	if err != nil {
		err = fmt.Errorf("Associate failed: %w", err)
		return nil, nil, err
	}

	return client, transportMock, nil
}
