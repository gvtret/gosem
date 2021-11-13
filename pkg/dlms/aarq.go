package dlms

import (
	"bytes"
	"encoding/binary"
)

// APDU types
const (
	PduTypeProtocolVersion            = 0
	PduTypeApplicationContextName     = 1
	PduTypeCalledAPTitle              = 2
	PduTypeCalledAEQualifier          = 3
	PduTypeCalledAPInvocationID       = 4
	PduTypeCalledAEInvocationID       = 5
	PduTypeCallingAPTitle             = 6
	PduTypeCallingAEQualifier         = 7
	PduTypeCallingAPInvocationID      = 8
	PduTypeCallingAEInvocationID      = 9
	PduTypeSenderAcseRequirements     = 10
	PduTypeMechanismName              = 11
	PduTypeCallingAuthenticationValue = 12
	PduTypeImplementationInformation  = 29
	PduTypeUserInformation            = 30
)

// BER encoding enumeration values
const (
	BERTypeContext     = 0x80
	BERTypeApplication = 0x40
	BERTypeConstructed = 0x20
)

// Application context definitions
const (
	ApplicationContextLNNoCiphering = 1
	ApplicationContextSNNoCiphering = 2
	ApplicationContextLNCiphering   = 3
	ApplicationContextSNCiphering   = 4
)

func EncodeAARQ(settings *Settings) (out []byte, err error) {
	var buf bytes.Buffer

	// Application Association Request
	buf.WriteByte(BERTypeApplication | BERTypeConstructed)

	// APDU length (to be filled in later)
	buf.WriteByte(0x00)

	buf.Write(generateApplicationContextName(settings))
	buf.Write(generateAuthentication(settings))
	buf.Write(generateUserInformation(settings))

	out = buf.Bytes()

	// Add length
	out[1] = byte(len(out) - 2)

	return
}

func generateApplicationContextName(settings *Settings) (out []byte) {
	var buf bytes.Buffer

	// Application context name
	buf.WriteByte(BERTypeContext | BERTypeConstructed | PduTypeApplicationContextName)
	buf.Write([]byte{0x09, 0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01})
	if settings.Ciphering.Security == SecurityNone && len(settings.Ciphering.SystemTitle) == 0 {
		buf.Write([]byte{ApplicationContextLNNoCiphering})
	} else {
		buf.Write([]byte{ApplicationContextLNCiphering})
	}

	if len(settings.Ciphering.SystemTitle) > 0 {
		// Add calling-AP-title
		buf.WriteByte(BERTypeContext | BERTypeConstructed | PduTypeCallingAPTitle)
		buf.Write([]byte{0x0A, 0x04, 0x08})
		buf.Write(settings.Ciphering.SystemTitle)
	}

	out = buf.Bytes()

	return
}

func generateAuthentication(settings *Settings) (out []byte) {
	var buf bytes.Buffer

	if settings.Authentication != AuthenticationNone {
		// Add sender ACSE-requirements field component.
		buf.WriteByte(BERTypeContext | PduTypeSenderAcseRequirements)
		buf.Write([]byte{0x02, 0x07, 0x80})

		// Add mechanism name.
		buf.WriteByte(BERTypeContext | PduTypeMechanismName)
		buf.Write([]byte{0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x02})
		buf.WriteByte(byte(settings.Authentication))

		// Add Calling authentication information.
		buf.WriteByte(BERTypeContext | BERTypeConstructed | PduTypeCallingAuthenticationValue)
		buf.WriteByte(byte(2 + len(settings.Password)))
		buf.WriteByte(0x80)
		buf.WriteByte(byte(len(settings.Password)))
		buf.Write(settings.Password)
	}

	out = buf.Bytes()

	return
}

func generateUserInformation(settings *Settings) (out []byte) {
	var buf bytes.Buffer

	// User information
	buf.WriteByte(BERTypeContext | BERTypeConstructed | PduTypeUserInformation)
	initiateRequest := getInitiateRequest(settings)

	if settings.Ciphering.Security != SecurityNone {
		initiateRequest = CipherData(&settings.Ciphering, initiateRequest, TagGloInitiateRequest, false)
	}

	buf.WriteByte(byte(2 + len(initiateRequest)))
	buf.WriteByte(0x04)
	buf.WriteByte(byte(len(initiateRequest)))
	buf.Write(initiateRequest)

	out = buf.Bytes()

	return
}

func getInitiateRequest(settings *Settings) (out []byte) {
	var buf bytes.Buffer

	// Application Association Request
	buf.WriteByte(byte(TagInitiateRequest))

	if settings.Ciphering.Security == SecurityNone && len(settings.Ciphering.DedicatedKey) == 0 {
		buf.WriteByte(0x00)
	} else {
		buf.WriteByte(0x01)
		buf.WriteByte(byte(len(settings.Ciphering.DedicatedKey)))
		buf.Write(settings.Ciphering.DedicatedKey)
	}

	buf.Write([]byte{0x00, 0x00, 0x06, 0x5F, 0x1F, 0x04, 0x00, 0x00, 0x18, 0x1F})

	maxPduSize := make([]byte, 2)
	binary.BigEndian.PutUint16(maxPduSize, uint16(settings.MaxPduSize))
	buf.Write(maxPduSize)

	out = buf.Bytes()

	return
}
