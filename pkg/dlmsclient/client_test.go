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
	"github.com/stretchr/testify/mock"
)

func TestClient_Connect(t *testing.T) {
	tm := mocks.NewTransportMock(t)
	tm.On("Connect").Return(nil).Once()

	tm.On("SetReception", mock.Anything).Once()
	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	err := c.Connect()
	assert.NoError(t, err)

	tm.On("IsConnected").Return(true).Once()
	assert.True(t, c.IsConnected())

	tm.AssertExpectations(t)
}

func TestClient_ConnectFail(t *testing.T) {
	tm := mocks.NewTransportMock(t)
	tm.On("Connect").Return(fmt.Errorf("error connecting"))

	tm.On("SetReception", mock.Anything).Once()
	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	err := c.Connect()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorCommunicationFailed, clientError.Code())
}

func TestClient_Disconnect(t *testing.T) {
	tm := mocks.NewTransportMock(t)
	tm.On("Connect").Return(nil).Once()
	tm.On("Disconnect").Return(fmt.Errorf("error disconnecting")).Once()

	tm.On("SetReception", mock.Anything).Once()
	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

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
	c, tm, _ := associate(t)

	tm.On("IsConnected").Return(true).Once()
	assert.True(t, c.IsAssociated())

	tm.On("Disconnect").Return(nil).Once()
	c.Disconnect()

	tm.On("IsConnected").Return(false).Once()
	assert.False(t, c.IsAssociated())

	tm.AssertExpectations(t)
}

func TestClient_InvalidPassword(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 1)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	settings, _ := dlms.NewSettingsWithLowAuthentication([]byte("00000001"))
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	c.Connect()

	tm.On("IsConnected").Return(true).Once()
	sendReceive(tm, rdc, "6036A1090607608574050801018A0207808B0760857405080201AC0A80083030303030303031BE10040E01000000065F1F040000181F0100", "6129A109060760857405080101A203020101A305A10302010DBE10040E0800065F1F040000101400800007")

	err := c.Associate()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidPassword, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_InvalidPasswordLGZ(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 1)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	settings, _ := dlms.NewSettingsWithLowAuthentication([]byte("00000002"))
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	c.Connect()

	tm.On("IsConnected").Return(true).Once()
	sendReceive(tm, rdc, "6036A1090607608574050801018A0207808B0760857405080201AC0A80083030303030303032BE10040E01000000065F1F040000181F0100", "6117A109060760857405080101A203020101A305A10302010D")

	err := c.Associate()
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidPassword, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_CloseAssociation(t *testing.T) {
	c, tm, rdc := associate(t)

	tm.On("IsConnected").Return(true).Once()
	sendReceive(tm, rdc, "6200", "6300")

	err := c.CloseAssociation()
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_Timeout(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	tm.On("SetReception", mock.Anything).Once()
	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 100*time.Millisecond)

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
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 1)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 100*time.Millisecond)

	tm.On("Connect").Return(nil).Once()
	assert.NoError(t, c.Connect())

	tm.On("IsConnected").Return(true).Once()
	sendReceive(tm, rdc, "601DA109060760857405080101BE10040E01000000065F1F040000181F0100", "6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")
	assert.NoError(t, c.Associate())

	time.Sleep(50 * time.Millisecond)

	sendReceive(tm, rdc, "C001C100080000010000FF0300", "C401C10010003C")
	err := c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3), nil)
	assert.NoError(t, err)

	time.Sleep(80 * time.Millisecond)
	tm.On("Disconnect").Return(nil).Once()
	time.Sleep(40 * time.Millisecond)

	tm.AssertExpectations(t)
}

func TestClient_GetAndSetSettings(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	tm.On("SetReception", mock.Anything).Once()
	settings, _ := dlms.NewSettingsWithLowAuthentication([]byte("00000002"))
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	settings = c.GetSettings()
	assert.Equal(t, []byte("00000002"), settings.Password)

	settings.Password = []byte("00000003")
	c.SetSettings(settings)

	settings = c.GetSettings()
	assert.Equal(t, []byte("00000003"), settings.Password)
}

func TestClient_DataNotification(t *testing.T) {
	c, tm, rdc := associate(t)

	dataNotification := make(chan dlms.DataNotification, 10)
	c.SetDataNotificationChannel(dataNotification)

	rdc <- decodeHexString("0F0063D76A0003FF")

	dn := <-dataNotification
	assert.Equal(t, uint32(6543210), dn.InvokeIDAndPriority)

	tm.AssertExpectations(t)
}

func sendReceive(tm *mocks.TransportMock, rdc dlms.DataChannel, in string, out string) {
	tm.On("Send", decodeHexString(in)).Run(func(args mock.Arguments) {
		if rdc != nil {
			rdc <- decodeHexString(out)
		}
	}).Return(nil).Once()
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
