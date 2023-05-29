package dlms

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/circutor-library/gosem/pkg/axdr"
)

func TestNew_SetRequestNormal(t *testing.T) {
	attrDesc := *CreateAttributeDescriptor(1, "1.0.0.3.0.255", 2)
	accsDesc := *CreateSelectiveAccessByEntryDescriptor(0, 5)
	dt := *axdr.CreateAxdrOctetString("0102030405")

	a := *CreateSetRequestNormal(81, attrDesc, &accsDesc, dt)
	t1, e := a.Encode()
	if e != nil {
		t.Errorf("t1 Encode Failed. err: %v", e)
	}
	result := []byte{193, 1, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 9, 5, 1, 2, 3, 4, 5}
	res := bytes.Compare(t1, result)
	if res != 0 {
		t.Errorf("t1 Failed. get: %d, should:%v", t1, result)
	}

	var nilAccsDesc *SelectiveAccessDescriptor
	b := *CreateSetRequestNormal(81, attrDesc, nilAccsDesc, dt)
	t2, e := b.Encode()
	if e != nil {
		t.Errorf("t2 Encode Failed. err: %v", e)
	}
	result = []byte{193, 1, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 0, 9, 5, 1, 2, 3, 4, 5}
	res = bytes.Compare(t2, result)
	if res != 0 {
		t.Errorf("t2 Failed. get: %d, should:%v", t2, result)
	}
}

func TestNew_SetRequestWithFirstDataBlock(t *testing.T) {
	attrDesc := *CreateAttributeDescriptor(1, "1.0.0.3.0.255", 2)
	accsDesc := *CreateSelectiveAccessByEntryDescriptor(0, 5)
	dt := *CreateDataBlockSA(true, 1, []byte{1, 2, 3, 4, 5})

	a := *CreateSetRequestWithFirstDataBlock(81, attrDesc, &accsDesc, dt)
	result, err := a.Encode()
	assert.NoError(t, err)
	expected := decodeHexString("C1025100010100000300FF0201020204060000000006000000051200001200000100000001050102030405")
	assert.Equal(t, expected, result)

	b := *CreateSetRequestWithFirstDataBlock(81, attrDesc, nil, dt)
	result, err = b.Encode()
	assert.NoError(t, err)
	expected = decodeHexString("C1025100010100000300FF02000100000001050102030405")
	assert.Equal(t, expected, result)
}

func TestNew_SetRequestWithDataBlock(t *testing.T) {
	dt := *CreateDataBlockSA(true, 1, []byte{1, 2, 3, 4, 5})
	a := *CreateSetRequestWithDataBlock(81, dt)
	result, e := a.Encode()
	assert.NoError(t, e)
	expected := decodeHexString("C103510100000001050102030405")
	assert.Equal(t, expected, result)
}

