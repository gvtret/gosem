package dlms

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSelectiveAccessDescriptor_Encode(t *testing.T) {
	a := *CreateSelectiveAccessByEntryDescriptor(0, 5)
	out, err := a.Encode()
	assert.NoError(t, err)

	expected := decodeHexString("02020406000000000600000005120000120000")
	assert.Equal(t, expected, out)

	timeStart := time.Date(2020, time.January, 1, 10, 0, 0, 0, time.UTC)
	timeEnd := time.Date(2020, time.January, 1, 11, 0, 0, 0, time.UTC)
	b := *CreateSelectiveAccessByRangeDescriptor(timeStart, timeEnd, nil)
	out, err = b.Encode()
	assert.NoError(t, err)

	expected = decodeHexString("010204020412000809060000010000FF0F02120000090C07E40101030A000000000000090C07E40101030B0000000000000100")
	assert.Equal(t, expected, out)

	vad := make([]AttributeDescriptor, 2)
	vad[0] = *CreateAttributeDescriptor(8, "0.0.1.0.0.255", 2)
	vad[1] = *CreateAttributeDescriptor(1, "0.0.96.10.7.255", 2)

	c := *CreateSelectiveAccessByRangeDescriptor(timeStart, timeEnd, vad)
	out, err = c.Encode()
	assert.NoError(t, err)

	expected = decodeHexString("010204020412000809060000010000FF0F02120000090C07E40101030A000000000000090C07E40101030B0000000000000102020412000809060000010000FF0F02120000020412000109060000600A07FF0F02120000")
	assert.Equal(t, expected, out)
}

func TestSelectiveAccessDescriptor_Decode(t *testing.T) {
	// ------------------------ AccessSelectorEntry
	src := decodeHexString("02020406000000000600000005120000120000")
	b := *CreateSelectiveAccessByEntryDescriptor(0, 5)

	a, err := DecodeSelectiveAccessDescriptor(&src)
	assert.NoError(t, err)
	assert.Equal(t, a.AccessSelector, b.AccessSelector)

	aByte, _ := a.AccessParameter.Encode()
	bByte, _ := b.AccessParameter.Encode()
	assert.Equal(t, aByte, bByte)

	// ------------------------ AccessSelectorRange
	src = decodeHexString("010204020412000809060000010000FF0F02120000090C07E40101030A000000000000090C07E40101030B0000000000000100")
	timeStart := time.Date(2020, time.January, 1, 10, 0, 0, 0, time.UTC)
	timeEnd := time.Date(2020, time.January, 1, 11, 0, 0, 0, time.UTC)
	b = *CreateSelectiveAccessByRangeDescriptor(timeStart, timeEnd, nil)

	a, err = DecodeSelectiveAccessDescriptor(&src)
	assert.NoError(t, err)
	assert.Equal(t, a.AccessSelector, b.AccessSelector)

	aByte, _ = a.AccessParameter.Encode()
	bByte, _ = b.AccessParameter.Encode()
	assert.Equal(t, aByte, bByte)

	// --- making sure src wont change if decode fail
	src = decodeHexString("02020406000000000600000005FF0000120000")
	oriLength := len(src)
	a, err = DecodeSelectiveAccessDescriptor(&src)
	assert.Error(t, err)
	assert.Len(t, src, oriLength)
}
