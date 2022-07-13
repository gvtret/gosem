package dlmsclient_test

import (
	"fmt"
	"testing"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
	"github.com/Circutor/gosem/pkg/dlmsclient"
	"github.com/stretchr/testify/assert"
)

func TestClient_ActionRequest(t *testing.T) {
	c, tm, err := associate()
	assert.NoError(t, err)

	in := decodeHexString("C301C10046000060030AFF01010F00")
	out := decodeHexString("C701C10000")
	tm.On("Send", in).Return(out, nil).Once()

	var data int8

	err = c.ActionRequest(dlms.CreateMethodDescriptor(70, "0-0:96.3.10.255", 1), data)
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_ActionRequestFail(t *testing.T) {
	c, tm, err := associate()
	assert.NoError(t, err)

	data := axdr.CreateAxdrInteger(0)
	disconnectorMethodDescriptor := dlms.CreateMethodDescriptor(70, "0-0:96.3.10.255", 1)

	// Action failed
	in := decodeHexString("C301C10046000060030AFF01010F00")
	out := decodeHexString("C701C1010102")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.ActionRequest(disconnectorMethodDescriptor, data)
	var clientError *dlmsclient.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlmsclient.ErrorActionRejected, clientError.Code())

	// Unexpected response
	in = decodeHexString("C301C10046000060030AFF01010F00")
	out = decodeHexString("0E010203")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.ActionRequest(disconnectorMethodDescriptor, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlmsclient.ErrorInvalidResponse, clientError.Code())

	// Invalid response
	in = decodeHexString("C301C10046000060030AFF01010F00")
	out = decodeHexString("AE12")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.ActionRequest(disconnectorMethodDescriptor, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlmsclient.ErrorInvalidResponse, clientError.Code())

	// Send failed
	in = decodeHexString("C301C10046000060030AFF01010F00")
	out = decodeHexString("")
	tm.On("Send", in).Return(out, fmt.Errorf("error")).Once()
	tm.On("IsConnected").Return(false).Once()

	err = c.ActionRequest(disconnectorMethodDescriptor, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlmsclient.ErrorCommunicationFailed, clientError.Code())

	// Not associated
	tm.On("Disconnect").Return(nil).Once()
	c.Disconnect()

	err = c.ActionRequest(disconnectorMethodDescriptor, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlmsclient.ErrorInvalidState, clientError.Code())

	// Invalid data
	err = c.ActionRequest(disconnectorMethodDescriptor, nil)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlmsclient.ErrorInvalidParameter, clientError.Code())

	// nil attribute descriptor
	err = c.ActionRequest(nil, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlmsclient.ErrorInvalidParameter, clientError.Code())

	tm.AssertExpectations(t)
}
