package dlms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_InitiateResponse(t *testing.T) {
	ir := *CreateInitiateResponse(nil, 0x0000101D, 128)
	out, err := ir.Encode()
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("0800065F1F040000101D00800007"), out)

	qualityOfService := uint8(2)
	ir = *CreateInitiateResponse(&qualityOfService, 0x0000101D, 128)
	out, err = ir.Encode()
	assert.NoError(t, err)
	assert.Equal(t, decodeHexString("080102065F1F040000101D00800007"), out)
}

func TestDecode_InitiateResponse(t *testing.T) {
	src := decodeHexString("0800065F1F040000101D00800007")
	ir, err := DecodeInitiateResponse(&src)
	assert.NoError(t, err)
	assert.Nil(t, ir.NegotiatedQualityOfService)
	assert.Equal(t, uint32(0x0000101D), ir.NegotiatedConformance)
	assert.Equal(t, uint16(128), ir.ServerMaxReceivePduSize)

	src = decodeHexString("080103065F1F040000101D00800007")
	ir, err = DecodeInitiateResponse(&src)
	assert.NoError(t, err)
	assert.NotNil(t, ir.NegotiatedQualityOfService)
	assert.Equal(t, uint8(3), *ir.NegotiatedQualityOfService)

	src = decodeHexString("0800065F1F040000101D008000")
	_, err = DecodeInitiateResponse(&src)
	assert.Error(t, err)

	src = decodeHexString("0900065F1F040000101D00800007")
	_, err = DecodeInitiateResponse(&src)
	assert.Error(t, err)

	src = decodeHexString("0800055F1F040000101D00800007")
	_, err = DecodeInitiateResponse(&src)
	assert.Error(t, err)

	src = decodeHexString("0800065B1F040000101D00800007")
	_, err = DecodeInitiateResponse(&src)
	assert.Error(t, err)
}
