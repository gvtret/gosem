package dlms

import (
	"bytes"
	"testing"
)

func TestNew_ExceptionResponse(t *testing.T) {
	var er ExceptionResponse = *CreateExceptionResponse(TagExcServiceNotAllowed, TagExcServiceNotSupported)
	out, err := er.Encode()
	if err != nil {
		t.Errorf("t1 Encode Failed. err: %v", err)
	}
	result := decodeHexString("D80102")
	if !bytes.Equal(out, result) {
		t.Errorf("Failed. Get: %s, should: %s", encodeHexString(out), encodeHexString(result))
	}
}

func TestDecode_ExceptionResponse(t *testing.T) {
	src := decodeHexString("D80203")
	er, err := DecodeExceptionResponse(&src)
	if err != nil {
		t.Errorf("Failed on DecodeExceptionResponse. Err: %v", err)
	}

	if er.StateError != TagExcServiceUnknown {
		t.Errorf("Invalid StateError. Get: %v", er.StateError)
	}

	if er.ServiceError != TagExcOtherReason {
		t.Errorf("Invalid ServiceError. Get: %v", er.ServiceError)
	}

	src = decodeHexString("D801")
	_, err = DecodeExceptionResponse(&src)
	if err == nil {
		t.Errorf("Should failed on DecodeExceptionResponse")
	}

	src = decodeHexString("D90102")
	_, err = DecodeExceptionResponse(&src)
	if err == nil {
		t.Errorf("Should failed on DecodeExceptionResponse")
	}
}
