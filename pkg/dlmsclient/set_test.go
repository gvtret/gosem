package dlmsclient_test

import (
	"fmt"
	"testing"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
	"github.com/stretchr/testify/assert"
)

func TestClient_SetRequest(t *testing.T) {
	c, tm, rdc := associate(t)

	var data uint32 = 10000

	sendReceive(tm, rdc, "C101C1000300015E230BFF02000600002710", "C501C100")
	err := c.SetRequest(dlms.CreateAttributeDescriptor(3, "0-1:94.35.11.255", 2), data)
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_SetRequestFail(t *testing.T) {
	c, tm, rdc := associate(t)

	data := axdr.CreateAxdrDoubleLongUnsigned(10000)
	demandAttributeDescriptor := dlms.CreateAttributeDescriptor(3, "0-1:94.35.11.255", 2)

	// Set failed
	sendReceive(tm, rdc, "C101C1000300015E230BFF02000600002710", "C501C10102")
	err := c.SetRequest(demandAttributeDescriptor, data)
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorSetRejected, clientError.Code())

	// Unexpected response
	sendReceive(tm, rdc, "C101C1000300015E230BFF02000600002710", "0E010203")
	err = c.SetRequest(demandAttributeDescriptor, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Invalid response
	sendReceive(tm, rdc, "C101C1000300015E230BFF02000600002710", "AE12")
	err = c.SetRequest(demandAttributeDescriptor, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Send failed
	tm.On("Send", decodeHexString("C101C1000300015E230BFF02000600002710")).Return(fmt.Errorf("error")).Once()
	tm.On("IsConnected").Return(false).Once()

	err = c.SetRequest(demandAttributeDescriptor, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorCommunicationFailed, clientError.Code())

	// Not associated
	tm.On("Disconnect").Return(nil).Once()
	c.Disconnect()

	err = c.SetRequest(demandAttributeDescriptor, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidState, clientError.Code())

	// Invalid data
	err = c.SetRequest(demandAttributeDescriptor, nil)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidParameter, clientError.Code())

	// nil attribute descriptor
	err = c.SetRequest(nil, data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidParameter, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_SetRequestWithStructOfElements(t *testing.T) {
	var value1 uint32 = 10000

	data := struct {
		Value1 *uint32 `obis:"3,0-1:94.35.11.255,2"`
		Value2 uint16  `obis:"1,1-1:94.34.104.255,2"`
		Value3 *int32  `obis:"70,0-0:96.3.10.255,3"`
	}{
		Value1: &value1,
		Value2: 12345,
		Value3: nil, // nil fields are ignored
	}

	c, tm, rdc := associate(t)

	var v interface{} = &data

	sendReceive(tm, rdc, "C101C1000300015E230BFF02000600002710", "C501C100")
	sendReceive(tm, rdc, "C101C1000101015E2268FF0200123039", "C501C100")
	err := c.SetRequestWithStructOfElements(&v)
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_SetRequestWithStructOfElementsWithFail(t *testing.T) {
	data := struct {
		ToSkip int
		Value1 uint16 `obis:"3,0-1:94.35.11.255,2"`
		Value2 uint16 `obis:"1,1-1:94.34.104.255,2"`
	}{
		ToSkip: 0,
		Value1: 6789,
		Value2: 12345,
	}

	c, tm, rdc := associate(t)

	// If second element fails, then we expect an ErrorSetPartial

	var v interface{} = &data

	sendReceive(tm, rdc, "C101C1000300015E230BFF0200121A85", "C501C100")
	sendReceive(tm, rdc, "C101C1000101015E2268FF0200123039", "C501C103")
	err := c.SetRequestWithStructOfElements(&v)
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorSetPartial, clientError.Code())

	// If first element fails, then we expect an ErrorSetRejected

	sendReceive(tm, rdc, "C101C1000300015E230BFF0200121A85", "C501C103")
	err = c.SetRequestWithStructOfElements(&v)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorSetRejected, clientError.Code())

	tm.AssertExpectations(t)
}