func TestNew_SetRequestWithList(t *testing.T) {
	sad := *CreateSelectiveAccessByEntryDescriptor(0, 5)
	a1 := *CreateAttributeDescriptorWithSelection(1, "1.0.0.3.0.255", 2, &sad)
	d1 := *axdr.CreateAxdrOctetString("0102030405")

	a := *CreateSetRequestWithList(69, []AttributeDescriptorWithSelection{a1}, []axdr.DlmsData{d1})
	t1, e := a.Encode()
	if e != nil {
		t.Errorf("t1 Encode Failed. err: %v", e)
	}
	result := []byte{193, 4, 69, 1, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 1, 9, 5, 1, 2, 3, 4, 5}
	res := bytes.Compare(t1, result)
	if res != 0 {
		t.Errorf("t1 Failed. get: %d, should:%v", t1, result)
	}

	a2 := *CreateAttributeDescriptorWithSelection(1, "0.0.8.0.0.255", 2, &sad)
	d2 := *axdr.CreateAxdrDoubleLong(69)
	b := *CreateSetRequestWithList(69, []AttributeDescriptorWithSelection{a1, a2}, []axdr.DlmsData{d1, d2})
	t2, e := b.Encode()
	if e != nil {
		t.Errorf("t2 Encode Failed. err: %v", e)
	}
	result = []byte{193, 4, 69, 2, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 0, 1, 0, 0, 8, 0, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 2, 9, 5, 1, 2, 3, 4, 5, 5, 0, 0, 0, 69}
	res = bytes.Compare(t2, result)
	if res != 0 {
		t.Errorf("t2 failed. get: %d, should:%v", t2, result)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("t3 should've panic on wrong Value")
		}
	}()
	c := *CreateSetRequestWithList(69, []AttributeDescriptorWithSelection{}, []axdr.DlmsData{d1, d2})
	c.Encode()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("t4 should've panic on wrong Value")
		}
	}()
	d := *CreateSetRequestWithList(69, []AttributeDescriptorWithSelection{a1, a2}, []axdr.DlmsData{})
	d.Encode()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("t5 should've panic on wrong Value")
		}
	}()
	ex := *CreateSetRequestWithList(69, []AttributeDescriptorWithSelection{}, []axdr.DlmsData{})
	ex.Encode()
}

func TestNew_SetRequestWithListAndFirstDataBlock(t *testing.T) {
	sad := *CreateSelectiveAccessByEntryDescriptor(0, 5)
	a1 := *CreateAttributeDescriptorWithSelection(1, "1.0.0.3.0.255", 2, &sad)
	dt := *CreateDataBlockSA(true, 1, []byte{1, 2, 3, 4, 5})

	a := *CreateSetRequestWithListAndFirstDataBlock(69, []AttributeDescriptorWithSelection{a1}, dt)
	result, err := a.Encode()
	assert.NoError(t, err)
	expected := decodeHexString("C105450100010100000300FF0201020204060000000006000000051200001200000100000001050102030405")
	assert.Equal(t, expected, result)

	a2 := *CreateAttributeDescriptorWithSelection(1, "0.0.8.0.0.255", 2, &sad)
	b := *CreateSetRequestWithListAndFirstDataBlock(69, []AttributeDescriptorWithSelection{a1, a2}, dt)
	result, err = b.Encode()
	assert.NoError(t, err)
	expected = decodeHexString("C105450200010100000300FF02010202040600000000060000000512000012000000010000080000FF0201020204060000000006000000051200001200000100000001050102030405")
	assert.Equal(t, expected, result)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("t3 should've panic on wrong Value")
		}
	}()
	c := *CreateSetRequestWithListAndFirstDataBlock(69, []AttributeDescriptorWithSelection{}, dt)
	c.Encode()
}

