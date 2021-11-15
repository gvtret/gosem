package dlms

import "bytes"

type AssociationResult uint8

const (
	AssociationResultAccepted          AssociationResult = 0
	AssociationResultPermanentRejected AssociationResult = 1
	AssociationResultTransientRejected AssociationResult = 2
)

type SourceDiagnostic uint8

const (
	SourceDiagnosticNone                                       SourceDiagnostic = 0
	SourceDiagnosticNoReasonGiven                              SourceDiagnostic = 1
	SourceDiagnosticApplicationContextNameNotSupported         SourceDiagnostic = 2
	SourceDiagnosticCallingAPTitleNotRecognized                SourceDiagnostic = 3
	SourceDiagnosticCallingAPInvocationIdentifierNotRecognized SourceDiagnostic = 4
	SourceDiagnosticCallingAEQualifierNotRecognized            SourceDiagnostic = 5
	SourceDiagnosticCallingAEInvocationIdentifierNotRecognized SourceDiagnostic = 6
	SourceDiagnosticCalledAPTitleNotRecognized                 SourceDiagnostic = 7
	SourceDiagnosticCalledAPInvocationIdentifierNotRecognized  SourceDiagnostic = 8
	SourceDiagnosticCalledAEQualifierNotRecognized             SourceDiagnostic = 9
	SourceDiagnosticCalledAEInvocationIdentifierNotRecognized  SourceDiagnostic = 10
	SourceDiagnosticAuthenticationMechanismNameNotRecognized   SourceDiagnostic = 11
	SourceDiagnosticAuthenticationMechanismNameRequired        SourceDiagnostic = 12
	SourceDiagnosticAuthenticationFailure                      SourceDiagnostic = 13
	SourceDiagnosticAuthenticationRequired                     SourceDiagnostic = 14
)

type AARE struct {
	ApplicationContext ApplicationContext
	AssociationResult  AssociationResult
	SourceDiagnostic   SourceDiagnostic
	SourceSystemTitle  []byte
}

func DecodeAARE(ori *[]byte) (out AARE, err error) {
	src := append([]byte(nil), (*ori)...)

	if len(src) < 2 {
		err = ErrWrongLength(len(src), 3)
		return
	}

	if src[0] != TagAARE.Value() {
		err = ErrWrongTag(0, src[0], byte(TagAARE))
		return
	}

	length := int(2 + src[1])
	if len(src) < length {
		err = ErrWrongLength(len(src), length)
		return
	}

	src = src[2:]
	length -= 2

	for {
		if length == 0 {
			break
		}

		if len(src) < 2 {
			err = ErrWrongLength(len(src), 2)
			return
		}

		tagLength := int(src[1])
		if len(src) < (2 + tagLength) {
			err = ErrWrongLength(len(src), 2+tagLength)
			return
		}

		tag := src[0]
		switch tag {
		case BERTypeContext | BERTypeConstructed | PduTypeApplicationContextName:
			// Application context name - 0xA1
			if tagLength != 9 {
				err = ErrWrongLength(tagLength, 9)
				return
			}
			rsp := []byte{0x06, 0x07, 0x60, 0x85, 0x74, 0x05, 0x08, 0x01}
			if !bytes.Equal(src[2:10], rsp) {
				err = ErrWrongSlice(src[2:10], rsp)
				return
			}
			out.ApplicationContext = ApplicationContext(src[10])
		case BERTypeContext | BERTypeConstructed | PduTypeCalledAPTitle:
			// Association result - 0xA2
			if tagLength != 3 {
				err = ErrWrongLength(tagLength, 3)
				return
			}
			rsp := []byte{0x02, 0x01}
			if !bytes.Equal(src[2:4], rsp) {
				err = ErrWrongSlice(src[2:4], rsp)
				return
			}
			out.AssociationResult = AssociationResult(src[4])
		case BERTypeContext | BERTypeConstructed | PduTypeCalledAEQualifier:
			// Associate source diagnostic - 0xA3
			if tagLength != 5 {
				err = ErrWrongLength(tagLength, 5)
				return
			}
			rsp := []byte{0x03, 0x02, 0x01}
			if !bytes.Equal(src[3:6], rsp) {
				err = ErrWrongSlice(src[3:6], rsp)
				return
			}
			out.SourceDiagnostic = SourceDiagnostic(src[6])
		case BERTypeContext | BERTypeConstructed | PduTypeCalledAPInvocationID:
			// AP title - 0xA4
			if tagLength != 10 {
				err = ErrWrongLength(tagLength, 10)
				return
			}
			rsp := []byte{0x04, 0x08}
			if !bytes.Equal(src[2:4], rsp) {
				err = ErrWrongSlice(src[2:4], rsp)
				return
			}
			out.SourceSystemTitle = make([]byte, 8)
			copy(out.SourceSystemTitle, src[4:12])
		case BERTypeContext | BERTypeConstructed | PduTypeUserInformation:
			// User information - 0xBE
		}

		src = src[2+tagLength:]
		length -= 2 + tagLength
	}

	(*ori) = (*ori)[len((*ori))-len(src):]
	return
}
