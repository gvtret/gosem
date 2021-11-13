package dlms

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"gosem/pkg/axdr"
)

func CipherData(ciphering *Ciphering, data []byte, tag cosemTag, useDedicatedKey bool) ([]byte, error) {
	var key []byte
	if useDedicatedKey {
		key = ciphering.DedicatedKey
	} else {
		key = ciphering.BlockCipherKey
	}

	// Generate a new AES cipher using our 32 byte long key
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// GCM or Galois/Counter Mode, is a mode of operation for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCMWithTagSize(c, 12)
	if err != nil {
		return nil, err
	}

	// Initialization vector (or Nonce)
	iv := make([]byte, gcm.NonceSize())
	copy(iv, ciphering.SystemTitle)
	binary.BigEndian.PutUint32(iv[8:], ciphering.InvocationCounter)

	// Associated data
	ad := make([]byte, 17)
	ad[0] = byte(ciphering.Security)
	copy(ad[1:], ciphering.AuthenticationKey)

	// Ciphered data prefix
	size, _ := axdr.EncodeLength(5 + len(data) + 12)
	dst := make([]byte, 6+len(size))
	dst[0] = byte(tag)
	copy(dst[1:], size)
	dst[1+len(size)] = byte(ciphering.Security)
	binary.BigEndian.PutUint32(dst[2+len(size):], ciphering.InvocationCounter)

	// Encrypt data
	return gcm.Seal(dst, iv, data, ad), nil
}
