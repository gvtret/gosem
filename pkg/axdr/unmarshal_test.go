package axdr

import (
	"encoding/hex"
	"testing"
	"time"
)

func TestUnmarshalData(t *testing.T) {
	type TestData struct {
		Time1  time.Time
		Value1 uint16
		Value2 int
		Value3 uint
	}

	src := decodeHexString("01020204090C07D00106050F0030FF003C01121234050ABBCCDD11010204090C07D00106050F0030FF003C0112567805000000001102")
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
	if result[0].Time1.Unix() != 947167248 {
		t.Errorf("Expected time to be 947167248, got %s", result[0].Time1)
	}
	if result[0].Value1 != 0x1234 {
		t.Errorf("Expected 0x1234, got 0x%X", result[0].Value1)
	}
	if result[0].Value2 != 0xABBCCDD {
		t.Errorf("Expected 0xABBCCDD, got 0x%X", result[0].Value2)
	}
	if result[1].Value3 != 0x02 {
		t.Errorf("Expected 0x02, got 0x%X", result[1].Value3)
	}
}

func TestUnmarshalPartial(t *testing.T) {
	src := decodeHexString("01020204090C07D00106050F0030FF003C01121234050ABBCCDD11010204090C07D00106050F0030FF003C0112567805000000001102")
	var result [][]DlmsData

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

	if len(result[0]) != 4 {
		t.Errorf("Expected 4 results, got %d", len(result))
	}
}

func TestUnmarshalDataFail(t *testing.T) {
	// nil data
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

	// Invalid variable kind
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

	// Missing struct element
	type anotherInvalidTestData struct {
		Value1 uint16
		Value2 int32
	}

	var anotherInvalidResult []anotherInvalidTestData

	err = UnmarshalData(data, &anotherInvalidResult)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Invalid time format
	src = decodeHexString("090B07D00106050F0030FF003C")
	dec = NewDataDecoder(&src)
	data, _ = dec.Decode(&src)

	var time1 time.Time

	err = UnmarshalData(data, &time1)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
