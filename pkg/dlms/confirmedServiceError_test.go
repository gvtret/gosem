package dlms

import (
	"bytes"
	"testing"
)

func TestNew_ConfirmedServiceError(t *testing.T) {
	cse := *CreateConfirmedServiceError(TagErrInitiateError, TagErrInitiate, 1)
	out, err := cse.Encode()
	if err != nil {
		t.Errorf("Encode Failed. Err: %v", err)
	}
	result := decodeHexString("0E010601")
	if !bytes.Equal(out, result) {
		t.Errorf("Failed. Get: %s, should: %s", encodeHexString(out), encodeHexString(result))
	}
}

func TestDecode_ConfirmedServiceError(t *testing.T) {
	src := decodeHexString("0E010601")
	cse, err := DecodeConfirmedServiceError(&src)
	if err != nil {
		t.Errorf("Failed on DecodeConfirmedServiceError. Err: %v", err)
	}

	if cse.ConfirmedServiceError != TagErrInitiateError {
		t.Errorf("Invalid ConfirmedServiceError. Get: %v", cse.ConfirmedServiceError)
	}

	if cse.ServiceError != TagErrInitiate {
		t.Errorf("Invalid ServiceError. Get: %v", cse.ServiceError)
	}

	if cse.Value != 1 {
		t.Errorf("Invalid Value. Get: %v", cse.Value)
	}

	src = decodeHexString("0E0106")
	_, err = DecodeConfirmedServiceError(&src)
	if err == nil {
		t.Errorf("Should failed on DecodeConfirmedServiceError")
	}

	src = decodeHexString("0F010601")
	_, err = DecodeConfirmedServiceError(&src)
	if err == nil {
		t.Errorf("Should failed on DecodeConfirmedServiceError")
	}
}
