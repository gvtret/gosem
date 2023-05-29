package dlms

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/circutor-library/gosem/pkg/axdr"
)

func TestNew_ActionResponseNormal(t *testing.T) {
	ret := *CreateGetDataResultAsResult(TagAccSuccess)
	ares := *CreateActResponse(TagActSuccess, &ret)
	a := *CreateActionResponseNormal(81, ares)
	t1, e := a.Encode()
	if e != nil {
		t.Errorf("t1 Encode Failed. err: %v", e)
	}

	result := []byte{199, 1, 81, 0, 1, 0, 0}
	res := bytes.Compare(t1, result)
	if res != 0 {
		t.Errorf("t1 Failed. get: %d, should:%v", t1, result)
	}
}

func TestNew_ActionResponseWithPBlock(t *testing.T) {
	dt := *CreateDataBlockSA(true, 1, []byte{1, 2, 3, 4, 5})
	a := *CreateActionResponseWithPBlock(81, dt)
	result, err := a.Encode()
	assert.NoError(t, err)
	expected := []byte{199, 2, 81, 1, 0, 0, 0, 1, 5, 1, 2, 3, 4, 5}
	assert.Equal(t, expected, result)
}

func TestNew_ActionResponseWithList(t *testing.T) {
	// with 1 ActResponse
	ret := *CreateGetDataResultAsResult(TagAccSuccess)
	ares1 := *CreateActResponse(TagActSuccess, &ret)
	a := *CreateActionResponseWithList(81, []ActResponse{ares1})
	t1, e := a.Encode()
	if e != nil {
		t.Errorf("t1 Encode Failed. err: %v", e)
	}

	result := []byte{199, 3, 81, 1, 0, 1, 0, 0}
	res := bytes.Compare(t1, result)
	if res != 0 {
		t.Errorf("t1 Failed. get: %d, should:%v", t1, result)
	}

	// with 2 ActResponse
	dt := *axdr.CreateAxdrDoubleLong(69)
	ret2 := *CreateGetDataResultAsData(dt)
	ares2 := *CreateActResponse(TagActSuccess, &ret2)
	b := *CreateActionResponseWithList(81, []ActResponse{ares1, ares2})
	t2, e := b.Encode()
	if e != nil {
		t.Errorf("t2 Encode Failed. err: %v", e)
	}

	result = []byte{199, 3, 81, 2, 0, 1, 0, 0, 0, 1, 1, 5, 0, 0, 0, 69}
	res = bytes.Compare(t2, result)
	if res != 0 {
		t.Errorf("t2 Failed. get: %d, should:%v", t2, result)
	}
}

func TestNew_ActionResponseNextPBlock(t *testing.T) {
	a := *CreateActionResponseNextPBlock(81, 1)
	t1, e := a.Encode()
	if e != nil {
		t.Errorf("t1 Encode Failed. err: %v", e)
	}
	result := []byte{199, 4, 81, 0, 0, 0, 1}
	res := bytes.Compare(t1, result)
	if res != 0 {
		t.Errorf("t1 Failed. get: %d, should:%v", t1, result)
	}
}

func TestDecode_ActionResponseNormal(t *testing.T) {
	src := []byte{199, 1, 81, 0, 1, 1, 0}
	a, err := DecodeActionResponseNormal(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeActionResponseNormal. err:%v", err)
	}

	ret := *CreateGetDataResultAsResult(TagAccSuccess)
	ares := *CreateActResponse(TagActSuccess, &ret)
	b := *CreateActionResponseNormal(81, ares)

	if a.InvokePriority != b.InvokePriority {
		t.Errorf("t1 Failed. InvokePriority get: %v, should:%v", a.InvokePriority, b.InvokePriority)
	}
	if a.Response.ReturnParam.IsData != b.Response.ReturnParam.IsData {
		t.Errorf("t1 Failed. Response.IsData get: %v, should:%v", a.Response.ReturnParam.IsData, b.Response.ReturnParam.IsData)
	}
	if a.Response.ReturnParam.Value != b.Response.ReturnParam.Value {
		t.Errorf("t1 Failed. Response.Value get: %v, should:%v", a.Response.ReturnParam.Value, b.Response.ReturnParam.Value)
	}

	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}
}

