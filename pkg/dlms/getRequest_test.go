package dlms

import (
	"bytes"
	"reflect"
	"testing"
)

func TestNewGetRequestNormal(t *testing.T) {
	attrDesc := *CreateAttributeDescriptor(1, "1.0.0.3.0.255", 2)
	accsDesc := *CreateSelectiveAccessByEntryDescriptor(0, 5)

	a := *CreateGetRequestNormal(81, attrDesc, &accsDesc)
	t1, e := a.Encode()
	if e != nil {
		t.Errorf("t1 Encode Failed. err: %v", e)
	}
	result := []byte{192, 1, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0}
	res := bytes.Compare(t1, result)
	if res != 0 {
		t.Errorf("t1 Failed. get: %d, should:%v", t1, result)
	}

	var nilAccsDesc *SelectiveAccessDescriptor
	b := *CreateGetRequestNormal(81, attrDesc, nilAccsDesc)
	t2, e := b.Encode()
	if e != nil {
		t.Errorf("t2 Encode Failed. err: %v", e)
	}
	result = []byte{192, 1, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 0}
	res = bytes.Compare(t2, result)
	if res != 0 {
		t.Errorf("t2 failed. get: %d, should:%v", t2, result)
	}
}

func TestNewGetRequestNext(t *testing.T) {
	a := *CreateGetRequestNext(81, 2)
	t1, e := a.Encode()
	if e != nil {
		t.Errorf("t1 Encode Failed. err: %v", e)
	}
	result := []byte{192, 2, 81, 0, 0, 0, 2}
	res := bytes.Compare(t1, result)
	if res != 0 {
		t.Errorf("t1 Failed. get: %d, should:%v", t1, result)
	}
}

func TestNewGetRequestWithList(t *testing.T) {
	sad := *CreateSelectiveAccessByEntryDescriptor(0, 5)
	a1 := *CreateAttributeDescriptorWithSelection(1, "1.0.0.3.0.255", 2, &sad)

	a := *CreateGetRequestWithList(69, []AttributeDescriptorWithSelection{a1})
	t1, e := a.Encode()
	if e != nil {
		t.Errorf("t1 Encode Failed. err: %v", e)
	}
	result := []byte{192, 3, 69, 1, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0}
	res := bytes.Compare(t1, result)
	if res != 0 {
		t.Errorf("t1 Failed. get: %d, should:%v", t1, result)
	}

	a2 := *CreateAttributeDescriptorWithSelection(1, "0.0.8.0.0.255", 2, &sad)
	b := *CreateGetRequestWithList(69, []AttributeDescriptorWithSelection{a1, a2})
	t2, e := b.Encode()
	if e != nil {
		t.Errorf("t2 Encode Failed. err: %v", e)
	}
	result = []byte{192, 3, 69, 2, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 0, 1, 0, 0, 8, 0, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0}
	res = bytes.Compare(t2, result)
	if res != 0 {
		t.Errorf("t2 failed. get: %d, should:%v", t2, result)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("t3 should've panic on wrong Value")
		}
	}()
	c := *CreateGetRequestWithList(69, []AttributeDescriptorWithSelection{})
	c.Encode()
}

func TestDecode_GetRequestNormal(t *testing.T) {
	src := []byte{192, 1, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0}
	a, err := DecodeGetRequestNormal(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeGetRequestNormal. err:%v", err)
	}

	attrDesc := *CreateAttributeDescriptor(1, "1.0.0.3.0.255", 2)
	accsDesc := *CreateSelectiveAccessByEntryDescriptor(0, 5)
	b := *CreateGetRequestNormal(81, attrDesc, &accsDesc)

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
	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}

	// ------------------ t2 without SelectiveAccessDescriptor

	src = []byte{192, 1, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 0}
	a, err = DecodeGetRequestNormal(&src)

	if err != nil {
		t.Errorf("t1 Failed to DecodeGetRequestNormal. err:%v", err)
	}

	attrDesc = *CreateAttributeDescriptor(1, "1.0.0.3.0.255", 2)
	var nilAccsDesc *SelectiveAccessDescriptor
	b = *CreateGetRequestNormal(81, attrDesc, nilAccsDesc)

	if a.InvokePriority != b.InvokePriority {
		t.Errorf("t1 Failed. InvokePriority get: %v, should:%v", a.InvokePriority, b.InvokePriority)
	}
	if a.AttributeInfo.ClassID != b.AttributeInfo.ClassID {
		t.Errorf("t1 Failed. AttributeInfo.ClassID get: %v, should:%v", a.AttributeInfo.ClassID, b.AttributeInfo.ClassID)
	}
	res = bytes.Compare(a.AttributeInfo.InstanceID.Bytes(), b.AttributeInfo.InstanceID.Bytes())
	if res != 0 {
		t.Errorf("t1 Failed. AttributeInfo.InstanceID get: %v, should:%v", a.AttributeInfo.InstanceID.Bytes(), b.AttributeInfo.InstanceID.Bytes())
	}
	if a.AttributeInfo.AttributeID != b.AttributeInfo.AttributeID {
		t.Errorf("t1 Failed. AttributeInfo.AttributeID get: %v, should:%v", a.AttributeInfo.AttributeID, b.AttributeInfo.AttributeID)
	}
	if a.SelectiveAccessInfo != nilAccsDesc {
		t.Errorf("t1 Failed. SelectiveAccessInfo.AccessSelector should be nil get: %v", a.SelectiveAccessInfo)
	}
	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}
}

