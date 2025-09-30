package protocol

import (
	//"fmt"
	"bytes"
	"testing"

	"centi/cryptography"
)

func TestNewPeer(t *testing.T) {
	p := NewPeer("test")
	if p.Alias != "test" {
		t.Errorf("Expected alias 'test', got %s", p.Alias)
	}
}

func TestSetAndGetKey(t *testing.T) {
	peer := NewPeer("user")
	key := make([]byte, cryptography.SymKeySize)
	copy(key, []byte("12345678901234567890123456789012")) // 32 bytes
	err := peer.SetKey(key)
	if err != nil {
		t.Fatal(err)
	}
	retrieved := peer.GetKey()
	if !bytes.Equal(retrieved, key) {
		t.Errorf("Keys do not match")
	}
}

func TestValidSymKey(t *testing.T) {
	peer := NewPeer("user")
	if peer.ValidSymKey() {
		t.Error("Should be invalid when key is nil")
	}
	key := make([]byte, cryptography.SymKeySize)
	peer.key = key
	if !peer.ValidSymKey() {
		t.Error("Should be valid when key is set properly")
	}
}



func TestEncryptDecrypt(t *testing.T) {
	peer := NewPeer("user")
	data := []byte("Hello World")
	// Set a key
	key := make([]byte, cryptography.SymKeySize)
	copy(key, []byte("12345678901234567890123456789012"))
	if err := peer.SetKey(key); err != nil {
		t.Fatal(err)
	}

	encrypted, err := peer.Encrypt(data)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := peer.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decrypted, data) {
		t.Errorf("Decrypted data does not match original")
	}
}

func TestPackAndUnpack(t *testing.T) {
	peer := NewPeer("user")
	// Set key
	key := make([]byte, cryptography.SymKeySize)
	if err := peer.SetKey(key); err != nil {
		t.Fatal(err)
	}
	data := []byte("This is a test message for packing")
	packetSize := uint(4096)

	packets, err := peer.PackToSend(data, packetSize)
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) == 0 {
		t.Fatal("No packets generated")
	}
	// Unpack each packet
	for _, packet := range packets {
		//fmt.Println("Unpacking packet number ", i)
		packetObj, err := peer.Unpack(packet, packetSize)
		if err != nil {
			t.Fatal(err)
		}
		decodedData, err := cryptography.DecodeData( packetObj.Body.Data )
		if err != nil {
			t.Errorf("Failed to decode data: %s", err.Error())
		}
		decodedData = decodedData[:packetObj.Body.OrigSize]
		if packetObj == nil || string(decodedData) != string(data) {
			t.Errorf("Unpacked data mismatch: got '%s', expected '%s'",
				string(decodedData),
				string(data),
			)
		}
	}
}