func TestDecode_SetRequestNormal(t *testing.T) {
	src := []byte{193, 1, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 9, 5, 1, 2, 3, 4, 5}
	a, err := DecodeSetRequestNormal(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeSetRequestNormal. err:%v", err)
	}

	attrDesc := *CreateAttributeDescriptor(1, "1.0.0.3.0.255", 2)
	accsDesc := *CreateSelectiveAccessByEntryDescriptor(0, 5)
	dt := *axdr.CreateAxdrOctetString("0102030405")
	b := *CreateSetRequestNormal(81, attrDesc, &accsDesc, dt)

	if a.InvokePriority != b.InvokePriority {
		t.Errorf("t1 Failed. InvokePriority get: %v, should:%v", a.InvokePriority, b.InvokePriority)
	}
	if a.AttributeInfo.ClassID != b.AttributeInfo.ClassID {
		t.Errorf("t1 Failed. AttributeInfo.ClassID get: %v, should:%v", a.AttributeInfo.ClassID, b.AttributeInfo.ClassID)
	}
	res := bytes.Compare(a.AttributeInfo.InstanceID.Bytes(), b.AttributeInfo.InstanceID.Bytes())
	if res != 0 {
		t.Errorf("t1 Failed. AttributeInfo.InstanceID get: %v, should:%v", a.AttributeInfo.InstanceID.Bytes(), b.AttributeInfo.InstanceID.Bytes())
	}
	if a.AttributeInfo.AttributeID != b.AttributeInfo.AttributeID {
		t.Errorf("t1 Failed. AttributeInfo.AttributeID get: %v, should:%v", a.AttributeInfo.AttributeID, b.AttributeInfo.AttributeID)
	}
	if a.SelectiveAccessInfo.AccessSelector != b.SelectiveAccessInfo.AccessSelector {
		t.Errorf("t1 Failed. SelectiveAccessInfo.AccessSelector get: %v, should:%v", a.SelectiveAccessInfo.AccessSelector, b.SelectiveAccessInfo.AccessSelector)
	}
	aByte, _ := a.SelectiveAccessInfo.AccessParameter.Encode()
	bByte, _ := b.SelectiveAccessInfo.AccessParameter.Encode()
	res = bytes.Compare(aByte, bByte)
	if res != 0 {
		t.Errorf("t1 Failed. SelectiveAccessInfo.AccessParameter get: %v, should:%v", aByte, bByte)
	}
	if a.Value.Tag != b.Value.Tag {
		t.Errorf("t1 Failed. Value.Tag get: %v, should:%v", a.Value.Tag, b.Value.Tag)
	}
	if a.Value.Value != b.Value.Value {
		t.Errorf("t1 Failed. Value.Value get: %v, should:%v", a.Value.Value, b.Value.Value)
	}
	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}

	// ------------------ t2 without SelectiveAccessDescriptor

	src = []byte{193, 1, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 0, 9, 5, 1, 2, 3, 4, 5}
	a, err = DecodeSetRequestNormal(&src)

	if err != nil {
		t.Errorf("t2 Failed to DecodeGetRequestNormal. err:%v", err)
	}

	var nilAccsDesc *SelectiveAccessDescriptor
	b = *CreateSetRequestNormal(81, attrDesc, nilAccsDesc, dt)

	if a.SelectiveAccessInfo != nilAccsDesc {
		t.Errorf("t2 Failed. SelectiveAccessInfo.AccessSelector should be nil get: %v", a.SelectiveAccessInfo)
	}
	if len(src) > 0 {
		t.Errorf("t2 Failed. src should be empty. get: %v", src)
	}
}

