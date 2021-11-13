package dlms

import (
	"bytes"
	"testing"
)

func TestEncodeAARQWithoutAuthentication(t *testing.T) {
	settings := &Settings{
		Authentication: AuthenticationNone,
		MaxPduSize:     512,
	}

	out, err := EncodeAARQ(settings)
	if err != nil {
		t.Errorf("Encode Failed. err: %v", err)
	}
	result := decodeHexString("601DA109060760857405080101BE10040E01000000065F1F040000181F0200")
	res := bytes.Compare(out, result)
	if res != 0 {
		t.Errorf("Failed. get: %s, should: %s", encodeHexString(out), encodeHexString(result))
	}
}

func TestEncodeAARQWithLowAuthentication(t *testing.T) {
	settings := &Settings{
		Authentication: AuthenticationLow,
		MaxPduSize:     256,
		Password:       []byte("12345678"),
	}

	out, err := EncodeAARQ(settings)
	if err != nil {
		t.Errorf("Encode Failed. err: %v", err)
	}
	result := decodeHexString("6036A1090607608574050801018A0207808B0760857405080201AC0A80083132333435363738BE10040E01000000065F1F040000181F0100")
	res := bytes.Compare(out, result)
	if res != 0 {
		t.Errorf("Failed. get: %s, should: %s", encodeHexString(out), encodeHexString(result))
	}

	settings.Password = nil
	_, err = EncodeAARQ(settings)
	if err == nil {
		t.Errorf("Should be error")
	}
}

func TestEncodeAARQWithLowAuthenticationAndCipher(t *testing.T) {
	ciphering := Ciphering{
		Security:          SecurityAuthenticationEncryption,
		SystemTitle:       decodeHexString("4349520000000001"),
		BlockCipherKey:    decodeHexString("00112233445566778899AABBCCDDEEFF"),
		AuthenticationKey: decodeHexString("00112233445566778899AABBCCDDEEFF"),
		DedicatedKey:      decodeHexString("E803739DBE338C3A790D8D1B12C63FE2"),
		InvocationCounter: 0x00000107,
	}

	settings := &Settings{
		Authentication: AuthenticationLow,
		Password:       []byte("JuS66BCZ"),
		Ciphering:      ciphering,
		MaxPduSize:     512,
	}

	out, err := EncodeAARQ(settings)
	if err != nil {
		t.Errorf("Encode Failed. err: %v", err)
	}
	result := decodeHexString("6066A109060760857405080103A60A040843495200000000018A0207808B0760857405080201AC0A80084A7553363642435ABE340432213030000001078E6341442275404C816C6BED3E33AE809EC51E1D0E428BE8F5F643E26C3DD89FD2E3F2220097124F58E0F4")
	res := bytes.Compare(out, result)
	if res != 0 {
		t.Errorf("Failed. get: %s, should: %s", encodeHexString(out), encodeHexString(result))
	}

	settings.Ciphering.BlockCipherKey = nil
	_, err = EncodeAARQ(settings)
	if err == nil {
		t.Errorf("Should be error")
	}
}
