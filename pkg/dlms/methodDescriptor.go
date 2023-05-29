package dlms

import (
	"encoding/binary"
	"fmt"

	"gitlab.com/circutor-library/gosem/pkg/axdr"
)

type MethodDescriptor struct {
	ClassID    uint16
	InstanceID Obis
	MethodID   int8
}

func CreateMethodDescriptor(c uint16, i string, a int8) *MethodDescriptor {
	ob := *CreateObis(i)

	return &MethodDescriptor{ClassID: c, InstanceID: ob, MethodID: a}
}

func (ad MethodDescriptor) Encode() (out []byte, err error) {
	var output []byte
	var c [2]byte
	binary.BigEndian.PutUint16(c[:], ad.ClassID)
	output = append(output, c[:]...)
	output = append(output, ad.InstanceID.Bytes()...)
	output = append(output, byte(ad.MethodID))

	out = output
	return
}

func (ad MethodDescriptor) String() string {
	return fmt.Sprintf("{ %d, %s, %d }", ad.ClassID, ad.InstanceID.String(), ad.MethodID)
}

func DecodeMethodDescriptor(ori *[]byte) (out MethodDescriptor, err error) {
	src := *ori

	if len(src) < 9 {
		err = fmt.Errorf("byte slice length must be at least 9 bytes")
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
	out.MethodID = int8(src[0])
	src = src[1:]

	(*ori) = (*ori)[len((*ori))-len(src):]
	return
}
