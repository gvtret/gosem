package dlmsclient_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/circutor-library/gosem/pkg/axdr"
	"gitlab.com/circutor-library/gosem/pkg/dlms"
)

func TestClient_ActionRequest(t *testing.T) {
	c, tm, rdc := associate(t)

	var data int8

	sendReceive(tm, rdc, "C301C10046000060030AFF01010F00", "C701C10000")
	err := c.ActionRequest(dlms.CreateMethodDescriptor(70, "0-0:96.3.10.255", 1), data)
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_ActionRequestFail(t *testing.T) {
	c, tm, rdc := associate(t)

	data := axdr.CreateAxdrInteger(0)
	disconnectorMethodDescriptor := dlms.CreateMethodDescriptor(70, "0-0:96.3.10.255", 1)

	// Action failed
	sendReceive(tm, rdc, "C301C10046000060030AFF01010F00", "C701C1010102")
	err := c.ActionRequest(disconnectorMethodDescriptor, data)
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorActionRejected, clientError.Code())

	// Unexpected response
	sendReceive(tm, rdc, "C301C10046000060030AFF01010F00", "0E010203")
	err = c.ActionRequest(disconnectorMethodDescriptor, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Invalid response
	sendReceive(tm, rdc, "C301C10046000060030AFF01010F00", "AE12")
	err = c.ActionRequest(disconnectorMethodDescriptor, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Send failed
	tm.On("Send", decodeHexString("C301C10046000060030AFF01010F00")).Return(fmt.Errorf("error")).Once()
	tm.On("IsConnected").Return(false).Once()

	err = c.ActionRequest(disconnectorMethodDescriptor, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorCommunicationFailed, clientError.Code())

	// Not associated
	tm.On("Disconnect").Return(nil).Once()
	c.Disconnect()

	err = c.ActionRequest(disconnectorMethodDescriptor, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidState, clientError.Code())

	// Invalid data
	err = c.ActionRequest(disconnectorMethodDescriptor, nil)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidParameter, clientError.Code())

	// nil attribute descriptor
	err = c.ActionRequest(nil, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidParameter, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_ActionBroadcast(t *testing.T) {
	c, tm, _ := associate(t)

	settings := c.GetSettings()
	settings.UseBroadcast = true
	c.SetSettings(settings)

	var data int8

	tm.On("Send", decodeHexString("C301870046000060030AFF01010F00")).Return(nil).Once()

	err := c.ActionRequest(dlms.CreateMethodDescriptor(70, "0-0:96.3.10.255", 1), data)
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}
