package dlms

import (
	"encoding/binary"
	"fmt"

	"github.com/Circutor/gosem/pkg/axdr"
)

type AttributeDescriptorWithSelection struct {
	ClassID          uint16
	InstanceID       Obis
	AttributeID      int8
	AccessDescriptor *SelectiveAccessDescriptor
}

// CreateAttributeDescriptorWithSelection will create AttributeDescriptorWithSelection object
// SelectiveAccessDescriptor is allowed to be nil value therefore pointer
func CreateAttributeDescriptorWithSelection(c uint16, i string, a int8, sad *SelectiveAccessDescriptor) *AttributeDescriptorWithSelection {
	ob := *CreateObis(i)

	return &AttributeDescriptorWithSelection{ClassID: c, InstanceID: ob, AttributeID: a, AccessDescriptor: sad}
}

func (ad AttributeDescriptorWithSelection) Encode() (out []byte, err error) {
	var output []byte
	var c [2]byte
	binary.BigEndian.PutUint16(c[:], ad.ClassID)
	output = append(output, c[:]...)
	output = append(output, ad.InstanceID.Bytes()...)
	output = append(output, byte(ad.AttributeID))
	if ad.AccessDescriptor == nil {
		output = append(output, 0)
	} else {
		output = append(output, 1)
		val, e := ad.AccessDescriptor.Encode()
		if e != nil {
			err = e
			return
		}
		output = append(output, val...)
	}

	out = output
	return
}

func DecodeAttributeDescriptorWithSelection(ori *[]byte) (out AttributeDescriptorWithSelection, err error) {
	src := *ori

	if len(src) < 11 {
		err = fmt.Errorf("byte slice length must be at least 11 bytes")
		return
	}

	_, out.ClassID, err = axdr.DecodeLongUnsigned(&src)
	if err != nil {
		return
	}

	out.InstanceID, err = DecodeObis(&src)
	if err != nil {
		return
	}

	out.AttributeID = int8(src[0])
	haveAccDesc := src[1]
	src = src[2:]

	if haveAccDesc == 0x0 {
		var nilAccDesc *SelectiveAccessDescriptor
		out.AccessDescriptor = nilAccDesc
	} else {
		accDesc, errAcc := DecodeSelectiveAccessDescriptor(&src)
		if errAcc != nil {
			err = errAcc
			return
		}
		out.AccessDescriptor = &accDesc
	}

	(*ori) = (*ori)[len((*ori))-len(src):]
	return
}