func TestDecode_SetRequestWithFirstDataBlock(t *testing.T) {
	src := []byte{193, 2, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 1, 0, 0, 0, 1, 5, 1, 2, 3, 4, 5}
	a, err := DecodeSetRequestWithFirstDataBlock(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeSetRequestWithFirstDataBlock. err:%v", err)
	}

	attrDesc := *CreateAttributeDescriptor(1, "1.0.0.3.0.255", 2)
	accsDesc := *CreateSelectiveAccessByEntryDescriptor(0, 5)
	dt := *CreateDataBlockSA(true, 1, []byte{1, 2, 3, 4, 5})
	b := *CreateSetRequestWithFirstDataBlock(81, attrDesc, &accsDesc, dt)

	if a.InvokePriority != b.InvokePriority {
		t.Errorf("t1 Failed. InvokePriority get: %v, should:%v", a.InvokePriority, b.InvokePriority)
	}
	if a.AttributeInfo.ClassID != b.AttributeInfo.ClassID {
		t.Errorf("t1 Failed. AttributeInfo.ClassID get: %v, should:%v", a.AttributeInfo.ClassID, b.AttributeInfo.ClassID)
	}
	res := bytes.Compare(a.AttributeInfo.InstanceID.Bytes(), b.AttributeInfo.InstanceID.Bytes())
	if res != 0 {
		t.Errorf("t1 Failed. AttributeInfo.InstanceID get: %v, should:%v", a.AttributeInfo.InstanceID.Bytes(), b.AttributeInfo.InstanceID.Bytes())
	}
	if a.AttributeInfo.AttributeID != b.AttributeInfo.AttributeID {
		t.Errorf("t1 Failed. AttributeInfo.AttributeID get: %v, should:%v", a.AttributeInfo.AttributeID, b.AttributeInfo.AttributeID)
	}
	if a.SelectiveAccessInfo.AccessSelector != b.SelectiveAccessInfo.AccessSelector {
		t.Errorf("t1 Failed. SelectiveAccessInfo.AccessSelector get: %v, should:%v", a.SelectiveAccessInfo.AccessSelector, b.SelectiveAccessInfo.AccessSelector)
	}
	aByte, _ := a.SelectiveAccessInfo.AccessParameter.Encode()
	bByte, _ := b.SelectiveAccessInfo.AccessParameter.Encode()
	res = bytes.Compare(aByte, bByte)
	if res != 0 {
		t.Errorf("t1 Failed. SelectiveAccessInfo.AccessParameter get: %v, should:%v", aByte, bByte)
	}
	if a.DataBlock.LastBlock != b.DataBlock.LastBlock {
		t.Errorf("t1 Failed. DataBlock.LastBlock get: %v, should:%v", a.DataBlock.LastBlock, b.DataBlock.LastBlock)
	}
	if a.DataBlock.BlockNumber != b.DataBlock.BlockNumber {
		t.Errorf("t1 Failed. DataBlock.BlockNumber get: %v, should:%v", a.DataBlock.BlockNumber, b.DataBlock.BlockNumber)
	}
	res = bytes.Compare(a.DataBlock.Raw, b.DataBlock.Raw)
	if res != 0 {
		t.Errorf("t1 Failed. DataBlock.Raw get: %v, should:%v", a.DataBlock.Raw, b.DataBlock.Raw)
	}
	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}

	// ------------------ t2 without SelectiveAccessDescriptor

	src = []byte{193, 2, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 0, 1, 0, 0, 0, 1, 5, 1, 2, 3, 4, 5}
	a, err = DecodeSetRequestWithFirstDataBlock(&src)

	if err != nil {
		t.Errorf("t2 Failed to DecodeSetRequestWithFirstDataBlock. err:%v", err)
	}

	var nilAccsDesc *SelectiveAccessDescriptor
	b = *CreateSetRequestWithFirstDataBlock(81, attrDesc, nilAccsDesc, dt)

	if a.SelectiveAccessInfo != nilAccsDesc {
		t.Errorf("t2 Failed. SelectiveAccessInfo.AccessSelector should be nil get: %v", a.SelectiveAccessInfo)
	}
	if len(src) > 0 {
		t.Errorf("t2 Failed. src should be empty. get: %v", src)
	}
}

func TestDecode_SetRequestWithDataBlock(t *testing.T) {
	src := []byte{193, 3, 81, 1, 0, 0, 0, 1, 5, 1, 2, 3, 4, 5}
	a, err := DecodeSetRequestWithDataBlock(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeSetRequestWithDataBlock. err:%v", err)
	}

	dt := *CreateDataBlockSA(true, 1, []byte{1, 2, 3, 4, 5})
	b := *CreateSetRequestWithDataBlock(81, dt)

	if a.InvokePriority != b.InvokePriority {
		t.Errorf("t1 Failed. InvokePriority get: %v, should:%v", a.InvokePriority, b.InvokePriority)
	}
	if a.DataBlock.LastBlock != b.DataBlock.LastBlock {
		t.Errorf("t1 Failed. DataBlock.LastBlock get: %v, should:%v", a.DataBlock.LastBlock, b.DataBlock.LastBlock)
	}
	if a.DataBlock.BlockNumber != b.DataBlock.BlockNumber {
		t.Errorf("t1 Failed. DataBlock.BlockNumber get: %v, should:%v", a.DataBlock.BlockNumber, b.DataBlock.BlockNumber)
	}
	res := bytes.Compare(a.DataBlock.Raw, b.DataBlock.Raw)
	if res != 0 {
		t.Errorf("t1 Failed. DataBlock.Raw get: %v, should:%v", a.DataBlock.Raw, b.DataBlock.Raw)
	}
	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}
}

