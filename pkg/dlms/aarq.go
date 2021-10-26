package dlms

import "bytes"

const (
	DlmsAuthenticationNone = 0
	DlmsAuthenticationLow  = 1
)

type DlmsSettings struct {
	Authentication byte
	Password       []byte
}

const (
	TagAARQApplicationContextName     = 1
	TagAARQCallingAPTitle             = 6
	TagAARQSenderAcseRequirements     = 10
	TagAARQMechanismName              = 11
	TagAARQCallingAuthenticationValue = 12
	TagAARQUserInformation            = 30
)

const (
	BERTypeContext = 0x80
)

func EncodeAARQ(settings *DlmsSettings) (out []byte, err error) {
	var buf bytes.Buffer
	buf.WriteByte(byte(TagAARQ))

	if settings.Authentication != DlmsAuthenticationNone {
		buf.WriteByte(BERTypeContext | TagAARQSenderAcseRequirements)
		buf.Write([]byte{0x02, 0x07, 0x80})

		buf.WriteByte(BERTypeContext | TagAARQMechanismName)
		buf.Write([]byte{0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02})
		buf.WriteByte(settings.Authentication)

		buf.WriteByte(BERTypeContext | TagAARQCallingAuthenticationValue)
		buf.WriteByte(byte(2 + len(settings.Password)))
		buf.WriteByte(0x80)
		buf.WriteByte(byte(len(settings.Password)))
		buf.Write(settings.Password)
	}

	out = buf.Bytes()
	return
}
