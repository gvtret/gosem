package dlmsclient_test

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/Circutor/gosem/pkg/dlms"
	"github.com/Circutor/gosem/pkg/dlms/mocks"
	"github.com/Circutor/gosem/pkg/dlmsclient"
	"github.com/stretchr/testify/assert"
)

func TestClient_Connect(t *testing.T) {
	tm := new(mocks.TransportMock)
	tm.On("Connect").Return(nil).Once()

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 0)

	err := c.Connect()
	assert.NoError(t, err)

	tm.On("IsConnected").Return(true).Once()
	assert.True(t, c.IsConnected())

	tm.AssertExpectations(t)
}

func TestClient_ConnectFail(t *testing.T) {
	tm := new(mocks.TransportMock)
	tm.On("Connect").Return(fmt.Errorf("error connecting"))

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 0)

	err := c.Connect()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorCommunicationFailed, clientError.Code())
}

func TestClient_Disconnect(t *testing.T) {
	tm := new(mocks.TransportMock)
	tm.On("Connect").Return(nil).Once()
	tm.On("Disconnect").Return(fmt.Errorf("error disconnecting")).Once()

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 0)

	err := c.Disconnect()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorCommunicationFailed, clientError.Code())

	c.Connect()

	tm.On("Disconnect").Return(nil).Once()
	err = c.Disconnect()
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
	tm.On("IsConnected").Return(true).Times(2)

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 0)

	c.Connect()

	err := c.Associate()
	assert.NoError(t, err)
	assert.True(t, c.IsAssociated())

	c.Disconnect()

	tm.On("IsConnected").Return(false).Once()
	assert.False(t, c.IsAssociated())

	tm.AssertExpectations(t)
}

func TestClient_InvalidPassword(t *testing.T) {
	in := decodeHexString("6036A1090607608574050801018A0207808B0760857405080201AC0A80083030303030303031BE10040E01000000065F1F040000181F0100")
	out := decodeHexString("6129A109060760857405080101A203020101A305A10302010DBE10040E0800065F1F040000101400800007")

	tm := new(mocks.TransportMock)
	tm.On("Connect").Return(nil).Once()
	tm.On("Send", in).Return(out, nil).Once()
	tm.On("IsConnected").Return(true).Once()

	settings, _ := dlms.NewSettingsWithLowAuthentication([]byte("00000001"))
	c := dlmsclient.New(settings, tm, 0)

	c.Connect()

	err := c.Associate()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidPassword, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_InvalidPasswordLGZ(t *testing.T) {
	in := decodeHexString("6036A1090607608574050801018A0207808B0760857405080201AC0A80083030303030303032BE10040E01000000065F1F040000181F0100")
	out := decodeHexString("6117A109060760857405080101A203020101A305A10302010D")

	tm := new(mocks.TransportMock)
	tm.On("Connect").Return(nil).Once()
	tm.On("Send", in).Return(out, nil).Once()
	tm.On("IsConnected").Return(true).Once()

	settings, _ := dlms.NewSettingsWithLowAuthentication([]byte("00000002"))
	c := dlmsclient.New(settings, tm, 0)

	c.Connect()

	err := c.Associate()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidPassword, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_CloseAssociation(t *testing.T) {
	in := decodeHexString("601DA109060760857405080101BE10040E01000000065F1F040000181F0100")
	out := decodeHexString("6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")

	tm := new(mocks.TransportMock)
	tm.On("Connect").Return(nil).Once()
	tm.On("Send", in).Return(out, nil).Once()
	tm.On("IsConnected").Return(true).Times(2)

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 0)

	c.Connect()

	err := c.Associate()
	assert.NoError(t, err)

	in = decodeHexString("6200")
	out = decodeHexString("6300")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.CloseAssociation()
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_Timeout(t *testing.T) {
	tm := new(mocks.TransportMock)

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 100*time.Millisecond)

	// Check connection is closed after timeout.
	tm.On("Connect").Return(nil).Once()
	err := c.Connect()
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	tm.On("Disconnect").Return(nil).Once()
	time.Sleep(60 * time.Millisecond)

	// Check connection isn't closed by timeout if already is closed.
	tm.On("Connect").Return(nil).Once()
	err = c.Connect()
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	tm.On("Disconnect").Return(nil).Once()
	c.Disconnect()

	time.Sleep(60 * time.Millisecond)

	tm.AssertExpectations(t)
}

func TestClient_TimeoutRefreshWithCommunications(t *testing.T) {
	tm := new(mocks.TransportMock)

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 100*time.Millisecond)

	tm.On("Connect").Return(nil).Once()
	err := c.Connect()
	assert.NoError(t, err)

	in := decodeHexString("601DA109060760857405080101BE10040E01000000065F1F040000181F0100")
	out := decodeHexString("6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")
	tm.On("Send", in).Return(out, nil).Once()
	tm.On("IsConnected").Return(true).Once()
	err = c.Associate()
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	in = decodeHexString("C001C100080000010000FF0300")
	out = decodeHexString("C401C10010003C")
	tm.On("Send", in).Return(out, nil).Once()
	err = c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3), nil)
	assert.NoError(t, err)

	time.Sleep(80 * time.Millisecond)
	tm.On("Disconnect").Return(nil).Once()
	time.Sleep(40 * time.Millisecond)

	tm.AssertExpectations(t)
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
