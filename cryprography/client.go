package cryptography
import (
	// general things
	"fmt"
	"sync"
	"strings"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/x509"
	//"encoding/hex"
	"encoding/base64"
	"github.com/cloudflare/circl/kem/kyber/kyber768"
)

type Keys struct {
	// kem keys
	pqPk		*kyber768.PublicKey
	pqSk		*kyber768.PrivateKey

	ecSk		*ecdh.PrivateKey	// we can get pk from sk

	peerPqPk	*kyber768.PublicKey
	peerEcPk	*ecdh.PublicKey
}

type CryptClient struct {
	mtx	sync.Mutex
	keys	*Keys
}

func NewClient() ( *CryptClient, error ) {

	pqPk, pqSk, err := kyber768.GenerateKeyPair( rand.Reader )
	if err != nil {
		return nil, err
	}

	x25519 := ecdh.X25519()
	ecSk, err := x25519.GenerateKey( rand.Reader )
	if err != nil {
		return nil, err
	}

	return &CryptClient{
		sync.Mutex{},
		&Keys{
			pqPk,
			pqSk,

			ecSk,

			nil,
			nil,
		},
	}, nil
}

func ClientFromKeys( pk string, sk string ) ( *CryptClient, error ) {

	// pk stores only pqPk, sk stores both ecSk and pqSk
	parts := strings.Split( sk, "|" )
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid key format")
	}


	pkBytes, err := base64.StdEncoding.DecodeString( pk )
	if err != nil {
		return nil, err
	}

	skBytes, err := base64.StdEncoding.DecodeString( parts[0] )
	if err != nil {
		return nil, err
	}
	pkKey := &kyber768.PublicKey{}
	pkKey.Unpack( pkBytes )

	skKey := &kyber768.PrivateKey{}
	skKey.Unpack( skBytes )

	pkcs8, err := base64.StdEncoding.DecodeString( parts[1] )
	if err != nil {
		return nil, err
	}

	ecSk, err := x509.ParsePKCS8PrivateKey( pkcs8 )
	if err != nil {
		return nil, err
	}

	ecSkTyped, ok := ecSk.(*ecdh.PrivateKey)
	if ok == false {
		return nil, fmt.Errorf("Failed to convert into ECDH private key.")
	}

	return &CryptClient{
		sync.Mutex{},
		&Keys{
			pkKey,
			skKey,

			ecSkTyped,

			nil,
			nil,
		},
	}, nil
}

func(c *CryptClient) ECCPrivateKey() *ecdh.PrivateKey {
	return c.keys.ecSk
}

func(c *CryptClient) Decapsulate( data []byte ) []byte {
	ss := make([]byte, kyber768.SharedKeySize)
	c.keys.pqSk.DecapsulateTo( ss, data )
	return ss
}

func(c *CryptClient) PackData( data []byte, packetSize uint ) ([]byte, error) {
	return PackData( data, packetSize )
}

func(c *CryptClient) UnpackData( data []byte, packetSize uint ) ([]byte, error) {
	return UnpackData( data, packetSize )
}


func(c *CryptClient) DecapsulateAndUnpack( packetSize uint, data []byte ) ([]byte, []byte, []byte, error) {
	unpacked, err := c.UnpackData( data, packetSize )
	if err != nil {
		return nil, nil, nil, err
	}

	if len(unpacked) < kyber768.CiphertextSize + Delta {
		return nil, nil, nil, fmt.Errorf("Too short, not an encapsulated key.")
	}

	ct := unpacked[:kyber768.CiphertextSize]
	encryptedEcPk := unpacked[kyber768.CiphertextSize:]
	hmac := encryptedEcPk[ len(encryptedEcPk) - HashSize : ]
	encryptedEcPk = encryptedEcPk[ : len(encryptedEcPk) - HashSize ]

	ss := c.Decapsulate( ct )
	ecPkDec, err := Decrypt( encryptedEcPk, ss )
	if err != nil {
		return nil, nil, nil, err
	}

	// set the peer's public key
	ecPk, err := x509.ParsePKIXPublicKey( ecPkDec )
	if err != nil {
		return nil, nil, nil, err
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	var ok bool
	c.keys.peerEcPk, ok = ecPk.(*ecdh.PublicKey)
	if !ok {
		return nil, nil, nil, fmt.Errorf("DecapsAndUnpack: invalid pk format")
	}

	// extract shared secret from elliptic curve public key
	ss2, err := c.keys.ecSk.ECDH( c.keys.peerEcPk )
	if err != nil {
		return nil, nil, nil, err
	}

	finalSS := DeriveSharedSecret( ss, ss2 )
	//fmt.Println("Shared secret [d]: ", hex.EncodeToString(finalSS) )
	return finalSS, ecPkDec, hmac, nil
}

// useful for key exchange
func(c *CryptClient) GetPublicKey( networkKey []byte ) ( []byte, error ) {

	//fmt.Println("Public key size:", kyber768.PublicKeySize)	// 1184
	//fmt.Println("Ciphertext size:", kyber768.CiphertextSize )	// 1088 (1160 with packed pk) 1160 - 1088 = 72
	buf := make([]byte, kyber768.PublicKeySize)
	c.keys.pqPk.Pack( buf )
	ec, err := x509.MarshalPKIXPublicKey( c.keys.ecSk.Public() )
	if err != nil {
		return nil, err
	}
	//fmt.Println("[CryptClient::GetPublicKey] Packed ecc public key:", len(ec) )
	//fmt.Println("Public key in string form:", string(append(buf, ec...))) // fine, non-printable
	pk := append( buf, ec... )
	// append hmac of the key so users from the same network could verify us.
	if networkKey != nil && len( networkKey ) == SymKeySize {
		hmac := HMACBytes( pk, networkKey )
		pk = append( pk, hmac... )
	}

	/*fmt.Println( "[CryptClient::GetPublicKey] public key bytes:",
		base64.StdEncoding.EncodeToString( pk[:16] ),
		base64.StdEncoding.EncodeToString( pk[kyber768.PublicKeySize:][:16] ),
	)*/
	return pk, nil
}

func(c *CryptClient) PkToString() string {
	buf := make( []byte, kyber768.PublicKeySize )
	c.keys.pqPk.Pack( buf )
	return base64.StdEncoding.EncodeToString( buf )
}

func(c *CryptClient) SkToString() (string, error) {

	buf := make( []byte, kyber768.PrivateKeySize )
	c.keys.pqSk.Pack( buf )

	res := base64.StdEncoding.EncodeToString( buf )
	tmpres, err := x509.MarshalPKCS8PrivateKey( c.keys.ecSk )
	if err != nil {
		return "", err
	}
	return res + "|" + base64.StdEncoding.EncodeToString( tmpres ), nil
}
