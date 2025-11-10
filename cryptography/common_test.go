package cryptography
import (
	"bytes"
	"testing"
)


func TestHashing( t *testing.T ) {
	// generate a key for HMAC
	key, err := GenRandom( SymKeySize )
	if err != nil {
		t.Errorf("Failed to generate encryption key: %s", err.Error())
	}

	// test data
	origData := [][]byte{
		nil,
		[]byte{},
		[]byte("Hello world!"),
	}
	// test hashing and hmac
	for _, orig := range origData {
		hash := Hash( orig )
		if VerifyHash( orig, hash ) == false {
			t.Errorf("Failed to verify hash for %v", orig)
		}
		hmacBytes := HMACBytes( orig, key )
		if VerifyHMACBytes( orig, key, hmacBytes ) == false {
			t.Errorf("Failed to verify hmac bytes for %v", orig)
		}

		hmac := HMAC( orig, key )
		if VerifyHMAC( orig, key, hmac ) == false {
			t.Errorf("Failed to verify hmac for %v", orig)
		}
	}
}

func TestEncryption( t *testing.T ) {
	// generate encryption key
	key, err := GenRandom( SymKeySize )
	if err != nil {
		t.Errorf("Failed to generate encryption key: %s", err.Error())
	}
	// test data
	origData := [][]byte{
		nil,
		[]byte{},
		[]byte("Hello world!"),
	}
	// just run test for each type of possible data...
	for _, orig := range origData {
		ct, err := Encrypt( orig, key )
		if err != nil {
			t.Errorf("Failed to encrypt: %s", err.Error())
		}
		pt, err := Decrypt( ct, key )
		if err != nil {
			t.Errorf("Failed to decrypt: %s", err.Error())
		}
		if bytes.Equal( pt, orig ) == false {
			t.Errorf("[CRITICAL] Encryption changed data: %v != %v", orig, pt)
		}
	}
}

func TestEncoding( t *testing.T ) {

	// yes, i know my crypto module is mostly a wrapper
	// but i will create tests for almost all the functions
	// just because i can.
	data := [][]byte{
		nil,
		[]byte{},
		[]byte("Hello world"),
	}
	for _, d := range data {
		enc := EncodePublicKey( d )
		dec, err := DecodePublicKey( enc )
		if err != nil {
			t.Errorf("Failed to decode pk for %v : %v (%s)", d, enc, err.Error())
		}

		if bytes.Equal(dec,d) == false {
			t.Errorf("Failed to decode data correctly: %v != %v", d, dec)
		}

		enc = EncodeData( d )
		dec, err = DecodeData( enc )
		if err != nil {
			t.Errorf("Failed to decode data for %v: %v (%s)", d, enc, err.Error())
		}
	}
}

func TestPacking( t *testing.T ) {

	packetSizes := []uint{
		24,
		32,
		64,
		128,
	}

	invalidPacketSizes := []uint{ 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15 }

	data := [][]byte{
		//nil,
		//[]byte{},
		[]byte("Hello world"),
	}

	for _, d := range data {

		for _, packetSize := range packetSizes {
			packed, err := PackData( d, packetSize )
			if err != nil {
				t.Errorf("Failed to pack data: %v with packet size %d: %s", d, packetSize, err.Error())
			}
			unpacked, err := UnpackData( packed, packetSize )
			if err != nil {
				t.Errorf("Failed to unpack data: %v with packet size %d: %s", d, packetSize, err.Error())
			}
			if d != nil && bytes.Equal( unpacked, d ) == false {
				t.Errorf("Unpacked != original:\n%v != %v", unpacked, d )
			}
		}

		for _, packetSize := range invalidPacketSizes {
			packed, err := PackData( d, packetSize )
			if err == nil {
				t.Errorf("[invalid] Failed to pack data: %v with packet size %d.", d, packetSize )
			}
			_, err = UnpackData( packed, packetSize )
			if err == nil {
				t.Errorf("[invalid] Failed to unpack data: %v with packet size %d.", d, packetSize )
			}
		}
	}

}

func TestOther( t *testing.T ) {
	// todo: add a fixed test value.
	password := []byte("password")
	key := DeriveKey( password )
	if len(key) != SymKeySize {
		t.Errorf("Invalid size of output key: %d", len(key))
	}

	// generate keys for derivation
	key1, err := GenRandom( SymKeySize )
	if err != nil {
		t.Errorf("Failed to generate encryption key: %s", err.Error())
	}

	key2, err := GenRandom( SymKeySize )
	if err != nil {
		t.Errorf("Failed to generate encryption key: %s", err.Error())
	}

	ss := DeriveSharedSecret( key1, key2 )
	if len(ss) != SymKeySize {
		t.Errorf("Failed to generate shared secret.")
	}
}