func TestDecode_SetRequestWithList(t *testing.T) {
	src := []byte{193, 4, 69, 1, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 1, 9, 5, 1, 2, 3, 4, 5}
	a, err := DecodeSetRequestWithList(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeSetRequestWithList. err:%v", err)
	}

	sad := *CreateSelectiveAccessByEntryDescriptor(0, 5)
	a1 := *CreateAttributeDescriptorWithSelection(1, "1.0.0.3.0.255", 2, &sad)
	d1 := *axdr.CreateAxdrOctetString("0102030405")
	b := *CreateSetRequestWithList(69, []AttributeDescriptorWithSelection{a1}, []axdr.DlmsData{d1})

	if a.InvokePriority != b.InvokePriority {
		t.Errorf("t1 Failed. InvokePriority get: %v, should:%v", a.InvokePriority, b.InvokePriority)
	}
	if len(a.AttributeInfoList) != len(b.AttributeInfoList) {
		t.Errorf("t1 Failed. AttributeInfoList count get: %v, should:%v", len(a.AttributeInfoList), len(b.AttributeInfoList))
	}
	if a.AttributeCount != b.AttributeCount {
		t.Errorf("t1 Failed. AttributeCount get: %v, should:%v", a.AttributeCount, b.AttributeCount)
	}
	aDescObis := a.AttributeInfoList[0].InstanceID.String()
	bDescObis := b.AttributeInfoList[0].InstanceID.String()
	if aDescObis != bDescObis {
		t.Errorf("t1 Failed. AttributeInfoList[0].InstanceID get: %v, should:%v", aDescObis, bDescObis)
	}
	if len(a.ValueList) != len(b.ValueList) {
		t.Errorf("t1 Failed. ValueList count get: %v, should:%v", len(a.ValueList), len(b.ValueList))
	}
	if a.ValueCount != b.ValueCount {
		t.Errorf("t1 Failed. ValueCount get: %v, should:%v", a.ValueCount, b.ValueCount)
	}
	aDataTag := a.ValueList[0].Tag
	bDataTag := b.ValueList[0].Tag
	if aDataTag != bDataTag {
		t.Errorf("t1 Failed. ValueList[0].Tag get: %v, should:%v", aDataTag, bDataTag)
	}

	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}

	// ---------------------- with 2 AttributeDescriptor
	src = []byte{193, 4, 69, 2, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 0, 1, 0, 0, 8, 0, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 2, 9, 5, 1, 2, 3, 4, 5, 5, 0, 0, 0, 69}
	a, err = DecodeSetRequestWithList(&src)
	if err != nil {
		t.Errorf("t2 Failed to DecodeActionRequestWithList. err:%v", err)
	}

	a2 := *CreateAttributeDescriptorWithSelection(1, "0.0.8.0.0.255", 2, &sad)
	d2 := *axdr.CreateAxdrDoubleLong(69)
	b = *CreateSetRequestWithList(69, []AttributeDescriptorWithSelection{a1, a2}, []axdr.DlmsData{d1, d2})

	if len(a.AttributeInfoList) != len(b.AttributeInfoList) {
		t.Errorf("t2 Failed. AttributeInfoList count get: %v, should:%v", len(a.AttributeInfoList), len(b.AttributeInfoList))
	}
	if a.AttributeCount != b.AttributeCount {
		t.Errorf("t2 Failed. AttributeCount get: %v, should:%v", a.AttributeCount, b.AttributeCount)
	}
	aDescObis = a.AttributeInfoList[1].InstanceID.String()
	bDescObis = b.AttributeInfoList[1].InstanceID.String()
	if aDescObis != bDescObis {
		t.Errorf("t2 Failed. AttributeInfoList[1].InstanceID get: %v, should:%v", aDescObis, bDescObis)
	}
	if len(a.ValueList) != len(b.ValueList) {
		t.Errorf("t2 Failed. ValueList count get: %v, should:%v", len(a.ValueList), len(b.ValueList))
	}
	if a.ValueCount != b.ValueCount {
		t.Errorf("t2 Failed. ValueCount get: %v, should:%v", a.ValueCount, b.ValueCount)
	}
	aDataTag = a.ValueList[1].Tag
	bDataTag = b.ValueList[1].Tag
	if aDataTag != bDataTag {
		t.Errorf("t2 Failed. ValueList[1].Tag get: %v, should:%v", aDataTag, bDataTag)
	}

	if len(src) > 0 {
		t.Errorf("t2 Failed. src should be empty. get: %v", src)
	}
}