func TestDecode_ActionResponseWithPBlock(t *testing.T) {
	src := []byte{199, 2, 81, 1, 0, 0, 0, 1, 5, 1, 2, 3, 4, 5}
	a, err := DecodeActionResponseWithPBlock(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeActionResponseWithPBlock. err:%v", err)
	}

	dt := *CreateDataBlockSA(true, 1, []byte{1, 2, 3, 4, 5})
	b := *CreateActionResponseWithPBlock(81, dt)

	if a.InvokePriority != b.InvokePriority {
		t.Errorf("t1 Failed. InvokePriority get: %v, should:%v", a.InvokePriority, b.InvokePriority)
	}

	if a.PBlock.LastBlock != b.PBlock.LastBlock {
		t.Errorf("t1 Failed. PBlock.LastBlock get: %v, should:%v", a.PBlock.LastBlock, b.PBlock.LastBlock)
	}
	if a.PBlock.BlockNumber != b.PBlock.BlockNumber {
		t.Errorf("t1 Failed. PBlock.BlockNumber get: %v, should:%v", a.PBlock.BlockNumber, b.PBlock.BlockNumber)
	}
	res := bytes.Compare(a.PBlock.Raw, b.PBlock.Raw)
	if res != 0 {
		t.Errorf("t1 Failed. PBlock.Raw get: %v, should:%v", a.PBlock.Raw, b.PBlock.Raw)
	}

	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}
}

func TestDecode_ActionResponseWithList(t *testing.T) {
	// ---------------------- with 1 ActResponse
	src := []byte{199, 3, 81, 1, 0, 1, 1, 0}
	a, err := DecodeActionResponseWithList(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeActionResponseWithList. err:%v", err)
	}

	ret := *CreateGetDataResultAsResult(TagAccSuccess)
	ares1 := *CreateActResponse(TagActSuccess, &ret)
	b := *CreateActionResponseWithList(81, []ActResponse{ares1})

	if a.InvokePriority != b.InvokePriority {
		t.Errorf("t1 Failed. InvokePriority get: %v, should:%v", a.InvokePriority, b.InvokePriority)
	}
	if a.ResponseCount != b.ResponseCount {
		t.Errorf("t1 Failed. ResponseCount get: %v, should:%v", a.ResponseCount, b.ResponseCount)
	}

	if a.ResponseList[0].Result != b.ResponseList[0].Result {
		t.Errorf("t1 Failed. ResponseList[0].Result get: %v, should:%v", a.ResponseList[0].Result, b.ResponseList[0].Result)
	}

	aData1, _ := a.ResponseList[0].ReturnParam.ValueAsAccess()
	bData1, _ := b.ResponseList[0].ReturnParam.ValueAsAccess()
	if aData1 != bData1 {
		t.Errorf("t1 Failed. ResponseList[0].ReturnParam.Value get: %v, should:%v", aData1, bData1)
	}

	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}

	// ---------------------- with 2 ActResponse
	src = []byte{199, 3, 81, 2, 0, 1, 1, 0, 0, 1, 0, 5, 0, 0, 0, 69}
	a, err = DecodeActionResponseWithList(&src)
	if err != nil {
		t.Errorf("t2 Failed to DecodeActionResponseWithList. err:%v", err)
	}

	dt := *axdr.CreateAxdrDoubleLong(69)
	ret2 := *CreateGetDataResultAsData(dt)
	ares2 := *CreateActResponse(TagActSuccess, &ret2)
	b = *CreateActionResponseWithList(81, []ActResponse{ares1, ares2})

	if a.ResponseCount != b.ResponseCount {
		t.Errorf("t1 Failed. ResponseCount get: %v, should:%v", a.ResponseCount, b.ResponseCount)
	}

	if a.ResponseList[1].Result != b.ResponseList[1].Result {
		t.Errorf("t1 Failed. ResponseList[1].Result get: %v, should:%v", a.ResponseList[1].Result, b.ResponseList[1].Result)
	}
	if a.ResponseList[1].ReturnParam.IsData != b.ResponseList[1].ReturnParam.IsData {
		t.Errorf("t1 Failed. ResponseList[1].ReturnParam.IsData get: %v, should:%v", a.ResponseList[1].ReturnParam.IsData, b.ResponseList[1].ReturnParam.IsData)
	}

	aData2, _ := a.ResponseList[1].ReturnParam.ValueAsData()
	bData2, _ := b.ResponseList[1].ReturnParam.ValueAsData()
	if aData2.Value != bData2.Value {
		t.Errorf("t1 Failed. ResponseList[1].ReturnParam.Value get: %v, should:%v", aData2.Value, bData2.Value)
	}

	if len(src) > 0 {
		t.Errorf("t2 Failed. src should be empty. get: %v", src)
	}
}