func TestDecode_GetRequestNext(t *testing.T) {
	src := []byte{192, 2, 81, 0, 0, 0, 2}
	a, err := DecodeGetRequestNext(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeGetRequestNext. err:%v", err)
	}

	b := *CreateGetRequestNext(81, 2)

	if a.InvokePriority != b.InvokePriority {
		t.Errorf("t1 Failed. InvokePriority get: %v, should:%v", a.InvokePriority, b.InvokePriority)
	}
	if a.BlockNum != b.BlockNum {
		t.Errorf("t1 Failed. BlockNum get: %v, should:%v", a.BlockNum, b.BlockNum)
	}
	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}
}

func TestDecode_GetRequestWithList(t *testing.T) {
	src := []byte{192, 3, 69, 1, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0}
	a, err := DecodeGetRequestWithList(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeGetRequestWithList. err:%v", err)
	}

	sad := *CreateSelectiveAccessByEntryDescriptor(0, 5)
	a1 := *CreateAttributeDescriptorWithSelection(1, "1.0.0.3.0.255", 2, &sad)
	b := *CreateGetRequestWithList(69, []AttributeDescriptorWithSelection{a1})

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
	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}

	// ---------------------- with 2 AttributeDescriptor
	src = []byte{192, 3, 69, 2, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0, 0, 1, 0, 0, 8, 0, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0}
	a, err = DecodeGetRequestWithList(&src)
	if err != nil {
		t.Errorf("t1 Failed to DecodeGetRequestWithList. err:%v", err)
	}

	a2 := *CreateAttributeDescriptorWithSelection(1, "0.0.8.0.0.255", 2, &sad)
	b = *CreateGetRequestWithList(69, []AttributeDescriptorWithSelection{a1, a2})

	if a.InvokePriority != b.InvokePriority {
		t.Errorf("t1 Failed. InvokePriority get: %v, should:%v", a.InvokePriority, b.InvokePriority)
	}
	if len(a.AttributeInfoList) != len(b.AttributeInfoList) {
		t.Errorf("t1 Failed. AttributeInfoList count get: %v, should:%v", len(a.AttributeInfoList), len(b.AttributeInfoList))
	}
	if a.AttributeCount != b.AttributeCount {
		t.Errorf("t1 Failed. AttributeCount get: %v, should:%v", a.AttributeCount, b.AttributeCount)
	}
	aDescObis = a.AttributeInfoList[0].InstanceID.String()
	bDescObis = b.AttributeInfoList[0].InstanceID.String()
	if aDescObis != bDescObis {
		t.Errorf("t1 Failed. AttributeInfoList[0].InstanceID get: %v, should:%v", aDescObis, bDescObis)
	}
	aDescObis = a.AttributeInfoList[1].InstanceID.String()
	bDescObis = b.AttributeInfoList[1].InstanceID.String()
	if aDescObis != bDescObis {
		t.Errorf("t1 Failed. AttributeInfoList[1].InstanceID get: %v, should:%v", aDescObis, bDescObis)
	}
	if len(src) > 0 {
		t.Errorf("t1 Failed. src should be empty. get: %v", src)
	}
}

func TestDecode_GetRequest(t *testing.T) {
	var gr GetRequest

	// ------------------  GetRequestNormal
	srcNormal := []byte{192, 1, 81, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0}
	a, e1 := gr.Decode(&srcNormal)
	if e1 != nil {
		t.Errorf("Decode for GetRequestNormal Failed. err:%v", e1)
	}
	_, assertGetRequestNormal := a.(GetRequestNormal)
	if !assertGetRequestNormal {
		t.Errorf("Decode supposed to return %v instead of %v", reflect.TypeOf(GetRequestNormal{}).Name(), reflect.TypeOf(a).Name())
	}

	// ------------------  GetRequestNext
	srcNext := []byte{192, 2, 81, 0, 0, 0, 2}
	b, e2 := gr.Decode(&srcNext)
	if e2 != nil {
		t.Errorf("Decode for GetRequestNext Failed. err:%v", e2)
	}
	_, assertGetRequestNext := b.(GetRequestNext)
	if !assertGetRequestNext {
		t.Errorf("Decode supposed to return GetRequestNext instead of %v", reflect.TypeOf(b).Name())
	}

	// ------------------  GetRequestWithList
	srcWithList := []byte{192, 3, 69, 1, 0, 1, 1, 0, 0, 3, 0, 255, 2, 1, 2, 2, 4, 6, 0, 0, 0, 0, 6, 0, 0, 0, 5, 18, 0, 0, 18, 0, 0}
	c, e3 := gr.Decode(&srcWithList)
	if e3 != nil {
		t.Errorf("Decode for GetRequestWithList Failed. err:%v", e3)
	}
	_, assertGetRequestWithList := c.(GetRequestWithList)
	if !assertGetRequestWithList {
		t.Errorf("Decode supposed to return GetRequestWithList instead of %v", reflect.TypeOf(c).Name())
	}

	// ------------------  Error test
	srcError := []byte{255, 255, 255}
	_, wow := gr.Decode(&srcError)
	if wow == nil {
		t.Errorf("Decode should've return error.")
	}
}