func TestDecode_SetRequestWithListAndFirstDataBlock(t *testing.T) {
	src := []byte{193, 5, 69, 1, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 1, 0, 0, 0, 1, 5, 1, 2, 3, 4, 5}
	a, err := DecodeSetRequestWithListAndFirstDataBlock(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeSetRequestWithListAndFirstDataBlock. err:%v", err)
	}

	sad := *CreateSelectiveAccessByEntryDescriptor(0, 5)
	a1 := *CreateAttributeDescriptorWithSelection(1, "1.0.0.3.0.255", 2, &sad)
	dt := *CreateDataBlockSA(true, 1, []byte{1, 2, 3, 4, 5})
	b := *CreateSetRequestWithListAndFirstDataBlock(69, []AttributeDescriptorWithSelection{a1}, dt)

	if a.InvokePriority != b.InvokePriority {
		t.Errorf("t1 Failed. InvokePriority get: %v, should:%v", a.InvokePriority, b.InvokePriority)
	}
	if len(a.AttributeInfoList) != len(b.AttributeInfoList) {
		t.Errorf("t1 Failed. AttributeInfoList count get: %v, should:%v", len(a.AttributeInfoList), len(b.AttributeInfoList))
	}
	if a.AttributeCount != b.AttributeCount {
		t.Errorf("t1 Failed. AttributeCount get: %v, should:%v", a.AttributeCount, b.AttributeCount)
	}
	aDescObis := a.AttributeInfoList[0].InstanceID.String()
	bDescObis := b.AttributeInfoList[0].InstanceID.String()
	if aDescObis != bDescObis {
		t.Errorf("t1 Failed. AttributeInfoList[0].InstanceID get: %v, should:%v", aDescObis, bDescObis)
	}
	if a.DataBlock.LastBlock != b.DataBlock.LastBlock {
		t.Errorf("t1 Failed. DataBlock.LastBlock get: %v, should:%v", a.DataBlock.LastBlock, b.DataBlock.LastBlock)
	}
	if a.DataBlock.BlockNumber != b.DataBlock.BlockNumber {
		t.Errorf("t1 Failed. DataBlock.BlockNumber get: %v, should:%v", a.DataBlock.BlockNumber, b.DataBlock.BlockNumber)
	}
	res := bytes.Compare(a.DataBlock.Raw, b.DataBlock.Raw)
	if res != 0 {
		t.Errorf("t1 Failed. DataBlock.Raw get: %v, should:%v", a.DataBlock.Raw, b.DataBlock.Raw)
	}

	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}

	// ---------------------- with 2 AttributeDescriptor
	src = []byte{193, 5, 69, 2, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 0, 1, 0, 0, 8, 0, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 1, 0, 0, 0, 1, 5, 1, 2, 3, 4, 5}
	a, err = DecodeSetRequestWithListAndFirstDataBlock(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeSetRequestWithListAndFirstDataBlock. err:%v", err)
	}

	a2 := *CreateAttributeDescriptorWithSelection(1, "0.0.8.0.0.255", 2, &sad)
	b = *CreateSetRequestWithListAndFirstDataBlock(69, []AttributeDescriptorWithSelection{a1, a2}, dt)

	if len(a.AttributeInfoList) != len(b.AttributeInfoList) {
		t.Errorf("t2 Failed. AttributeInfoList count get: %v, should:%v", len(a.AttributeInfoList), len(b.AttributeInfoList))
	}
	if a.AttributeCount != b.AttributeCount {
		t.Errorf("t2 Failed. AttributeCount get: %v, should:%v", a.AttributeCount, b.AttributeCount)
	}
	aDescObis = a.AttributeInfoList[1].InstanceID.String()
	bDescObis = b.AttributeInfoList[1].InstanceID.String()
	if aDescObis != bDescObis {
		t.Errorf("t2 Failed. AttributeInfoList[1].InstanceID get: %v, should:%v", aDescObis, bDescObis)
	}

	if len(src) > 0 {
		t.Errorf("t2 Failed. src should be empty. get: %v", src)
	}
}

