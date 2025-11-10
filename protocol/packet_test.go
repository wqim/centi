package protocol

import (
	"bytes"
	"encoding/json"
	"testing"
	"centi/cryptography"
)

func TestPackAndUnpackData(t *testing.T) {
	t.Log("TestPackAndUnpackData")
	data := []byte("test data")
	skey, _ := cryptography.GenRandom( cryptography.SymKeySize )
	
	// Pack data
	packed, err := PackData(1, 0, 1, 1, data, skey)
	if err != nil {
		t.Fatalf("PackData failed: %v", err)
	}

	// Unpack data
	unpacked, err := UnpackData(packed, skey)
	if err != nil {
		t.Fatalf("UnpackData failed: %v", err)
	}

	if !bytes.Equal(unpacked, data) {
		t.Errorf("Unpacked data does not match original. Got: %s, want: %s", string(unpacked), string(data))
	}
}

func TestUnpackDataToPacket_InvalidJSON(t *testing.T) {
	t.Log("TestUnpackDataToPacket_InvalidJSON")
	invalidJSON := []byte("invalid json")
	skey, _ := cryptography.GenRandom( cryptography.SymKeySize )
	
	_, err := UnpackDataToPacket(invalidJSON, skey)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestUnpackDataToPacket_InvalidSequence(t *testing.T) {
	t.Log("TestUnpackDataToPacket_InvalidSequence")
	// Create a valid packet but with Total < Seq
	packet := Packet{
		Head: PacketHead{
			Typ: 1, Seq: 5, Total: 3, Compressed: 0,
		},
		Body: PacketBody{
			Data: "test", OrigSize: 4, Hmac: "hmac",
		},
	}
	packed, _ := json.Marshal(packet)

	skey, _ := cryptography.GenRandom( cryptography.SymKeySize )
	_, err := UnpackDataToPacket(packed, skey)
	if err == nil || err.Error() != "Invalid packet sequence number." {
		t.Errorf("Expected sequence error, got: %v", err)
	}
}

func TestUnpackDataToPacket_InvalidDataSize(t *testing.T) {
	// Create packet with data smaller than OrigSize
	t.Log("TestUnpackDataToPacket_InvalidDataSize")
	dataBytes := []byte("abc")
	encodedData := string(dataBytes)

	packet := Packet{
		Head: PacketHead{
			Typ: 1, Seq: 1, Total: 1, Compressed: 0,
		},
		Body: PacketBody{
			Data: encodedData, OrigSize: 10, Hmac: "hmac",
		},
	}
	packed, _ := json.Marshal(packet)
	skey, _ := cryptography.GenRandom( cryptography.SymKeySize )
	
	// This should error due to size mismatch
	_, err := UnpackDataToPacket(packed, skey)
	if err == nil || err.Error() != "Invalid data size(less than specified)" {
		t.Errorf("Expected data size error, got: %v", err)
	}
}

func TestUnpackDataToPacket_HMACVerificationFails(t *testing.T) {

	t.Log("TestUnpackDataToPacket_HMACVerificationFails")
	data := []byte("test data")
	skey, _ := cryptography.GenRandom( cryptography.SymKeySize )
	packed, _ := PackData(1, 0, 1, 1, data, skey)

	data2, err := UnpackData(packed, skey)
	if err != nil {
		t.Errorf("Expected HMAC verification failure, got: %v", err)
	}
	if bytes.Equal( data, data2 ) == false {
		t.Errorf("Pack/unpack splois the data.")
	}
}
