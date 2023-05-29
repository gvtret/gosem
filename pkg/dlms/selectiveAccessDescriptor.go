package dlms

import (
	"bytes"
	"time"

	"gitlab.com/circutor-library/gosem/pkg/axdr"
)

type accessSelector uint8

const (
	AccessSelectorRange accessSelector = 0x1
	AccessSelectorEntry accessSelector = 0x2
)

// Value will return primitive value of the target.
// This is used for comparing with non custom typed object
func (s accessSelector) Value() uint8 {
	return uint8(s)
}

type SelectiveAccessDescriptor struct {
	AccessSelector  accessSelector
	AccessParameter axdr.DlmsData
}

func CreateSelectiveAccessByRangeDescriptor(from time.Time, to time.Time, values []AttributeDescriptor) *SelectiveAccessDescriptor {
	restrictingObject := createAttributeDescriptorWithIndex(8, "0.0.1.0.0.255", 2, 0)

	fromValue := axdr.CreateAxdrOctetString(from)
	toValue := axdr.CreateAxdrOctetString(to)

	selected := make([]*axdr.DlmsData, 0, len(values))
	for _, v := range values {
		selected = append(selected, createAttributeDescriptorWithIndex(v.ClassID, v.InstanceID.String(), v.AttributeID, 0))
	}
	selectedValues := axdr.CreateAxdrArray(selected)

	rangeDescriptor := *axdr.CreateAxdrStructure([]*axdr.DlmsData{restrictingObject, fromValue, toValue, selectedValues})

	return &SelectiveAccessDescriptor{AccessSelector: AccessSelectorRange, AccessParameter: rangeDescriptor}
}

func CreateSelectiveAccessByEntryDescriptor(from uint32, to uint32) *SelectiveAccessDescriptor {
	fromEntry := *axdr.CreateAxdrDoubleLongUnsigned(from)
	toEntry := *axdr.CreateAxdrDoubleLongUnsigned(to)

	fromSelectedValue := *axdr.CreateAxdrLongUnsigned(0)
	toSelectedValue := *axdr.CreateAxdrLongUnsigned(0)

	entryDescriptor := *axdr.CreateAxdrStructure([]*axdr.DlmsData{&fromEntry, &toEntry, &fromSelectedValue, &toSelectedValue})

	return &SelectiveAccessDescriptor{AccessSelector: AccessSelectorEntry, AccessParameter: entryDescriptor}
}

func createAttributeDescriptorWithIndex(class uint16, obis string, attribute int8, index uint16) *axdr.DlmsData {
	classID := *axdr.CreateAxdrLongUnsigned(class)
	obisCode := *axdr.CreateAxdrOctetString(obis)
	attributeID := *axdr.CreateAxdrInteger(attribute)
	dataIdx := *axdr.CreateAxdrLongUnsigned(index)

	return axdr.CreateAxdrStructure([]*axdr.DlmsData{&classID, &obisCode, &attributeID, &dataIdx})
}

func (s SelectiveAccessDescriptor) Encode() (out []byte, err error) {
	var buf bytes.Buffer
	buf.WriteByte(s.AccessSelector.Value())
	val, e := s.AccessParameter.Encode()
	if e != nil {
		err = e
		return
	}
	buf.Write(val)

	out = buf.Bytes()
	return
}

func DecodeSelectiveAccessDescriptor(ori *[]byte) (out SelectiveAccessDescriptor, err error) {
	src := *ori

	if src[0] == AccessSelectorRange.Value() {
		out.AccessSelector = AccessSelectorRange
	} else {
		out.AccessSelector = AccessSelectorEntry
	}
	src = src[1:] // remove access-selector byte

	axdrDecoder := *axdr.NewDataDecoder(&src)
	out.AccessParameter, err = axdrDecoder.Decode(&src)
	if err != nil {
		return
	}

	(*ori) = (*ori)[len((*ori))-len(src):]
	return
}
