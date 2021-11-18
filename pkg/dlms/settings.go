package dlms

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

type Security byte

const (
	SecurityNone                     Security = 0    // Transport security is not used.
	SecurityAuthentication           Security = 0x10 // Authentication security is used.
	SecurityEncryption               Security = 0x20 // Encryption security is used.
	SecurityAuthenticationEncryption Security = 0x30 // Authentication and encryption security are used.
)

type Ciphering struct {
	Security          Security
	SystemTitle       []byte
	SourceSystemTitle []byte
	BlockCipherKey    []byte
	AuthenticationKey []byte
	InvocationCounter uint32
	DedicatedKey      []byte
}

type Settings struct {
	Authentication Authentication
	Password       []byte
	Ciphering      Ciphering
	MaxPduSize     int
}
