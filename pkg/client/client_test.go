package client_test

import (
	"encoding/hex"
	"fmt"
	"gosem/pkg/client"
	"gosem/pkg/dlms"
	"gosem/pkg/dlms/mocks"
	"testing"
)

func TestClient_Connect(t *testing.T) {
	transportMock := new(mocks.TransportMock)
	transportMock.On("Connect").Return(nil)

	settings, _ := dlms.NewSettingsWithoutAuthentication()

	client, err := client.New(settings, transportMock)
	if err != nil {
		t.Errorf("Error creating client: %s", err)
	}

	err = client.Connect()
	if err != nil {
		t.Errorf("Error connecting: %s", err)
	}

	transportMock.AssertNumberOfCalls(t, "Connect", 1)

	if !client.IsConnected() {
		t.Errorf("Client is not connected")
	}

	err = client.Connect()
	if err == nil {
		t.Errorf("Client should be already connected")
	}
}

func TestClient_ConnectFail(t *testing.T) {
	transportMock := new(mocks.TransportMock)
	transportMock.On("Connect").Return(fmt.Errorf("error connecting"))

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	client, _ := client.New(settings, transportMock)

	err := client.Connect()
	if err == nil {
		t.Errorf("Error connecting should not be nil")
	}

	if client.IsConnected() {
		t.Errorf("Client is connected")
	}
}

func TestClient_Disconnect(t *testing.T) {
	transportMock := new(mocks.TransportMock)
	transportMock.On("Connect").Return(nil)
	transportMock.On("Disconnect").Return(nil)

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	client, _ := client.New(settings, transportMock)

	err := client.Disconnect()
	if err == nil {
		t.Errorf("Error disconnecting should not be nil")
	}

	client.Connect()

	err = client.Disconnect()
	if err != nil {
		t.Errorf("Error disconnecting: %s", err)
	}

	transportMock.AssertNumberOfCalls(t, "Connect", 1)
	transportMock.AssertNumberOfCalls(t, "Disconnect", 1)

	if client.IsConnected() {
		t.Errorf("Client is connected")
	}
}

func TestClient_DisconnectFail(t *testing.T) {
	transportMock := new(mocks.TransportMock)
	transportMock.On("Connect").Return(nil)
	transportMock.On("Disconnect").Return(fmt.Errorf("error disconnecting"))

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	client, _ := client.New(settings, transportMock)

	client.Connect()

	err := client.Disconnect()
	if err == nil {
		t.Errorf("Error disconnecting should not be nil")
	}

	transportMock.AssertNumberOfCalls(t, "Connect", 1)
	transportMock.AssertNumberOfCalls(t, "Disconnect", 1)
}

func TestClient_Associate(t *testing.T) {
	in := decodeHexString("601DA109060760857405080101BE10040E01000000065F1F040000181F0100")
	out := decodeHexString("6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")

	transportMock := new(mocks.TransportMock)
	transportMock.On("Connect").Return(nil)
	transportMock.On("Disconnect").Return(nil)
	transportMock.On("Send", in).Return(out, nil)

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	client, _ := client.New(settings, transportMock)

	client.Connect()

	err := client.Associate()
	if err != nil {
		t.Errorf("Error associating: %s", err)
	}

	if !client.IsAssociated() {
		t.Errorf("Client is not associated")
	}

	client.Disconnect()

	if client.IsAssociated() {
		t.Errorf("Client is associated")
	}

	transportMock.AssertNumberOfCalls(t, "Connect", 1)
	transportMock.AssertNumberOfCalls(t, "Disconnect", 1)
	transportMock.AssertNumberOfCalls(t, "Send", 1)
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
