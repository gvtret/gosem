package axdr

import (
	"encoding/hex"
	"testing"
)

func TestUnmarshalData(t *testing.T) {
	type TestData struct {
		Value1 uint16
		Value2 int32
		Value3 uint8
	}

	src := decodeHexString("0102020312123405000000001101020312567805000000001102")
	var result []TestData

	dec := NewDataDecoder(&src)
	data, err := dec.Decode(&src)
	if err != nil {
		t.Errorf("Error decoding data: %v", err)
	}

	err = UnmarshalData(data, &result)
	if err != nil {
		t.Errorf("Error unmarshaling data: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}
	if result[0].Value1 != 0x1234 {
		t.Errorf("Expected 0x1234, got %x", result[0].Value1)
	}
	if result[1].Value3 != 0x02 {
		t.Errorf("Expected 0x02, got %x", result[1].Value3)
	}
}

func TestUnmarshalDataFail(t *testing.T) {
	src := decodeHexString("0102020312123405000000001101020312567805000000001102")

	dec := NewDataDecoder(&src)
	data, err := dec.Decode(&src)
	if err != nil {
		t.Errorf("Error decoding data: %v", err)
	}

	err = UnmarshalData(data, nil)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	type invalidTestData struct {
		Value1 uint16
		Value2 int32
		Value3 uint16
	}

	var invalidResult []invalidTestData

	err = UnmarshalData(data, &invalidResult)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	type anotherInvalidTestData struct {
		Value1 uint16
		Value2 int32
	}

	var anotherInvalidResult []anotherInvalidTestData

	err = UnmarshalData(data, &anotherInvalidResult)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
