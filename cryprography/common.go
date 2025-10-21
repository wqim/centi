package cryptography
import (
	"fmt"	// for debug and errors
	//"bytes"
	"io"
	"strings"
	"runtime"
	"crypto/rand"
	"crypto/hmac"
	"crypto/sha512" // used for hashing data
	"encoding/hex"
	"encoding/base64"
	"encoding/binary"
	
	"encoding/ascii85"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)


// chacha20poly1305 encryption+authentication
func Encrypt( data, key []byte ) ( []byte, error ) {

	if data == nil || len(data) == 0 {
		return nil, nil	// should we return an error here?
	}

	if key == nil || len(key) != SymKeySize {
		return nil, fmt.Errorf("Invalid key")
	}
	nonce := make( []byte, chacha20poly1305.NonceSize )
	aead, err := chacha20poly1305.New( key )
	if err != nil {
		return nil, err
	}
	if _, err := rand.Read( nonce ); err != nil {
		return nil, err
	}

	ct := aead.Seal( nil, nonce, data, nil )
	data = append( nonce, ct... )

	return data, nil
}

func Decrypt( data, key []byte ) ( []byte, error ) {

	if data == nil || len(data) == 0 {
		return nil, nil	// should we return an error here?
	}

	if key == nil || len(key) != SymKeySize {
		return nil, fmt.Errorf("Invalid key")
	}

	if data == nil || len(data) < chacha20poly1305.NonceSize {
		return nil, fmt.Errorf("Invalid length of data")
	}

	nonce := data[:chacha20poly1305.NonceSize]
	data = data[chacha20poly1305.NonceSize:]
	aead, err := chacha20poly1305.New( key )
	if err != nil {
		return nil, err
	}
	pt, err := aead.Open( nil, nonce, data, nil )
	return pt, err
}


// generate a random amount of bytes
func GenRandom( size uint ) ([]byte, error) {
	if size == 0 {
		return nil, fmt.Errorf("[cryptography/common.go] GenRandom: Invalid size of random data")
	}
	data := make( []byte, size )
	if _, err := rand.Read( data ); err != nil {
		return nil, err
	}
	return data, nil
}

// calculate the hash of data
func Hash( data []byte ) string {
	if data == nil {
		return ""
	}
	hash := sha512.Sum512( data )
	return hex.EncodeToString( hash[:] )
}

// verify hash of data
func VerifyHash( data []byte, hash string ) bool {

	if data == nil && hash == "" {
		return true
	} else if data == nil || hash == "" {
		return false
	}

	if hash != Hash( data ) {
		return false
	}
	return true
}

// hmac function
func HMAC( data, skey []byte ) string {
	/*mac := hmac.New( sha512.New, skey )
	mac.Write( data )
	expectedMac := mac.Sum( nil ) */
	if data == nil || len(data) == 0 || skey == nil || len(skey) != SymKeySize {
		return ""
	}
	return hex.EncodeToString( HMACBytes( data, skey ) )
}

func HMACBytes( data, skey []byte ) []byte {
	if data == nil || len(data) == 0 {
		return nil
	}
	if skey == nil || len(skey) != SymKeySize {
		return nil
	}
	mac := hmac.New( sha512.New, skey )
	mac.Write( data )
	return mac.Sum( nil )
}

func VerifyHMAC( data, skey []byte, expected string ) bool {

	expectedBytes, err := hex.DecodeString( expected )
	if err != nil {
		return false
	}
	if len(expectedBytes) == 0 && (data == nil || len(data) == 0 || skey == nil || len(skey) != SymKeySize) {
		return true
	} else if len(data) == 0 || data == nil || skey == nil || len(skey) != SymKeySize {
		return false
	}

	mac := hmac.New( sha512.New, skey )
	mac.Write( data )
	msgHmac := mac.Sum( nil )
	return hmac.Equal( msgHmac, expectedBytes )
}

func VerifyHMACBytes( data, skey, expected []byte ) bool {
	if data == nil || len(data) == 0 || skey == nil || len(skey) != SymKeySize || expected == nil || len(expected) == 0 {
		// nothing to compare
		return true
	}
	real := HMACBytes( data, skey )
	return hmac.Equal( expected, real )
}


