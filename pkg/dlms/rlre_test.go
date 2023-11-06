package dlms

import (
	"testing"
)

func TestDecodeRLRE(t *testing.T) {
	src := decodeHexString("6300")
	rlre, err := DecodeRLRE(&src)
	if err != nil {
		t.Errorf("Failed on DecodeAARE. Err: %v", err)
	}

	if rlre.ReleaseResponseReason != nil {
		t.Errorf("Invalid ReleaseResponseReason. Should be nil but get %v", rlre.ReleaseResponseReason)
	}
}

func TestDecodeRLREWithUserInformation(t *testing.T) {
	src := decodeHexString("6328800100BE230421281F30000000097A01C161F198612BB535740660BAEDA2FC42C287E527543BBA97")
	rlre, err := DecodeRLRE(&src)
	if err != nil {
		t.Errorf("Failed on DecodeAARE. Err: %v", err)
	}

	if *rlre.ReleaseResponseReason != ReleaseResponseReasonNormal {
		t.Errorf("Invalid AssociationResult. Get %v", *rlre.ReleaseResponseReason)
	}
}
