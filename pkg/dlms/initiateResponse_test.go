package dlms

import (
	"bytes"
	"testing"
)

func TestNew_InitiateResponse(t *testing.T) {
	ir := *CreateInitiateResponse(nil, 0x0000101D, 128)
	out, err := ir.Encode()
	if err != nil {
		t.Errorf("Encode Failed. Err: %v", err)
	}

	result := decodeHexString("0800065F1F040000101D00800007")
	if !bytes.Equal(out, result) {
		t.Errorf("Failed. Get: %s, should: %s", encodeHexString(out), encodeHexString(result))
	}

	qualityOfService := uint8(2)
	ir = *CreateInitiateResponse(&qualityOfService, 0x0000101D, 128)
	out, err = ir.Encode()
	if err != nil {
		t.Errorf("Encode Failed. Err: %v", err)
	}

	result = decodeHexString("080102065F1F040000101D00800007")
	if !bytes.Equal(out, result) {
		t.Errorf("Failed. Get: %s, should: %s", encodeHexString(out), encodeHexString(result))
	}
}

func TestDecode_InitiateResponse(t *testing.T) {
	src := decodeHexString("0800065F1F040000101D00800007")
	ir, err := DecodeInitiateResponse(&src)
	if err != nil {
		t.Errorf("Failed on DecodeInitiateResponse. Err: %v", err)
	}

	if ir.NegotiatedQualityOfService != nil {
		t.Errorf("Invalid NegotiatedQualityOfService. Get: %v", ir.NegotiatedQualityOfService)
	}

	if ir.NegotiatedConformance != 0x0000101D {
		t.Errorf("Invalid NegotiatedConformance. Get: %v", ir.NegotiatedConformance)
	}

	if ir.ServerMaxReceivePduSize != 128 {
		t.Errorf("Invalid ServerMaxReceivePduSize. Get: %v", ir.ServerMaxReceivePduSize)
	}

	src = decodeHexString("080103065F1F040000101D00800007")
	ir, err = DecodeInitiateResponse(&src)
	if err != nil {
		t.Errorf("Failed on DecodeInitiateResponse. Err: %v", err)
	}

	if *ir.NegotiatedQualityOfService != 3 {
		t.Errorf("Invalid NegotiatedQualityOfService. Get: %v", *ir.NegotiatedQualityOfService)
	}

	src = decodeHexString("0800065F1F040000101D008000")
	_, err = DecodeInitiateResponse(&src)
	if err == nil {
		t.Errorf("Should failed on DecodeInitiateResponse")
	}

	src = decodeHexString("0900065F1F040000101D00800007")
	_, err = DecodeInitiateResponse(&src)
	if err == nil {
		t.Errorf("Should failed on DecodeInitiateResponse")
	}

	src = decodeHexString("0800055F1F040000101D00800007")
	_, err = DecodeInitiateResponse(&src)
	if err == nil {
		t.Errorf("Should failed on DecodeInitiateResponse")
	}

	src = decodeHexString("0800065B1F040000101D00800007")
	_, err = DecodeInitiateResponse(&src)
	if err == nil {
		t.Errorf("Should failed on DecodeInitiateResponse")
	}

	src = decodeHexString("0800065F1F040000101D00800008")
	_, err = DecodeInitiateResponse(&src)
	if err == nil {
		t.Errorf("Should failed on DecodeInitiateResponse")
	}
}
