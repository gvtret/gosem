package dlms

import (
	"crypto/rand"
	"fmt"
)

type Authentication byte

// Authentication mechanism definitions
const (
	AuthenticationNone       Authentication = 0 // No authentication is used.
	AuthenticationLow        Authentication = 1 // Low authentication is used.
	AuthenticationHigh       Authentication = 2 // High authentication is used.
	AuthenticationHighMD5    Authentication = 3 // High authentication is used. Password is hashed with MD5.
	AuthenticationHighSHA1   Authentication = 4 // High authentication is used. Password is hashed with SHA1.
	AuthenticationHighGmac   Authentication = 5 // High authentication is used. Password is hashed with GMAC.
	AuthenticationHighSha256 Authentication = 6 // High authentication is used. Password is hashed with SHA-256.
	AuthenticationHighEcdsa  Authentication = 7 // High authentication is used. Password is hashed with ECDSA.
)

type SecurityLevel byte

const (
	SecurityLevelNone SecurityLevel = iota
	SecurityLevelGlobalKey
	SecurityLevelDedicatedKey
)

type Security byte

const (
	SecurityNone            Security = 0    // Transport security is not used.
	SecurityAuthentication  Security = 0x10 // Authentication security is used.
	SecurityEncryption      Security = 0x20 // Encryption security is used.
	SecurityKeySetBroadcast Security = 0x40 // Key set broadcast security is used.
)

type Ciphering struct {
	Level             SecurityLevel
	Security          Security
	SystemTitle       []byte
	SourceSystemTitle []byte
	UnicastKey        []byte
	AuthenticationKey []byte
	UnicastKeyIC      uint32
	DedicatedKey      []byte
	DedicatedKeyIC    uint32
}

type Settings struct {
	Authentication   Authentication
	Password         []byte
	Ciphering        Ciphering
	MaxPduSize       int
	ConformanceBlock int
}

func NewSettingsWithoutAuthentication() (Settings, error) {
	s := Settings{
		Authentication: AuthenticationNone,
		Password:       nil,
		Ciphering:      Ciphering{},
		MaxPduSize:     256,
		ConformanceBlock: ConformanceBlockBlockTransferWithGetOrRead | ConformanceBlockBlockTransferWithSetOrWrite |
			ConformanceBlockGet | ConformanceBlockSet | ConformanceBlockSelectiveAccess | ConformanceBlockEventNotification |
			ConformanceBlockAction,
	}

	return s, nil
}

func NewSettingsWithLowAuthentication(password []byte) (Settings, error) {
	return NewSettingsWithLowAuthenticationAndCiphering(password, Ciphering{})
}

func NewSettingsWithLowAuthenticationAndCiphering(password []byte, cipher Ciphering) (Settings, error) {
	if len(password) == 0 {
		return Settings{}, fmt.Errorf("password must not be empty")
	}

	s := Settings{
		Authentication: AuthenticationLow,
		Password:       password,
		Ciphering:      cipher,
		MaxPduSize:     256,
		ConformanceBlock: ConformanceBlockBlockTransferWithGetOrRead | ConformanceBlockBlockTransferWithSetOrWrite |
			ConformanceBlockGet | ConformanceBlockSet | ConformanceBlockSelectiveAccess | ConformanceBlockEventNotification |
			ConformanceBlockAction,
	}

	return s, nil
}

func NewCiphering(level SecurityLevel, security Security, systemTitle []byte, unicastKey []byte, unicastKeyIC uint32, authenticationKey []byte) (Ciphering, error) {
	if len(systemTitle) != 8 {
		return Ciphering{}, fmt.Errorf("system title must be 8 bytes long")
	}

	if len(unicastKey) != 16 {
		return Ciphering{}, fmt.Errorf("unicast key must be 16 bytes long")
	}

	if len(authenticationKey) != 16 {
		return Ciphering{}, fmt.Errorf("authentication key must be 16 bytes long")
	}

	dk, err := generateKey()
	if err != nil {
		return Ciphering{}, fmt.Errorf("could not generate dedicated key: %w", err)
	}

	c := Ciphering{
		Level:             level,
		Security:          security,
		SystemTitle:       systemTitle,
		SourceSystemTitle: nil,
		UnicastKey:        unicastKey,
		AuthenticationKey: authenticationKey,
		UnicastKeyIC:      unicastKeyIC,
		DedicatedKey:      dk,
		DedicatedKeyIC:    1,
	}

	return c, nil
}

func generateKey() ([]byte, error) {
	dk := make([]byte, 16)
	_, err := rand.Read(dk)
	if err != nil {
		return nil, err
	}

	return dk, nil
}
