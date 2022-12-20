package dlmsclient_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Circutor/gosem/pkg/dlms"
	"github.com/Circutor/gosem/pkg/dlms/mocks"
	"github.com/Circutor/gosem/pkg/dlmsclient"
	"github.com/stretchr/testify/assert"
)

func TestClient_GetRequest(t *testing.T) {
	c, tm, err := associate()
	assert.NoError(t, err)

	in := decodeHexString("C001C100080000010000FF0300")
	out := decodeHexString("C401C10010003C")
	tm.On("Send", in).Return(out, nil).Once()

	var data int16

	err = c.GetRequest(dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3), &data)
	assert.NoError(t, err)
	assert.Equal(t, int16(0x003C), data)

	tm.AssertExpectations(t)
}

func TestClient_GetRequestFail(t *testing.T) {
	c, tm, err := associate()
	assert.NoError(t, err)

	var data int32
	clockAttributeDescriptor := dlms.CreateAttributeDescriptor(8, "0-0:1.0.0.255", 3)

	// Get failed
	in := decodeHexString("C001C100080000010000FF0300")
	out := decodeHexString("C401C10102")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.GetRequest(clockAttributeDescriptor, &data)
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorGetRejected, clientError.Code())

	// Unexpected response
	in = decodeHexString("C001C100080000010000FF0300")
	out = decodeHexString("0E010203")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.GetRequest(clockAttributeDescriptor, &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Invalid response
	in = decodeHexString("C001C100080000010000FF0300")
	out = decodeHexString("AE12")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.GetRequest(clockAttributeDescriptor, &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Response type doesn't match
	in = decodeHexString("C001C100080000010000FF0300")
	out = decodeHexString("C401C10010003C")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.GetRequest(clockAttributeDescriptor, &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Send failed
	in = decodeHexString("C001C100080000010000FF0300")
	out = decodeHexString("C401C10102")
	tm.On("Send", in).Return(out, fmt.Errorf("error")).Once()
	tm.On("IsConnected").Return(false).Once()

	err = c.GetRequest(clockAttributeDescriptor, &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorCommunicationFailed, clientError.Code())

	// Not associated
	tm.On("Disconnect").Return(nil).Once()
	c.Disconnect()

	err = c.GetRequest(clockAttributeDescriptor, &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidState, clientError.Code())

	// nil attribute descriptor
	err = c.GetRequest(nil, &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidParameter, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_GetRequestWithDataBlock(t *testing.T) {
	c, tm, err := associate()
	assert.NoError(t, err)

	in := decodeHexString("C001C100070100630100FF0200")
	out := decodeHexString("C402C10000000001000C010506000000010600000002")
	tm.On("Send", in).Return(out, nil).Once()

	in = decodeHexString("C002C100000001")
	out = decodeHexString("C402C10000000002000A06000000030600000004")
	tm.On("Send", in).Return(out, nil).Once()

	in = decodeHexString("C002C100000002")
	out = decodeHexString("C402C1010000000300050600000005")
	tm.On("Send", in).Return(out, nil).Once()

	var data []uint32

	err = c.GetRequest(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), &data)
	assert.NoError(t, err)
	assert.Len(t, data, 5)

	tm.AssertExpectations(t)
}

func TestClient_GetRequestWithDataBlockFail(t *testing.T) {
	c, tm, err := associate()
	assert.NoError(t, err)

	var data []uint32

	// Get failed
	in := decodeHexString("C001C100070100630100FF0200")
	out := decodeHexString("C402C100000000010102")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.GetRequest(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), &data)
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorGetRejected, clientError.Code())

	// Invalid block number
	in = decodeHexString("C001C100070100630100FF0200")
	out = decodeHexString("C402C10000000002000C010506000000010600000002")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.GetRequest(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Invalid response
	in = decodeHexString("C001C100070100630100FF0200")
	out = decodeHexString("C402C10000000001000C010506000000010600000002")
	tm.On("Send", in).Return(out, nil).Once()

	in = decodeHexString("C002C100000001")
	out = decodeHexString("AE12")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.GetRequest(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Unexpected response
	in = decodeHexString("C001C100070100630100FF0200")
	out = decodeHexString("C402C10000000001000C010506000000010600000002")
	tm.On("Send", in).Return(out, nil).Once()

	in = decodeHexString("C002C100000001")
	out = decodeHexString("0E010203")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.GetRequest(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	// Invalid data
	in = decodeHexString("C001C100070100630100FF0200")
	out = decodeHexString("C402C10100000001000C010506000000010600000002")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.GetRequest(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), &data)
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorInvalidResponse, clientError.Code())

	tm.AssertExpectations(t)
}

func TestClient_GetRequestRequestWithSelectiveAccess(t *testing.T) {
	c, tm, err := associate()
	assert.NoError(t, err)

	in := decodeHexString("C001C100070100630100FF0201010204020412000809060000010000FF0F02120000090C07E40101030A000000000000090C07E40101030B0000000000000100")
	out := decodeHexString("C401C100010206000000010600000002")
	tm.On("Send", in).Return(out, nil).Once()

	var data []uint32

	timeStart := time.Date(2020, time.January, 1, 10, 0, 0, 0, time.UTC)
	timeEnd := time.Date(2020, time.January, 1, 11, 0, 0, 0, time.UTC)
	err = c.GetRequestWithSelectiveAccessByDate(dlms.CreateAttributeDescriptor(7, "1-0:99.1.0.255", 2), timeStart, timeEnd, &data)
	assert.NoError(t, err)
	assert.Len(t, data, 2)

	tm.AssertExpectations(t)
}

func TestClient_GetRequestWithStructOfElements(t *testing.T) {
	var data struct {
		Value1 uint  `obis:"1,1-1:94.34.100.255,2"`
		Value2 *uint `obis:"1,1-1:94.34.104.255,2"`
		Value3 *uint `obis:"70,0-0:96.3.10.255,3"`
	}

	c, tm, err := associate()
	assert.NoError(t, err)

	in := decodeHexString("C001C1000101015E2264FF0200")
	out := decodeHexString("C401C1001104")
	tm.On("Send", in).Return(out, nil).Once()

	in = decodeHexString("C001C1000101015E2268FF0200")
	out = decodeHexString("C401C1001101")
	tm.On("Send", in).Return(out, nil).Once()

	in = decodeHexString("C001C10046000060030AFF0300")
	out = decodeHexString("C401C10109")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.GetRequestWithStructOfElements(&data)
	assert.NoError(t, err)
	assert.Equal(t, uint(4), data.Value1)
	assert.Equal(t, uint(1), *data.Value2)
	assert.Nil(t, data.Value3)

	tm.AssertExpectations(t)
}

func TestClient_GetRequestWithNestedStructOfElements(t *testing.T) {
	type data2 struct {
		Value uint `obis:"1,1-1:94.34.104.255,2"`
	}

	type data3 struct {
		Value *uint `obis:"70,0-0:96.3.10.255,3"`
	}

	var data struct {
		Value1 uint `obis:"1,1-1:94.34.100.255,2"`
		Data2  data2
		Data3  data3
	}

	c, tm, err := associate()
	assert.NoError(t, err)

	in := decodeHexString("C001C1000101015E2264FF0200")
	out := decodeHexString("C401C1001104")
	tm.On("Send", in).Return(out, nil).Once()

	in = decodeHexString("C001C1000101015E2268FF0200")
	out = decodeHexString("C401C1001101")
	tm.On("Send", in).Return(out, nil).Once()

	in = decodeHexString("C001C10046000060030AFF0300")
	out = decodeHexString("C401C10109")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.GetRequestWithStructOfElements(&data)
	assert.NoError(t, err)
	assert.Equal(t, uint(4), data.Value1)
	assert.Equal(t, uint(1), data.Data2.Value)
	assert.Nil(t, data.Data3.Value)

	tm.AssertExpectations(t)
}

func TestClient_CheckRequestWithStructOfElements(t *testing.T) {
	var data struct {
		Value1 *uint8 `obis:"1,1-1:94.34.104.255,2"`
		Value2 string `obis:"1,0-0:96.1.1.255,2"`
	}

	value1 := uint8(4)
	value2 := "2043594B3132"

	data.Value1 = &value1
	data.Value2 = value2

	c, tm, err := associate()
	assert.NoError(t, err)

	in := decodeHexString("C001C1000101015E2268FF0200")
	out := decodeHexString("C401C1001104")
	tm.On("Send", in).Return(out, nil).Once()

	in = decodeHexString("C001C100010000600101FF0200")
	out = decodeHexString("C401C10009062043594B3132")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.CheckRequestWithStructOfElements(&data)
	assert.NoError(t, err)

	tm.AssertExpectations(t)

	// If the first value is nil, just check the second value
	data.Value1 = nil

	in = decodeHexString("C001C100010000600101FF0200")
	out = decodeHexString("C401C10009062043594B3132")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.CheckRequestWithStructOfElements(&data)
	assert.NoError(t, err)

	tm.AssertExpectations(t)

	// If the first value doesn't match, should fail
	value1 = 8
	data.Value1 = &value1

	in = decodeHexString("C001C1000101015E2268FF0200")
	out = decodeHexString("C401C1001104")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.CheckRequestWithStructOfElements(&data)
	var clientError *dlms.Error
	assert.ErrorAs(t, err, &clientError)
	assert.Equal(t, dlms.ErrorCheckDoesNotMatch, clientError.Code())

	tm.AssertExpectations(t)
}

func associate() (dlms.Client, *mocks.TransportMock, error) {
	in := decodeHexString("601DA109060760857405080101BE10040E01000000065F1F040000181F0100")
	out := decodeHexString("6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")

	transportMock := new(mocks.TransportMock)
	transportMock.On("Connect").Return(nil).Once()
	transportMock.On("Send", in).Return(out, nil).Once()
	transportMock.On("IsConnected").Return(true).Once()

	settings, _ := dlms.NewSettingsWithoutAuthentication()
	c := dlmsclient.New(settings, transportMock, 0)

	c.Connect()

	err := c.Associate()
	if err != nil {
		err = fmt.Errorf("Associate failed: %w", err)
		return nil, nil, err
	}

	return c, transportMock, nil
}
