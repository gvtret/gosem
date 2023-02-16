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

func TestClient_CompleteCommunication(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 1)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	assert.NoError(t, c.Connect())

	tm.On("IsConnected").Return(true)
	sendReceive(tm, rdc, "601DA109060760857405080101BE10040E01000000065F1F040000181F0100", "6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")
	assert.NoError(t, c.Associate())

	sendReceive(tm, rdc, "C001C100080000010000FF0300", "C401C10010003C")
	err := c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3), nil)
	assert.NoError(t, err)

	sendReceive(tm, rdc, "6200", "6300")
	err = c.CloseAssociation()
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_CompleteSecureCommunication(t *testing.T) {
	tm := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 1)
	tm.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	ciphering, _ := dlms.NewCiphering(
		dlms.SecurityLevelDedicatedKey,
		dlms.SecurityEncryption|dlms.SecurityAuthentication,
		decodeHexString("4349520000000001"),
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
		0x00000059,
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
	)
	ciphering.DedicatedKey = decodeHexString("5E168412318BA71848C99B2B2AB33294")

	settings, _ := dlms.NewSettingsWithLowAuthenticationAndCiphering([]byte("JuS66BCZ"), ciphering)
	settings.MaxPduSize = 512

	c := dlmsclient.New(settings, tm, 5*time.Second, 0)

	tm.On("Connect").Return(nil).Once()
	assert.NoError(t, c.Connect())

	tm.On("IsConnected").Return(true)
	sendReceive(tm, rdc, "6066A109060760857405080103A60A040843495200000000018A0207808B0760857405080201AC0A80084A7553363642435ABE3404322130300000005992D807DBCF8533E9AD675AE0948241FB8E6CF9AFA7006BAA134A473C9151B3362F56DC12F89E85DA97E176",
		"6148A109060760857405080103A203020100A305A103020100A40A04084C475A2022604828BE230421281F300000005AE916783AF33B5317AD0E453A799A65F26AE97660CF8B14FEB7B0")
	assert.NoError(t, c.Associate())

	sendReceive(tm, rdc, "D01E3000000001D3B903996D9508C5B6BCDEB025DD1800A5C92775FB55F317CF", "D4233000000001AA07A549F82E6B8EEA919659D91689BF995BE6F93C95A7208718A3B84EE4")
	err := c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 2), nil)
	assert.NoError(t, err)

	sendReceive(tm, rdc, "6239800100BE3404322130300000005A8E9B83D641B89FAAF36DA504132C34F87E4BA66175A7DCED015460239699C72C18C06DB29C54673B83BAC0", "6328800100BE230421281F300000005BCD34827974EDCF8B1DAB306F62C58AB42052DB67361377507825")
	assert.NoError(t, c.CloseAssociation())

	assert.Equal(t, uint32(0x0000005B), c.GetSettings().Ciphering.UnicastKeyIC)
	assert.Equal(t, uint32(0x00000002), c.GetSettings().Ciphering.DedicatedKeyIC)

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
