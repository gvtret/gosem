package dlms

import (
	"testing"
	"time"

	"github.com/Circutor/gosem/pkg/axdr"
	"github.com/stretchr/testify/assert"
)

func TestDataNotification_New(t *testing.T) {
	invokeIDAndPriority := uint32(6543210)
	tm := time.Date(2023, time.January, 31, 18, 34, 23, 0, time.UTC)
	dataValue := *axdr.CreateAxdrBoolean(true)

	dn := *CreateDataNotification(invokeIDAndPriority, &tm, dataValue)
	encoded, err := dn.Encode()
	assert.NoError(t, err)

	expected := decodeHexString("0F0063D76A0C07E7011F021222170000000003FF")
	assert.Equal(t, expected, encoded)

	// With nil time
	dn = *CreateDataNotification(invokeIDAndPriority, nil, dataValue)
	encoded, err = dn.Encode()
	assert.NoError(t, err)

	expected = decodeHexString("0F0063D76A0003FF")
	assert.Equal(t, expected, encoded)
}

func TestDataNotification_Decode(t *testing.T) {
	src := decodeHexString("0F0063D76A0C07E7011F021222170000000003FF")
	dn, err := DecodeDataNotification(&src)
	assert.NoError(t, err)

	invokeIDAndPriority := uint32(6543210)
	tm := time.Date(2023, time.January, 31, 18, 34, 23, 0, time.UTC)
	dataValue := *axdr.CreateAxdrBoolean(true)

	assert.Equal(t, invokeIDAndPriority, dn.InvokeIDAndPriority)
	assert.Equal(t, tm, *dn.DateTime)
	assert.Equal(t, dataValue.Tag, dn.DataValue.Tag)
	assert.Equal(t, dataValue.Value, dn.DataValue.Value)

	// With nil time
	src = decodeHexString("0F0063D76A0003FF")
	dn, err = DecodeDataNotification(&src)
	assert.NoError(t, err)
	assert.Nil(t, dn.DateTime)

	// Invalid frames
	src = decodeHexString("")
	_, err = DecodeDataNotification(&src)
	assert.Error(t, err)

	src = decodeHexString("00")
	_, err = DecodeDataNotification(&src)
	assert.Error(t, err)

	src = decodeHexString("0F00")
	_, err = DecodeDataNotification(&src)
	assert.Error(t, err)

	src = decodeHexString("0F0063D76A")
	_, err = DecodeDataNotification(&src)
	assert.Error(t, err)

	src = decodeHexString("0F0063D76A0103FF")
	_, err = DecodeDataNotification(&src)
	assert.Error(t, err)

	src = decodeHexString("0F0063D76A0003")
	_, err = DecodeDataNotification(&src)
	assert.Error(t, err)
}
