package dlms

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

func TestCipherData(t *testing.T) {
	cfg := Cipher{
		Tag:          TagGloInitiateRequest,
		Security:     SecurityAuthenticationEncryption,
		SystemTitle:  decodeHexString("4D4D4D0000BC614E"),
		Key:          decodeHexString("000102030405060708090A0B0C0D0E0F"),
		AuthKey:      decodeHexString("D0D1D2D3D4D5D6D7D8D9DADBDCDDDEDF"),
		FrameCounter: 0x01234567,
	}
	data := decodeHexString("01011000112233445566778899AABBCCDDEEFF0000065F1F0400007E1F04B0")
	result := decodeHexString("21303001234567801302FF8A7874133D414CED25B42534D28DB0047720606B175BD52211BE6841DB204D39EE6FDB8E356855")

	out, err := CipherData(cfg, data)
	if err != nil {
		t.Errorf("Got an error when ciphering: %v", err)
	}

	res := bytes.Compare(out, result)
	if res != 0 {
		t.Errorf("Failed. get: %s, should: %s", encodeHexString(out), encodeHexString(result))
	}
}

func TestCipherError(t *testing.T) {
	cfg := Cipher{}
	data := decodeHexString("01011000112233445566778899AABBCCDDEEFF0000065F1F0400007E1F04B0")

	_, err := CipherData(cfg, data)
	if err == nil {
		t.Errorf("Should get an error when ciphering")
	}
}

func TestDecipherData(t *testing.T) {
	cfg := Cipher{
		Tag:          TagGloInitiateRequest,
		Security:     SecurityAuthenticationEncryption,
		SystemTitle:  decodeHexString("4D4D4D0000BC614E"),
		Key:          decodeHexString("000102030405060708090A0B0C0D0E0F"),
		AuthKey:      decodeHexString("D0D1D2D3D4D5D6D7D8D9DADBDCDDDEDF"),
		FrameCounter: 0x01234567,
	}

	data := decodeHexString("21303001234567801302FF8A7874133D414CED25B42534D28DB0047720606B175BD52211BE6841DB204D39EE6FDB8E356855")
	result := decodeHexString("01011000112233445566778899AABBCCDDEEFF0000065F1F0400007E1F04B0")

	out, err := DecipherData(cfg, data)
	if err != nil {
		t.Errorf("Got an error when deciphering: %v", err)
	}

	res := bytes.Compare(out, result)
	if res != 0 {
		t.Errorf("Failed. get: %s, should: %s", encodeHexString(out), encodeHexString(result))
	}

	_, err = DecipherData(cfg, data[:len(data)-1])
	if err == nil {
		t.Errorf("Should get an error when deciphering")
	}

	cfg.Key[1] = 0x00
	_, err = DecipherData(cfg, data)
	if err == nil {
		t.Errorf("Should get an error when deciphering")
	}

	cfg.Key = nil
	_, err = DecipherData(cfg, data)
	if err == nil {
		t.Errorf("Should get an error when deciphering")
	}

	data[2] = 0x00
	_, err = DecipherData(cfg, data)
	if err == nil {
		t.Errorf("Should get an error when deciphering")
	}

	data[1] = 0xFF
	_, err = DecipherData(cfg, data)
	if err == nil {
		t.Errorf("Should get an error when deciphering")
	}

	data[0] = 0x31
	_, err = DecipherData(cfg, data)
	if err == nil {
		t.Errorf("Should get an error when deciphering")
	}
}

func decodeHexString(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

func encodeHexString(b []byte) string {
	return strings.ToUpper(hex.EncodeToString(b))
}