func TestDecode_ActionResponseNextPBlock(t *testing.T) {
	x := *CreateActionResponseNextPBlock(81, 1)
	src := []byte{199, 4, 81, 0, 0, 0, 1}

	a, err := DecodeActionResponseNextPBlock(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeActionResponseNormal. err:%v", err)
	}

	if a.InvokePriority != x.InvokePriority {
		t.Errorf("t1 Failed. InvokePriority get: %v, should:%v", a.InvokePriority, x.InvokePriority)
	}

	if a.BlockNum != x.BlockNum {
		t.Errorf("t1 Failed. Result get: %v, should:%v", a.BlockNum, x.BlockNum)
	}

	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}
}

func TestDecode_ActionResponse(t *testing.T) {
	var sr ActionResponse

	// ------------------  ActionResponseNormal
	srcActionResponseNormal := []byte{199, 1, 81, 0, 1, 1, 0}
	res, e := sr.Decode(&srcActionResponseNormal)
	if e != nil {
		t.Errorf("Decode for ActionResponseNormal Failed. err:%v", e)
	}
	_, assertTrue := res.(ActionResponseNormal)
	if !assertTrue {
		t.Errorf("Decode supposed to return ActionResponseNormal instead of %v", reflect.TypeOf(res).Name())
	}

	// ------------------  ActionResponseWithPBlock
	srcActionResponseWithPBlock := []byte{199, 2, 81, 1, 0, 0, 0, 1, 5, 1, 2, 3, 4, 5}
	res, e = sr.Decode(&srcActionResponseWithPBlock)
	if e != nil {
		t.Errorf("Decode for ActionResponseWithPBlock Failed. err:%v", e)
	}
	_, assertTrue = res.(ActionResponseWithPBlock)
	if !assertTrue {
		t.Errorf("Decode supposed to return ActionResponseWithPBlock instead of %v", reflect.TypeOf(res).Name())
	}

	// ------------------  ActionResponseWithList
	srcActionResponseWithList := []byte{199, 3, 81, 1, 0, 1, 1, 0}
	res, e = sr.Decode(&srcActionResponseWithList)
	if e != nil {
		t.Errorf("Decode for ActionResponseWithList Failed. err:%v", e)
	}
	_, assertTrue = res.(ActionResponseWithList)
	if !assertTrue {
		t.Errorf("Decode supposed to return ActionResponseWithList instead of %v", reflect.TypeOf(res).Name())
	}

	// ------------------  ActionResponseNextPBlock
	srcActionResponseNextPBlock := []byte{199, 4, 81, 0, 0, 0, 1}
	res, e = sr.Decode(&srcActionResponseNextPBlock)
	if e != nil {
		t.Errorf("Decode for ActionResponseNextPBlock Failed. err:%v", e)
	}
	_, assertTrue = res.(ActionResponseNextPBlock)
	if !assertTrue {
		t.Errorf("Decode supposed to return ActionResponseNextPBlock instead of %v", reflect.TypeOf(res).Name())
	}

	// ------------------  Error test
	srcError := []byte{255, 255, 255}
	_, wow := sr.Decode(&srcError)
	if wow == nil {
		t.Errorf("Decode should've return error.")
	}
}
