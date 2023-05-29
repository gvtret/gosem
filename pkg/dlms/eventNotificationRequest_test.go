package dlms

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/circutor-library/gosem/pkg/axdr"
)

func TestNew_EventNotificationRequest(t *testing.T) {
	tm := time.Date(1500, time.January, 1, 0, 0, 0, 0, time.UTC)
	attrDesc := *CreateAttributeDescriptor(1, "1.0.0.3.0.255", 2)
	attrVal := *axdr.CreateAxdrBoolean(true)

	enr := *CreateEventNotificationRequest(&tm, attrDesc, attrVal)
	encoded, err := enr.Encode()
	assert.NoError(t, err)

	expected := decodeHexString("C2010C05DC0101010000000000000000010100000300FF0203FF")
	assert.Equal(t, expected, encoded)

	// With nil time
	enr = *CreateEventNotificationRequest(nil, attrDesc, attrVal)
	encoded, err = enr.Encode()
	assert.NoError(t, err)

	expected = decodeHexString("C20000010100000300FF0203FF")
	assert.Equal(t, expected, encoded)
}

func TestDecode_EventNotificationRequest(t *testing.T) {
	src := decodeHexString("C2010C05DC0101010000000000000000010100000300FF0203FF")
	enr, err := DecodeEventNotificationRequest(&src)
	assert.NoError(t, err)

	tm := time.Date(1500, time.January, 1, 0, 0, 0, 0, time.UTC)
	attrDesc := *CreateAttributeDescriptor(1, "1.0.0.3.0.255", 2)
	attrVal := *axdr.CreateAxdrBoolean(true)

	assert.Equal(t, tm, *enr.Time)
	assert.Equal(t, attrDesc.ClassID, enr.AttributeInfo.ClassID)
	assert.Equal(t, attrDesc.InstanceID.Bytes(), enr.AttributeInfo.InstanceID.Bytes())
	assert.Equal(t, attrDesc.AttributeID, enr.AttributeInfo.AttributeID)
	assert.Equal(t, attrVal.Tag, enr.AttributeValue.Tag)
	assert.Equal(t, attrVal.Value, enr.AttributeValue.Value)

	// With nil time
	src = decodeHexString("C20000010100000300FF0203FF")
	enr, err = DecodeEventNotificationRequest(&src)
	assert.NoError(t, err)
	assert.Nil(t, enr.Time)

	// Invalid frames
	src = decodeHexString("")
	_, err = DecodeEventNotificationRequest(&src)
	assert.Error(t, err)

	src = decodeHexString("00")
	_, err = DecodeEventNotificationRequest(&src)
	assert.Error(t, err)

	src = decodeHexString("C2")
	_, err = DecodeEventNotificationRequest(&src)
	assert.Error(t, err)

	src = decodeHexString("C200")
	_, err = DecodeEventNotificationRequest(&src)
	assert.Error(t, err)

	src = decodeHexString("C20101")
	_, err = DecodeEventNotificationRequest(&src)
	assert.Error(t, err)

	src = decodeHexString("C20000010100000300")
	_, err = DecodeEventNotificationRequest(&src)
	assert.Error(t, err)
}
