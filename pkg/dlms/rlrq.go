package dlms

import (
	"bytes"
)

type ReleaseRequestReason uint8

const (
	ReleaseRequestReasonNormal      ReleaseRequestReason = 0
	ReleaseRequestReasonUrgent      ReleaseRequestReason = 1
	ReleaseRequestReasonUserDefined ReleaseRequestReason = 30
)

func EncodeRLRQ(settings *Settings) (out []byte, err error) {
	var buf bytes.Buffer

	// Application Association Request
	buf.WriteByte(byte(TagRLRQ))

	// APDU length (to be filled in later)
	buf.WriteByte(0x00)

	if settings != nil && settings.Ciphering.Security != SecurityNone {
		releaseRequest := []byte{BERTypeContext, 1, byte(ReleaseRequestReasonNormal)}
		buf.Write(releaseRequest)

		userInfo, err := generateUserInformation(settings)
		if err != nil {
			return nil, err
		}
		buf.Write(userInfo)
	}

	out = buf.Bytes()

	// Add length
	out[1] = byte(len(out) - 2)

	return
}
