package dlms

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	DlmsVersion = 0x06

	VAANameLN = 0x0007
	VAANameSN = 0xFA00
)

var (
	ErrWrongVersion = errors.New("wrong DLMS version")
	ErrWrongVAAName = errors.New("wrong VAA name")
)

type InitiateResponse struct {
	NegotiatedQualityOfService *uint8
	NegotiatedConformance      uint32
	ServerMaxReceivePduSize    uint16
}

func CreateInitiateResponse(qualityOfService *uint8, conformance uint32, maxReceivePduSize uint16) *InitiateResponse {
	return &InitiateResponse{
		NegotiatedQualityOfService: qualityOfService,
		NegotiatedConformance:      conformance,
		ServerMaxReceivePduSize:    maxReceivePduSize,
	}
}

func (ir InitiateResponse) Encode() (out []byte, err error) {
	var buf bytes.Buffer

	buf.WriteByte(TagInitiateResponse.Value())

	if ir.NegotiatedQualityOfService != nil {
		buf.WriteByte(0x01)
		buf.WriteByte(*ir.NegotiatedQualityOfService)
	} else {
		buf.WriteByte(0x00)
	}

	buf.WriteByte(DlmsVersion)

	buf.Write([]byte{0x5F, 0x1F, 0x04})

	negotiatedConformance := make([]byte, 4)
	binary.BigEndian.PutUint32(negotiatedConformance, ir.NegotiatedConformance)
	buf.Write(negotiatedConformance)

	serverMaxReceivePduSize := make([]byte, 2)
	binary.BigEndian.PutUint16(serverMaxReceivePduSize, ir.ServerMaxReceivePduSize)
	buf.Write(serverMaxReceivePduSize)

	vaaName := make([]byte, 2)
	binary.BigEndian.PutUint16(vaaName, VAANameLN)
	buf.Write(vaaName)

	out = buf.Bytes()
	return
}

func DecodeInitiateResponse(ori *[]byte) (out InitiateResponse, err error) {
	src := *ori

	if len(src) < 14 {
		err = ErrWrongLength(len(src), 14)
		return
	}

	if src[0] != TagInitiateResponse.Value() {
		err = ErrWrongTag(0, src[0], byte(TagInitiateResponse))
		return
	}

	if src[1] == 0x01 {
		negotiatedQualityOfService := src[2]
		out.NegotiatedQualityOfService = &negotiatedQualityOfService
		src = src[3:]
	} else {
		src = src[2:]
	}

	if src[0] != DlmsVersion {
		err = ErrWrongVersion
		return
	}

	if !bytes.Equal(src[1:4], []byte{0x5F, 0x1F, 0x04}) {
		err = ErrWrongSlice(src[1:4], []byte{0x5F, 0x1F, 0x04})
		return
	}

	out.NegotiatedConformance = binary.BigEndian.Uint32(src[4:8])
	out.ServerMaxReceivePduSize = binary.BigEndian.Uint16(src[8:10])

	vaaName := binary.BigEndian.Uint16(src[10:12])
	if vaaName != VAANameLN && vaaName != VAANameSN {
		err = ErrWrongVAAName
		return
	}

	src = src[12:]

	(*ori) = (*ori)[len((*ori))-len(src):]
	return
}
