package hdlc_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/circutor-library/gosem/pkg/dlms/mocks"
	"gitlab.com/circutor-library/gosem/pkg/hdlc"
)

func TestHDLC_Connect(t *testing.T) {
	transportMock := mocks.NewTransportMock(t)

	transportMock.On("SetReception", mock.Anything).Once()
	w := hdlc.New(transportMock, 16, 2, 1)

	in := decodeHexString("7EA0080221059356957E")
	transportMock.On("Send", in).Return(nil).Once()

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

	// transportMock.On("SetReception", mock.Anything).Once()
	w := hdlc.New(transportMock, 16, 2, 1)

	transportMock.On("Connect").Return(assert.AnError).Once()
	assert.Error(t, w.Connect())

	transportMock.On("IsConnected").Return(false).Once()
	assert.False(t, w.IsConnected())

	transportMock.On("Close").Return(nil).Once()
	w.Close()

	transportMock.AssertExpectations(t)
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
