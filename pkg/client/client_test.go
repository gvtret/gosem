package client_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/Circutor/gosem/pkg/client"
	"github.com/Circutor/gosem/pkg/dlms"
	"github.com/Circutor/gosem/pkg/dlms/mocks"
	"github.com/stretchr/testify/assert"
)

func TestClient_Connect(t *testing.T) {
	tm := new(mocks.TransportMock)
	tm.On("Connect").Return(nil).Once()

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	client := client.New(settings, tm)

	err := client.Connect()
	assert.NoError(t, err)

	tm.On("IsConnected").Return(true).Once()
	assert.True(t, client.IsConnected())

	tm.AssertExpectations(t)
}

func TestClient_ConnectFail(t *testing.T) {
	tm := new(mocks.TransportMock)
	tm.On("Connect").Return(fmt.Errorf("error connecting"))

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	client := client.New(settings, tm)

	err := client.Connect()
	assert.Error(t, err)
}

func TestClient_Disconnect(t *testing.T) {
	tm := new(mocks.TransportMock)
	tm.On("Connect").Return(nil).Once()
	tm.On("Disconnect").Return(fmt.Errorf("error disconnecting")).Once()

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	client := client.New(settings, tm)

	err := client.Disconnect()
	assert.Error(t, err)

	client.Connect()

	tm.On("Disconnect").Return(nil).Once()
	err = client.Disconnect()
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_Associate(t *testing.T) {
	in := decodeHexString("601DA109060760857405080101BE10040E01000000065F1F040000181F0100")
	out := decodeHexString("6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")

	tm := new(mocks.TransportMock)
	tm.On("Connect").Return(nil).Once()
	tm.On("Disconnect").Return(nil).Once()
	tm.On("Send", in).Return(out, nil).Once()
	tm.On("IsConnected").Return(true).Times(3)

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	client := client.New(settings, tm)

	client.Connect()

	err := client.Associate()
	assert.NoError(t, err)
	assert.True(t, client.IsAssociated())

	client.Disconnect()

	tm.On("IsConnected").Return(false).Once()
	assert.False(t, client.IsAssociated())

	tm.AssertExpectations(t)
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
