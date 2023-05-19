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
	err := c.SetRequestWithStructOfElements(&v, true)
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
	err := c.SetRequestWithStructOfElements(&v, true)
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorSetPartial, clientError.Code())

	// If first element fails, then we expect an ErrorSetPartial

	sendReceive(tm, rdc, "C101C1000300015E230BFF0200121A85", "C501C103")
	sendReceive(tm, rdc, "C101C1000101015E2268FF0200123039", "C501C100")
	err = c.SetRequestWithStructOfElements(&v, true)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorSetPartial, clientError.Code())

	// If both fails, then we expect an ErrorSetRejected

	sendReceive(tm, rdc, "C101C1000300015E230BFF0200121A85", "C501C103")
	sendReceive(tm, rdc, "C101C1000101015E2268FF0200123039", "C501C103")
	err = c.SetRequestWithStructOfElements(&v, true)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorSetRejected, clientError.Code())

	// If first element fails, don't continue and we expect an ErrorSetRejected

	sendReceive(tm, rdc, "C101C1000300015E230BFF0200121A85", "C501C103")
	err = c.SetRequestWithStructOfElements(&v, false)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorSetRejected, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_SetRequestWithDataBlock(t *testing.T) {
	c, tm, rdc := associate(t)

	data := struct {
		ValueStr1 string
		ValueStr2 string
		ValueStr3 string
		ValueStr4 string
		ValueStr5 string
		Value1    uint64
		Value2    uint64
		Value3    uint64
		Value4    uint64
		Value5    uint64
	}{
		ValueStr1: "00010203040506070809000102030405060708090001020304050607080900010203040506070809",
		ValueStr2: "00010203040506070809000102030405060708090001020304050607080900010203040506070809",
		ValueStr3: "00010203040506070809000102030405060708090001020304050607080900010203040506070809",
		ValueStr4: "00010203040506070809000102030405060708090001020304050607080900010203040506070809",
		ValueStr5: "00010203040506070809000102030405060708090001020304050607080900010203040506070809",
		Value1:    123,
		Value2:    234,
		Value3:    345,
		Value4:    456,
		Value5:    567,
	}

	sendReceive(tm, rdc, "C102C1000300015E230BFF020000000000016B020A092800010203040506070809000102030405060708090001020304050607080900010203040506070809092800010203040506070809000102030405060708090001020304050607080900010203040506070809092800010203040506070809000102030405060708", "C502C100000001")
	sendReceive(tm, rdc, "C103C100000000027509000102030405060708090001020304050607080909280001020304050607080900010203040506070809000102030405060708090001020304050607080909280001020304050607080900010203040506070809000102030405060708090001020304050607080915000000000000007B150000", "C502C100000002")
	sendReceive(tm, rdc, "C103C10100000003210000000000EA1500000000000001591500000000000001C8150000000000000237", "C503C10000000003")

	err := c.SetRequest(dlms.CreateAttributeDescriptor(3, "0-1:94.35.11.255", 2), data)
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_SetRequestWithDataBlockFail(t *testing.T) {
	c, tm, rdc := associate(t)

	data := struct {
		ValueStr1 string
		ValueStr2 string
		ValueStr3 string
		ValueStr4 string
		ValueStr5 string
		Value1    uint64
		Value2    uint64
		Value3    uint64
		Value4    uint64
		Value5    uint64
	}{
		ValueStr1: "00010203040506070809000102030405060708090001020304050607080900010203040506070809",
		ValueStr2: "00010203040506070809000102030405060708090001020304050607080900010203040506070809",
		ValueStr3: "00010203040506070809000102030405060708090001020304050607080900010203040506070809",
		ValueStr4: "00010203040506070809000102030405060708090001020304050607080900010203040506070809",
		ValueStr5: "00010203040506070809000102030405060708090001020304050607080900010203040506070809",
		Value1:    123,
		Value2:    234,
		Value3:    345,
		Value4:    456,
		Value5:    567,
	}

	// If block number doesn't match, then we expect an ErrorInvalidResponse
	sendReceive(tm, rdc, "C102C1000300015E230BFF020000000000016B020A092800010203040506070809000102030405060708090001020304050607080900010203040506070809092800010203040506070809000102030405060708090001020304050607080900010203040506070809092800010203040506070809000102030405060708", "C502C100000002")
	err := c.SetRequest(dlms.CreateAttributeDescriptor(3, "0-1:94.35.11.255", 2), data)
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// If set failed, then we expect an ErrorSetRejected
	sendReceive(tm, rdc, "C102C1000300015E230BFF020000000000016B020A092800010203040506070809000102030405060708090001020304050607080900010203040506070809092800010203040506070809000102030405060708090001020304050607080900010203040506070809092800010203040506070809000102030405060708", "C502C100000001")
	sendReceive(tm, rdc, "C103C100000000027509000102030405060708090001020304050607080909280001020304050607080900010203040506070809000102030405060708090001020304050607080909280001020304050607080900010203040506070809000102030405060708090001020304050607080915000000000000007B150000", "C502C100000002")
	sendReceive(tm, rdc, "C103C10100000003210000000000EA1500000000000001591500000000000001C8150000000000000237", "C503C10200000003")
	err = c.SetRequest(dlms.CreateAttributeDescriptor(3, "0-1:94.35.11.255", 2), data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorSetRejected, clientError.Code())

	// If block number doesn't match in last block, then we expect an ErrorInvalidResponse
	sendReceive(tm, rdc, "C102C1000300015E230BFF020000000000016B020A092800010203040506070809000102030405060708090001020304050607080900010203040506070809092800010203040506070809000102030405060708090001020304050607080900010203040506070809092800010203040506070809000102030405060708", "C502C100000001")
	sendReceive(tm, rdc, "C103C100000000027509000102030405060708090001020304050607080909280001020304050607080900010203040506070809000102030405060708090001020304050607080909280001020304050607080900010203040506070809000102030405060708090001020304050607080915000000000000007B150000", "C502C100000002")
	sendReceive(tm, rdc, "C103C10100000003210000000000EA1500000000000001591500000000000001C8150000000000000237", "C503C10000000004")
	err = c.SetRequest(dlms.CreateAttributeDescriptor(3, "0-1:94.35.11.255", 2), data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// If we receive an unexpected response, then we expect an ErrorInvalidResponse
	sendReceive(tm, rdc, "C102C1000300015E230BFF020000000000016B020A092800010203040506070809000102030405060708090001020304050607080900010203040506070809092800010203040506070809000102030405060708090001020304050607080900010203040506070809092800010203040506070809000102030405060708", "0E010203")
	err = c.SetRequest(dlms.CreateAttributeDescriptor(3, "0-1:94.35.11.255", 2), data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// If we receive an unexpected response in last block, then we expect an ErrorInvalidResponse
	sendReceive(tm, rdc, "C102C1000300015E230BFF020000000000016B020A092800010203040506070809000102030405060708090001020304050607080900010203040506070809092800010203040506070809000102030405060708090001020304050607080900010203040506070809092800010203040506070809000102030405060708", "C502C100000001")
	sendReceive(tm, rdc, "C103C100000000027509000102030405060708090001020304050607080909280001020304050607080900010203040506070809000102030405060708090001020304050607080909280001020304050607080900010203040506070809000102030405060708090001020304050607080915000000000000007B150000", "C502C100000002")
	sendReceive(tm, rdc, "C103C10100000003210000000000EA1500000000000001591500000000000001C8150000000000000237", "0E010203")
	err = c.SetRequest(dlms.CreateAttributeDescriptor(3, "0-1:94.35.11.255", 2), data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	tm.AssertExpectations(t)
}
