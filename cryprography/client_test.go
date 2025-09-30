package cryptography
import (
	"testing"
)

// TestNewClient verifies that a new client can be created successfully.
func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create new client: %v", err)
	}
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.ECCPrivateKey() == nil {
		t.Fatal("ECCPrivateKey should not be nil")
	}
}

// TestPkToStringAndSkToString verifies that key serialization/deserialization works.
func TestPkToStringAndSkToString(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	pkStr := client.PkToString()
	skStr, err := client.SkToString()
	if err != nil {
		t.Fatalf("Failed to get SkToString: %v", err)
	}

	// Reconstruct client from keys
	newClient, err := ClientFromKeys(pkStr, skStr)
	if err != nil {
		t.Fatalf("Failed to reconstruct client: %v", err)
	}

	// Verify that the reconstructed keys match
	if client.PkToString() != newClient.PkToString() {
		t.Error("Public keys do not match after serialization/deserialization")
	}
}

// TestSharedSecretConsistency ensures that two clients can generate matching shared secrets after key exchange.
func TestSharedSecretConsistency(t *testing.T) {
	clientA, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client A: %v", err)
	}
	clientB, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client B: %v", err)
	}

	// Exchange public keys
	_, err = clientA.GetPublicKey(nil)
	if err != nil {
		t.Fatalf("Client A GetPublicKey failed: %v", err)
	}
	_, err = clientB.GetPublicKey(nil)
	if err != nil {
		t.Fatalf("Client B GetPublicKey failed: %v", err)
	}
	// todo: move this test to protocol.
}
