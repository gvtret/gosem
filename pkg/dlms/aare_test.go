package dlms

import (
	"bytes"
	"testing"
)

func TestDecodeAARQ(t *testing.T) {
	src := decodeHexString("6129A109060760857405080101A203020100A305A103020100BE10040E0800065F1F040000101D00800007")
	a, err := DecodeAARE(&src)
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
	src := decodeHexString("6148A109060760857405080103A203020100A305A103020100A40A04084C475A2022604828BE230421281F300000003149963E23D6DA824A369644B66A9A17C60C3CA3F63E58608FA192")
	// Decoded: 08 00 06 5F 1F 04 00 00 18 1F 00 FA 00 07

	a, err := DecodeAARE(&src)
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
