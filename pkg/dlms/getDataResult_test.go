package dlms

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/circutor-library/gosem/pkg/axdr"
)

func TestAccessResult(t *testing.T) {
	t1 := TagAccSuccess
	if t1.String() != "success" {
		t.Errorf("t1 should return string with value 'success'")
	}
	t2 := TagAccObjectUnavailable
	if t2.String() != "object-unavailable" {
		t.Errorf("t1 should return string with value 'object-unavailable'")
	}
}

func TestGetDataResultAsResult(t *testing.T) {
	a := *CreateGetDataResultAsResult(TagAccSuccess)

	t1, e := a.Encode()
	if e != nil {
		t.Errorf("t1 Encode Failed. err: %v", e)
	}
	result := []byte{0, 0}

	res := bytes.Compare(t1, result)
	if res != 0 {
		t.Errorf("t1 Failed. get: %d, should:%v", t1, result)
	}
}

func TestGetDataResultAsData(t *testing.T) {
	dt := *axdr.CreateAxdrDoubleLong(69)
	a := *CreateGetDataResultAsData(dt)

	t1, e := a.Encode()
	if e != nil {
		t.Errorf("t1 Encode Failed. err: %v", e)
	}
	result := []byte{1, 5, 0, 0, 0, 69}

	res := bytes.Compare(t1, result)
	if res != 0 {
		t.Errorf("t1 Failed. get: %d, should:%v", t1, result)
	}
}

func TestGetDataResult(t *testing.T) {
	rs := TagAccSuccess
	a := *CreateGetDataResult(rs)

	t1, e := a.Encode()
	if e != nil {
		t.Errorf("t1 Encode Failed. err: %v", e)
	}
	result := []byte{0, 0}
	res := bytes.Compare(t1, result)
	if res != 0 {
		t.Errorf("t1 Failed. get: %d, should:%v", t1, result)
	}

	dt := *axdr.CreateAxdrDoubleLong(69)
	b := *CreateGetDataResult(dt)
	t2, e := b.Encode()
	if e != nil {
		t.Errorf("t2 Encode Failed. err: %v", e)
	}
	result = []byte{1, 5, 0, 0, 0, 69}

	res = bytes.Compare(t2, result)
	if res != 0 {
		t.Errorf("t2 Failed. get: %d, should:%v", t2, result)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("t3 should've panic on wrong Value")
		}
	}()
	c := *CreateGetDataResult(999)
	c.Encode()
}

func TestDataBlockGAsData(t *testing.T) {
	a := *CreateDataBlockGAsData(true, 1, "07D20C04030A060BFF007800")
	encoded, err := a.Encode()
	assert.NoError(t, err)
	expected := decodeHexString("0100000001000C07D20C04030A060BFF007800")
	assert.Equal(t, expected, encoded)

	b := *CreateDataBlockGAsData(true, 1, []byte{1, 0, 0, 3, 0, 255})
	encoded, err = b.Encode()
	assert.NoError(t, err)
	expected = decodeHexString("010000000100060100000300FF")
	assert.Equal(t, expected, encoded)

	assert.Panics(t, func() {
		c := *CreateDataBlockGAsData(true, 1, TagAccSuccess)
		c.Encode()
	})
}

func TestDataBlockGAsResult(t *testing.T) {
	a := *CreateDataBlockGAsResult(true, 1, TagAccSuccess)
	encoded, err := a.Encode()
	assert.NoError(t, err)
	expected := decodeHexString("01000000010100")
	assert.Equal(t, expected, encoded)
}

func TestDataBlockG(t *testing.T) {
	a := *CreateDataBlockG(true, 1, "07D20C04030A060BFF007800")
	encoded, err := a.Encode()
	assert.NoError(t, err)
	expected := decodeHexString("0100000001000C07D20C04030A060BFF007800")
	assert.Equal(t, expected, encoded)

	b := *CreateDataBlockG(true, 1, []byte{1, 0, 0, 3, 0, 255})
	encoded, err = b.Encode()
	assert.NoError(t, err)
	expected = decodeHexString("010000000100060100000300FF")
	assert.Equal(t, expected, encoded)

	c := *CreateDataBlockG(true, 1, TagAccSuccess)
	encoded, err = c.Encode()
	assert.NoError(t, err)
	expected = decodeHexString("01000000010100")
	assert.Equal(t, expected, encoded)
}

func TestDataBlockSA(t *testing.T) {
	// with hexstring
	a := *CreateDataBlockSA(true, 1, "07D20C04030A060BFF007800")

	result, err := a.Encode()
	assert.NoError(t, err)
	expected := []byte{1, 0, 0, 0, 1, 12, 7, 210, 12, 4, 3, 10, 6, 11, 255, 0, 120, 0}
	assert.Equal(t, expected, result)

	// with byte slice
	b := *CreateDataBlockSA(true, 1, []byte{1, 0, 0, 3, 0, 255})
	result, err = b.Encode()
	assert.NoError(t, err)
	expected = []byte{1, 0, 0, 0, 1, 6, 1, 0, 0, 3, 0, 255}
	assert.Equal(t, expected, result)

	// with wrong value
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("t3 should've panic on wrong Value")
		}
	}()
	c := *CreateDataBlockSA(true, 1, TagAccSuccess)
	c.Encode()
}

