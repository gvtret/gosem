package dlms

import (
	"bytes"
	"testing"
)

func TestDecodeAARQ(t *testing.T) {
	src := decodeHexString("6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")
	aare, err := DecodeAARE(nil, &src)
	if err != nil {
		t.Errorf("Failed on DecodeAARE. Err: %v", err)
	}

	if aare.ApplicationContext != ApplicationContextLNNoCiphering {
		t.Errorf("Invalid ApplicationContext. Get %v", aare.ApplicationContext)
	}

	if aare.AssociationResult != AssociationResultAccepted {
		t.Errorf("Invalid AssociationResult. Get %v", aare.AssociationResult)
	}

	if aare.SourceDiagnostic != SourceDiagnosticNone {
		t.Errorf("Invalid SourceDiagnostic. Get %v", aare.SourceDiagnostic)
	}

	if aare.InitiateResponse == nil {
		t.Errorf("InitiateResponse is nil")
		return
	}

	if aare.ConfirmedServiceError != nil {
		t.Errorf("ConfirmedServiceError is not nil")
	}

	if aare.InitiateResponse.ServerMaxReceivePduSize != 128 {
		t.Errorf("Invalid ServerMaxReceivePduSize. Get %v", aare.InitiateResponse.ServerMaxReceivePduSize)
	}
}

func TestDecodeRejectedAARQ(t *testing.T) {
	src := decodeHexString("611FA109060760857405080101A203020101A305A10302010DBE0604040E010600")
	aare, err := DecodeAARE(nil, &src)
	if err != nil {
		t.Errorf("Failed on DecodeAARE. Err: %v", err)
	}

	if aare.ApplicationContext != ApplicationContextLNNoCiphering {
		t.Errorf("Invalid ApplicationContext. Get %v", aare.ApplicationContext)
	}

	if aare.AssociationResult != AssociationResultPermanentRejected {
		t.Errorf("Invalid AssociationResult. Get %v", aare.AssociationResult)
	}

	if aare.SourceDiagnostic != SourceDiagnosticAuthenticationFailure {
		t.Errorf("Invalid SourceDiagnostic. Get %v", aare.SourceDiagnostic)
	}

	if aare.InitiateResponse != nil {
		t.Errorf("InitiateResponse is not nil")
	}

	if aare.ConfirmedServiceError == nil {
		t.Errorf("ConfirmedServiceError is nil")
		return
	}

	if aare.ConfirmedServiceError.ConfirmedServiceError != TagErrInitiateError {
		t.Errorf("Invalid confirmed service error. Get %v", aare.ConfirmedServiceError.ConfirmedServiceError)
	}
}

func TestDecodeAARQWithSecurity(t *testing.T) {
	ciphering := Ciphering{
		Security:          SecurityAuthenticationEncryption,
		SourceSystemTitle: decodeHexString("4349520000000001"),
		BlockCipherKey:    decodeHexString("00112233445566778899AABBCCDDEEFF"),
		AuthenticationKey: decodeHexString("00112233445566778899AABBCCDDEEFF"),
	}

	settings := &Settings{
		Ciphering: ciphering,
	}

	src := decodeHexString("6148A109060760857405080103A203020100A305A103020100A40A04084C475A2022604828BE230421281F300000003149963E23D6DA824A369644B66A9A17C60C3CA3F63E58608FA192")
	aare, err := DecodeAARE(settings, &src)
	if err != nil {
		t.Errorf("Failed on DecodeAARE. Err: %v", err)
	}

	if aare.ApplicationContext != ApplicationContextLNCiphering {
		t.Errorf("ApplicationContext is not correct. Get %v", aare.ApplicationContext)
	}

	if aare.AssociationResult != AssociationResultAccepted {
		t.Errorf("AssociationResult is not accepted. Get %v", aare.AssociationResult)
	}

	if aare.SourceDiagnostic != SourceDiagnosticNone {
		t.Errorf("SourceDiagnostic is not None. Get %v", aare.AssociationResult)
	}

	sourceSystemTitle := decodeHexString("4C475A2022604828")
	res := bytes.Compare(aare.SourceSystemTitle, sourceSystemTitle)
	if res != 0 {
		t.Errorf("SourceSystemTitle is not correct. Get %v", encodeHexString(aare.SourceSystemTitle))
	}
}