// working with public keys encoding
func EncodePublicKey( data []byte ) string {
	return base64.StdEncoding.EncodeToString( data )
}

func DecodePublicKey( data string ) ([]byte, error) {
	return base64.StdEncoding.DecodeString( data )
}

// encoding binary data inside packets
func EncodeData( data []byte ) string {
	if data == nil || len(data) == 0 {
		return ""
	}

	//return base64.StdEncoding.EncodeToString( data )
	buf := make( []byte, ascii85.MaxEncodedLen( len(data) ) )
	ascii85.Encode( buf, data )
	return string(buf)	// ascii85 is printable to return the string anyway
	
	//return base64.StdEncoding.EncodeToString( data )
}

func DecodeData( data string ) ([]byte, error) {

	if data == "" {	// nothing to decode.
		return nil, nil
	}

	buf := make( []byte, len(data))
	n, _, err := ascii85.Decode( buf, []byte(data), false )
	if err != nil {
		return nil, err
	}
	//fmt.Println("N =", n)
	//fmt.Println("M =", m)
	return buf[:n], nil
	//return base64.StdEncoding.DecodeString( data )*/
	//return base64.StdEncoding.DecodeString( data )
}

// format: <base64-encoded-salt>:<password>
func SplitWithSalt( password string ) ([]byte, []byte, error) {
	parts := strings.Split( password, ":" )
	if len(parts) < 2 {
		return nil, nil, fmt.Errorf("no salt supplied")
	} else if len(parts) > 2 {
		// consider the first ':' is a delimeter
		parts[1] = strings.Join(parts[1:], ":")
	}
	saltBytes, err := base64.StdEncoding.DecodeString( parts[0] )
	if err != nil {
		return nil, nil, err
	}

	return []byte( parts[1] ), saltBytes, nil
}

// derive encryption key from password. used for local configuration storage
func DeriveKey( password, saltBytes []byte ) []byte {
	/*
	 * the draft RFC recommends time=3 and memory=32*1024 (32 MB) is a sensible number.
	 */
	threads := uint8(runtime.NumCPU())
	key := argon2.Key( password, saltBytes, 3, 32 * 1024, threads, SymKeySize )
	return key
}

func DeriveSharedSecret( pkss, ecss []byte ) []byte {
	/*threads := uint8( runtime.NumCPU() )
	key := argon2.Key( pkss, ecss, 3, 32 * 1024, threads, SymKeySize )
	return key */
	hash := sha512.New
	hkdf := hkdf.New( hash, pkss, ecss, nil )
	key := make( []byte, SymKeySize )
	if _, err := io.ReadFull( hkdf, key ); err != nil {
		return nil
	}
	return key
}

// align data to specified packet size
// requirements: packetSize must be bigger or equal than len(data) + 8
func PackData( data []byte, packetSize uint ) ([]byte, error) {
	
	if packetSize < uint(len(data) + 8) {
		return nil, fmt.Errorf("[PackData] packet is too small (%d bytes)", len(data) + 8 )
	}

	buf := make( []byte, packetSize )
	binary.LittleEndian.PutUint64( buf, uint64(len(data)) )
	copy( buf[8:], data )
	if _, err := rand.Read( buf[8 + len(data):] ); err != nil {
		return nil, err
	}
	return buf, nil
}

// reverse function
func UnpackData( data []byte, packetSize uint ) ([]byte, error) {
	// check if we have anything to unpack
	if data == nil || len(data) == 0 {
		return nil, fmt.Errorf("Nothing to unpack")	// should we return an error?
	}

	// i don't really know if it is neccessary but still...
	if uint(len(data)) != packetSize {
		return nil, fmt.Errorf("[UnpackData] (where is packet?)")
	}
	length := binary.LittleEndian.Uint64( data[:8] )
	if 8 + length <= uint64(len( data )) {
		return data[ 8 : 8+length ], nil
	}
	return nil, fmt.Errorf("[UnpackData] data is too short.")
}
