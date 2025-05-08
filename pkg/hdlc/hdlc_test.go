package hdlc_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/circutor-library/gosem/pkg/dlms"
	"gitlab.com/circutor-library/gosem/pkg/dlms/mocks"
	"gitlab.com/circutor-library/gosem/pkg/hdlc"
)

func TestHDLC_Connect(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	rdc := make(dlms.DataChannel, 10)
	transportMock.On("SetReception", mock.Anything).Run(func(args mock.Arguments) {
		rdc = args.Get(0).(dlms.DataChannel)
	}).Once()

	w := hdlc.New(transportMock, 16, 73, 1)

	sendReceive(transportMock, rdc, "7EA00802219393DBD87E", "7EA01F93022173BCAC8180120501F80601F00704000000010804000000013D9B7E")

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

func TestHDLC_ConnectFail(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	transportMock.On("SetReception", mock.Anything).Once()
	w := hdlc.New(transportMock, 16, 73, 1)

	transportMock.On("Connect").Return(assert.AnError).Once()
	assert.Error(t, w.Connect())

	transportMock.On("IsConnected").Return(false).Once()
	assert.False(t, w.IsConnected())

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func sendReceive(tm *mocks.TransportMock, rdc dlms.DataChannel, in string, out string) {
	tm.On("Send", decodeHexString(in)).Run(func(_ mock.Arguments) {
		if rdc != nil {
			rdc <- decodeHexString(out)
		}
	}).Return(nil).Once()
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
