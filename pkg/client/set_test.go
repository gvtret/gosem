package client_test

import (
	"fmt"
	"testing"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/Circutor/gosem/pkg/dlms"
	"github.com/stretchr/testify/assert"
)

func TestClient_SetRequest(t *testing.T) {
	c, tm, err := associate()
	assert.NoError(t, err)

	in := decodeHexString("C101C1000300015E230BFF02000600002710")
	out := decodeHexString("C501C100")
	tm.On("Send", in).Return(out, nil).Once()

	var data uint32 = 10000

	err = c.SetRequest(dlms.CreateAttributeDescriptor(3, "0-1:94.35.11.255", 2), data)
	assert.NoError(t, err)

	tm.AssertExpectations(t)
}

func TestClient_SetRequestFail(t *testing.T) {
	c, tm, err := associate()
	assert.NoError(t, err)

	data := axdr.CreateAxdrDoubleLongUnsigned(10000)
	demandAttributeDescriptor := dlms.CreateAttributeDescriptor(3, "0-1:94.35.11.255", 2)

	// Set failed
	in := decodeHexString("C101C1000300015E230BFF02000600002710")
	out := decodeHexString("C501C10102")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.SetRequest(demandAttributeDescriptor, data)
	assert.Error(t, err)

	// Unexpected response
	in = decodeHexString("C101C1000300015E230BFF02000600002710")
	out = decodeHexString("0E010203")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.SetRequest(demandAttributeDescriptor, data)
	assert.Error(t, err)

	// Invalid response
	in = decodeHexString("C101C1000300015E230BFF02000600002710")
	out = decodeHexString("AE12")
	tm.On("Send", in).Return(out, nil).Once()

	err = c.SetRequest(demandAttributeDescriptor, data)
	assert.Error(t, err)

	// Send failed
	in = decodeHexString("C101C1000300015E230BFF02000600002710")
	out = decodeHexString("C501C10102")
	tm.On("Send", in).Return(out, fmt.Errorf("error")).Once()
	tm.On("IsConnected").Return(false).Once()

	err = c.SetRequest(demandAttributeDescriptor, data)
	assert.Error(t, err)

	// Not associated
	tm.On("Disconnect").Return(nil).Once()
	c.Disconnect()

	err = c.SetRequest(demandAttributeDescriptor, data)
	assert.Error(t, err)

	// Invalid data
	err = c.SetRequest(demandAttributeDescriptor, nil)
	assert.Error(t, err)

	// nil attribute descriptor
	err = c.SetRequest(nil, data)
	assert.Error(t, err)

	tm.AssertExpectations(t)
}