func TestActResponse(t *testing.T) {
	dt := *axdr.CreateAxdrDoubleLong(69)
	ret := *CreateGetDataResultAsData(dt)
	a := *CreateActResponse(TagActSuccess, &ret)

	t1, e := a.Encode()
	if e != nil {
		t.Errorf("t1 Encode Failed. err: %v", e)
	}
	result := []byte{0, 1, 1, 5, 0, 0, 0, 69}

	res := bytes.Compare(t1, result)
	if res != 0 {
		t.Errorf("t1 Failed. get: %d, should:%v", t1, result)
	}

	// with nil GetDataResult
	var nilRet *GetDataResult
	b := *CreateActResponse(TagActReadWriteDenied, nilRet)
	t2, e := b.Encode()
	if e != nil {
		t.Errorf("t2 Encode Failed. err: %v", e)
	}
	result = []byte{3, 0}

	res = bytes.Compare(t2, result)
	if res != 0 {
		t.Errorf("t2 Failed. get: %d, should:%v", t2, result)
	}
}

func TestDecode_GetDataResult(t *testing.T) {
	// with AccessResultTag
	src := []byte{1, 0}
	a, ae := DecodeGetDataResult(&src)

	if ae != nil {
		t.Errorf("t1 Failed. got error: %v", ae)
	}
	if a.IsData {
		t.Errorf("t1 Failed. Value should be access")
	}
	if a.Value != TagAccSuccess {
		t.Errorf("t1 Failed. get: %d, should:%v", a.Value, TagAccSuccess)
	}

	// with dlms data
	src = []byte{0, 5, 0, 0, 0, 69}
	b, be := DecodeGetDataResult(&src)
	if be != nil {
		t.Errorf("t2 Failed. got error: %v", be)
	}
	if !b.IsData {
		t.Errorf("t2 Failed. Value should be data")
	}
	val := b.Value.(axdr.DlmsData)
	if val.Tag != axdr.TagDoubleLong {
		t.Errorf("t2 Failed. get: %d, should:%v", val.Tag, axdr.TagDoubleLong)
	}
	if v := val.Value.(int32); v != 69 {
		t.Errorf("t2 Failed. get: %d, should:%v", v, 69)
	}
}

func TestDecode_DataBlockG(t *testing.T) {
	// with byte slice
	src := []byte{1, 0, 0, 0, 1, 0, 12, 7, 210, 12, 4, 3, 10, 6, 11, 255, 0, 120, 0}
	a, ae := DecodeDataBlockG(&src)

	if ae != nil {
		t.Errorf("t1 Failed. got error: %v", ae)
	}
	if !a.LastBlock {
		t.Errorf("t1 Failed. LastBlock should be true")
	}
	if a.BlockNumber != 1 {
		t.Errorf("t1 Failed. BlockNumber should be 1 (%v)", a.BlockNumber)
	}
	if a.IsResult {
		t.Errorf("t1 Failed. IsResult should be false")
	}
	val, _ := a.ResultAsBytes()
	res := bytes.Compare(val, []byte{7, 210, 12, 4, 3, 10, 6, 11, 255, 0, 120, 0})
	if res != 0 {
		t.Errorf("t1 Failed. Result is not correct (%v)", val)
	}

	// with AccessResultTag
	src = []byte{1, 0, 0, 0, 1, 1, 0}
	b, be := DecodeDataBlockG(&src)

	if be != nil {
		t.Errorf("t2 Failed. got error: %v", be)
	}
	if !b.LastBlock {
		t.Errorf("t2 Failed. LastBlock should be true")
	}
	if b.BlockNumber != 1 {
		t.Errorf("t2 Failed. BlockNumber should be 1 (%v)", b.BlockNumber)
	}
	if !b.IsResult {
		t.Errorf("t2 Failed. IsResult should be true")
	}
	_, eTag := b.ResultAsAccess()
	if eTag != nil {
		t.Errorf("t2 Failed. Result should be TagAccSuccess (%v)", eTag)
	}
}

func TestDecode_DataBlockSA(t *testing.T) {
	src := []byte{1, 0, 0, 0, 1, 12, 7, 210, 12, 4, 3, 10, 6, 11, 255, 0, 120, 0}
	a, ae := DecodeDataBlockSA(&src)

	if ae != nil {
		t.Errorf("t1 Failed. got error: %v", ae)
	}
	if !a.LastBlock {
		t.Errorf("t1 Failed. LastBlock should be true")
	}
	if a.BlockNumber != 1 {
		t.Errorf("t1 Failed. BlockNumber should be 1 (%v)", a.BlockNumber)
	}
	res := bytes.Compare(a.Raw, []byte{7, 210, 12, 4, 3, 10, 6, 11, 255, 0, 120, 0})
	if res != 0 {
		t.Errorf("t1 Failed. Result is not correct (%v)", a.Raw)
	}
}

func TestDecode_ActResponse(t *testing.T) {
	src := []byte{0, 1, 1, 0}
	a, ae := DecodeActResponse(&src)

	if ae != nil {
		t.Errorf("Error on DecodeActResponse: %v", ae)
	}
	if a.Result != TagActSuccess {
		t.Errorf("Result should be TagActSuccess")
	}
	if a.ReturnParam == nil {
		t.Errorf("ReturnParam should not be nil (%v)", a.ReturnParam)
	}
	if a.ReturnParam.IsData == true {
		t.Errorf("ReturnParam.IsData should not be true (%v)", a.ReturnParam.IsData)
	}
	tag, err := a.ReturnParam.ValueAsAccess()
	if tag != TagAccSuccess || err != nil {
		t.Errorf("ReturnParam.Value should be TagAccSuccess (%v, err: %v)", tag, err)
	}
}
