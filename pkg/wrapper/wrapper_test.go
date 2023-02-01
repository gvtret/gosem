package wrapper_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/Circutor/gosem/pkg/dlms"
	"github.com/Circutor/gosem/pkg/dlms/mocks"
	"github.com/Circutor/gosem/pkg/wrapper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var errFoo = fmt.Errorf("foo")

func TestWrapper_Connect(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	transportMock.On("SetReception", mock.Anything).Once()
	w := wrapper.New(transportMock, 1, 3)

	transportMock.On("Connect").Return(nil).Once()
	assert.NoError(t, w.Connect())

	transportMock.On("IsConnected").Return(true).Once()
	assert.True(t, w.IsConnected())

	transportMock.On("Disconnect").Return(nil).Once()
	assert.NoError(t, w.Disconnect())

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func TestWrapper_ConnectFail(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	transportMock.On("SetReception", mock.Anything).Once()
	w := wrapper.New(transportMock, 1, 3)

	transportMock.On("Connect").Return(errFoo).Once()
	assert.Error(t, w.Connect())

	transportMock.On("IsConnected").Return(false).Once()
	assert.False(t, w.IsConnected())

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func TestWrapper_Send(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	transportMock.On("SetReception", mock.Anything).Once()
	w := wrapper.New(transportMock, 1, 3)

	transportMock.On("Connect").Return(nil).Once()
	w.Connect()

	transportMock.On("IsConnected").Return(true).Once()

	in := decodeHexString("0001000100030006AABBCCDDEEFF")
	transportMock.On("Send", in).Return(nil).Once()

	src := decodeHexString("AABBCCDDEEFF")
	assert.NoError(t, w.Send(src))

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func TestWrapper_SendFailed(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	transportMock.On("SetReception", mock.Anything).Once()
	w := wrapper.New(transportMock, 1, 3)

	// Send failed
	in := decodeHexString("0001000100030006AABBCCDDEEFF")
	transportMock.On("Send", in).Return(errFoo).Once()
	transportMock.On("IsConnected").Return(true).Once()

	src := decodeHexString("AABBCCDDEEFF")
	assert.Error(t, w.Send(src))

	transportMock.AssertExpectations(t)

	// Not connected
	transportMock.On("IsConnected").Return(false).Once()

	src = decodeHexString("AABBCCDDEEFF")
	assert.Error(t, w.Send(src))

	transportMock.AssertExpectations(t)

	// Too long
	transportMock.On("IsConnected").Return(true).Once()

	src = make([]byte, 3000)
	assert.Error(t, w.Send(src))

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func TestWrapper_ChangeAddress(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	transportMock.On("SetReception", mock.Anything).Once()
	w := wrapper.New(transportMock, 1, 3)

	transportMock.On("Connect").Return(nil).Once()
	w.Connect()

	w.SetAddress(2, 4)

	transportMock.On("IsConnected").Return(true).Once()

	in := decodeHexString("0001000200040006AABBCCDDEEFF")
	transportMock.On("Send", in).Return(nil).Once()

	src := decodeHexString("AABBCCDDEEFF")
	assert.NoError(t, w.Send(src))

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func TestWrapper_Receive(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	var tdc dlms.DataChannel
	wdc := make(dlms.DataChannel, 1)

	transportMock.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		tdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	w := wrapper.New(transportMock, 1, 3)
	w.SetReception(wdc)

	// Invalid version
	tdc <- decodeHexString("00020003000100050123456789")

	// Invalid destination
	tdc <- decodeHexString("00010001000100050123456789")

	// Invalid source
	tdc <- decodeHexString("00010003000300050123456789")

	// Too short
	tdc <- decodeHexString("0001")

	// Too long
	tdc <- decodeHexString("00010003000110000123456789")

	// Length mismatch
	tdc <- decodeHexString("00010003000100040123456789")

	// Valid
	tdc <- decodeHexString("00010003000100050123456789")
	assert.Equal(t, decodeHexString("0123456789"), <-wdc)

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