func TestDecode_SetRequest(t *testing.T) {
	var sr SetRequest

	// ------------------  SetRequestNormal
	src := []byte{193, 1, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 9, 5, 1, 2, 3, 4, 5}
	res, e := sr.Decode(&src)
	if e != nil {
		t.Errorf("Decode for SetRequestNormal Failed. err:%v", e)
	}
	_, assertTrue := res.(SetRequestNormal)
	if !assertTrue {
		t.Errorf("Decode supposed to return SetRequestNormal instead of %v", reflect.TypeOf(res).Name())
	}

	// ------------------  SetRequestWithFirstDataBlock
	src = []byte{193, 2, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 1, 0, 0, 0, 1, 5, 1, 2, 3, 4, 5}
	res, e = sr.Decode(&src)
	if e != nil {
		t.Errorf("Decode for SetRequestWithFirstDataBlock Failed. err:%v", e)
	}
	_, assertTrue = res.(SetRequestWithFirstDataBlock)
	if !assertTrue {
		t.Errorf("Decode supposed to return SetRequestWithFirstDataBlock instead of %v", reflect.TypeOf(res).Name())
	}

	// ------------------  SetRequestWithDataBlock
	src = []byte{193, 3, 81, 1, 0, 0, 0, 1, 5, 1, 2, 3, 4, 5}
	res, e = sr.Decode(&src)
	if e != nil {
		t.Errorf("Decode for SetRequestWithDataBlock Failed. err:%v", e)
	}
	_, assertTrue = res.(SetRequestWithDataBlock)
	if !assertTrue {
		t.Errorf("Decode supposed to return SetRequestWithDataBlock instead of %v", reflect.TypeOf(res).Name())
	}

	// ------------------  SetRequestWithList
	src = []byte{193, 4, 69, 1, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 1, 9, 5, 1, 2, 3, 4, 5}
	res, e = sr.Decode(&src)
	if e != nil {
		t.Errorf("Decode for SetRequestWithList Failed. err:%v", e)
	}
	_, assertTrue = res.(SetRequestWithList)
	if !assertTrue {
		t.Errorf("Decode supposed to return SetRequestWithList instead of %v", reflect.TypeOf(res).Name())
	}

	// ------------------  SetRequestWithListAndFirstDataBlock
	src = []byte{193, 5, 69, 1, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 1, 0, 0, 0, 1, 5, 1, 2, 3, 4, 5}
	res, e = sr.Decode(&src)
	if e != nil {
		t.Errorf("Decode for SetRequestWithListAndFirstDataBlock Failed. err:%v", e)
	}
	_, assertTrue = res.(SetRequestWithListAndFirstDataBlock)
	if !assertTrue {
		t.Errorf("Decode supposed to return SetRequestWithListAndFirstDataBlock instead of %v", reflect.TypeOf(res).Name())
	}

	// ------------------  Error test
	srcError := []byte{255, 255, 255}
	_, wow := sr.Decode(&srcError)
	if wow == nil {
		t.Errorf("Decode should've return error.")
	}
}
