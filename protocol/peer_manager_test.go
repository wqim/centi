package protocol

import (
	"bytes"
	"testing"
	//"centi/cryptography"
)

func TestNewPeerManager(t *testing.T) {
	nk := []byte("test_network_key_1234567890123456")
	pm := NewPeerManager(nk)

	if pm.NetworkKey() == nil {
		t.Error("Network key should not be nil")
	}
	if !bytes.Equal(pm.NetworkKey(), nk) {
		t.Error("Network key mismatch")
	}
	if len(pm.GetPeers()) != 0 {
		t.Error("Expected empty peer list")
	}
}

func TestAddAndGetPeer(t *testing.T) {
	//testCli, _ := cryptography.NewClient()
	key := make([]byte, 32)
	pm := NewPeerManager( key )
	peer := NewPeer("alice")
	pm.AddPeer(peer)

	if len(pm.GetPeers()) != 1 {
		t.Fatal("Peer not added properly")
	}

	retrieved := pm.GetPeerByName("alice")
	if retrieved == nil || retrieved.Alias != "alice" {
		t.Error("Failed to retrieve peer by name")
	}

	/*
	if retrieved.GetPublicKey() != nil {
		idx, peerByPK := pm.GetPeerByPublicKey( retrieved.GetPublicKey() )
		if idx != 0 || peerByPK == nil {
			t.Error("Failed to retrieve peer by public key")
		}
	} */
}

func TestExists(t *testing.T) {
	key := make([]byte, 32)
	pm := NewPeerManager(key)
	peer := NewPeer("bob")
	pm.AddPeer(peer)

	if !pm.Exists("bob") {
		t.Error("Exists should return true for existing alias")
	}
	if pm.Exists("nonexistent") {
		t.Error("Exists should return false for non-existing alias")
	}
}

/*
func TestExistsWithKey(t *testing.T) {
	key := make([]byte, 32)
	pm := NewPeerManager(key)
	peer := NewPeer("charlie")
	pm.AddPeer(peer)

	if pm.ExistsWithKey([]byte("unknown")) {
		t.Error("ExistsWithKey should return false for unknown key")
	}
}*/

func TestDropDuplicates(t *testing.T) {
	pm := NewPeerManager([]byte("network"))
	// Create two peers with same alias and similar public keys

	peer1 := NewPeer("duplicate")
	peer2 := NewPeer("duplicate")

	pm.AddPeer(peer1)
	pm.AddPeer(peer2)

	if len(pm.GetPeers()) != 2 {
		t.Fatal("Should have 2 peers before dropping duplicates")
	}

	// Set the second peer's public key similar to first to simulate duplicate
	peer2.SetPk(peer1.GetPublicKey(), pm.networkKey)
	//pm.DropDuplicates()
	/*
	peers := pm.GetPeers()
	if len(peers) != 1 {
		t.Errorf("Expected 1 peer after dropping duplicates, got %d", len(peers))
	}
	if peers[0].GetAlias() != "duplicate" {
		t.Error("Remaining peer should have alias 'duplicate'")
	}*/
}

func TestConcurrencySafety(t *testing.T) {
	pm := NewPeerManager([]byte("key"))
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			peer := NewPeer("name")
			pm.AddPeer(peer)
		}
		done <- true
	}()

	go func() {
		pm.GetPeers()
		done <- true
	}()

	<-done
	<-done
}
