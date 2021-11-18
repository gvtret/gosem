package dlms

import (
	"bytes"
	"testing"
)

func TestDecodeAARQ(t *testing.T) {
	src := decodeHexString("6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")
	a, err := DecodeAARE(nil, &src)
	if err != nil {
		t.Errorf("Failed on DecodeAARE. Err: %v", err)
	}

	if a.ApplicationContext != ApplicationContextLNNoCiphering {
		t.Errorf("ApplicationContext is not correct. Get %v", a.ApplicationContext)
	}

	if a.AssociationResult != AssociationResultAccepted {
		t.Errorf("AssociationResult is not accepted. Get %v", a.AssociationResult)
	}

	if a.SourceDiagnostic != SourceDiagnosticNone {
		t.Errorf("SourceDiagnostic is not None. Get %v", a.AssociationResult)
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
	a, err := DecodeAARE(settings, &src)
	if err != nil {
		t.Errorf("Failed on DecodeAARE. Err: %v", err)
	}

	if a.ApplicationContext != ApplicationContextLNCiphering {
		t.Errorf("ApplicationContext is not correct. Get %v", a.ApplicationContext)
	}

	if a.AssociationResult != AssociationResultAccepted {
		t.Errorf("AssociationResult is not accepted. Get %v", a.AssociationResult)
	}

	if a.SourceDiagnostic != SourceDiagnosticNone {
		t.Errorf("SourceDiagnostic is not None. Get %v", a.AssociationResult)
	}

	sourceSystemTitle := decodeHexString("4C475A2022604828")
	res := bytes.Compare(a.SourceSystemTitle, sourceSystemTitle)
	if res != 0 {
		t.Errorf("SourceSystemTitle is not correct. Get %v", encodeHexString(a.SourceSystemTitle))
	}
}
