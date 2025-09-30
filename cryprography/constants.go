package cryptography
import (
	"crypto/sha512"
	"github.com/cloudflare/circl/kem/kyber/kyber768"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	// encryption+signature modes
	SymKeySize = 32
	TagSize = 16
	NonceSize = chacha20poly1305.NonceSize

	// amount of bytes to encrypt in order to check if this is really a
	// capsulated secret key packet
	SeedSize = 16
	PkSize = kyber768.PublicKeySize
	Delta = 72	// how many bytes encapsulated and encrypted ecdh pk takes
	HashSize = sha512.Size
	TotalPkSize = 1292	// public keys + hmac of them.
)
