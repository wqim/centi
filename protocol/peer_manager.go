package protocol
import (
	"sync"
	//"bytes"
	//"centi/util"
	"centi/cryptography"
)

type PeerManager struct {
	peers		[]*Peer
	networkKey	[]byte
	mtx		sync.RWMutex
}

func NewPeerManager(
	networkKey []byte,
	/*networkSubkeys map[string][]byte*/ ) *PeerManager {

	return &PeerManager{
		[]*Peer{},
		networkKey,
		//networkSubkeys,
		sync.RWMutex{},
	}
}

func(p *PeerManager) PeersFromKeys( publicKeys []string ) error {
	// sync of course!
	p.mtx.Lock()
	defer p.mtx.Unlock()

	// build a peer from public keys known.
	for _, pk := range publicKeys {
		// decode public key
		pkBytes, err := cryptography.DecodePublicKey( pk )
		if err != nil {
			return err
		}
		// take a hash of public key as an alias
		alias := cryptography.Hash( pkBytes )
		newPeer := NewPeer( alias )
		// set public key bytes
		if err = newPeer.SetPk( pkBytes, nil ); err != nil { //p.networkKey ); err != nil {
			return err
		}
		p.peers = append( p.peers, newPeer )
	}
	return nil
}

func(p *PeerManager) NetworkKey() []byte {
	return p.networkKey
}

func(p *PeerManager) AddPeer( peer *Peer ) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.peers = append( p.peers, peer )
}

func(p *PeerManager) Exists( alias string ) bool {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	for _, peer := range p.peers {
		if peer.Alias == alias {
			return true
		}
	}
	return false
}

/*
func(p *PeerManager) ExistsWithKey( pk []byte ) bool {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	for _, peer := range p.peers {
		if peer.Equals( pk ) {
			return true
		}
	}
	return false
}
*/

func(p *PeerManager) GetPeerByName( alias string ) *Peer {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	for _, peer := range p.peers {
		if peer.Alias == alias {
			return peer
		}
	}
	return nil
}

/*
func(p *PeerManager) GetPeerByPublicKey( pk []byte ) (int, *Peer) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	for idx, peer := range p.peers {
		if peer.Equals( pk ) {
			return idx, peer
		}
	}
	return -1, nil
}
*/

func(p *PeerManager) GetPeers() []*Peer {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	return p.peers
}

/*
func(p *PeerManager) DropDuplicates() {

	// drop peer duplicates.
	// if peer has updated their public key
	// then we have 2 different peers with similar alias
	// if we didn't have a key exchange with peer yet,
	// just leave the last entry with peer's alias the same

	p.mtx.Lock()
	defer p.mtx.Unlock()

	idx := len( p.peers ) - 1
	for {
		// pick the last element in our peers	
		//duplicateFound := false

		// go over all the peers except the last one
		for i := 0; i < idx; i++ {
			// check if this peer's name
			// is the same as last one has
			//util.DebugPrintln( "PeerManager::DropDuplicates: ", i, idx, p.peers[i].GetAlias() )

			isDuplicate := p.peers[i].GetAlias() == p.peers[idx].GetAlias()

			pk1 := p.peers[i].GetPublicKey()
			pk2 := p.peers[idx].GetPublicKey()

			// compare the public key
			// we compare only elliptic public key
			if len(pk1) > len(pk2) {
				pk1 = pk1[ cryptography.PkSize: ]
			} else if len(pk2) > len(pk1) {
				pk2 = pk2[ cryptography.PkSize: ]
			}

			if bytes.Equal( pk1, pk2 ) == true {
				isDuplicate = true
			} 
			if isDuplicate {
				// we have one, replace it

				// change public key as we are already know it
				p.peers[idx].SetPk( p.peers[i].GetPublicKey(), p.networkKey )
				// also set an encryption key
				p.peers[idx].SetKey( p.peers[i].GetKey() )

				util.DebugPrintln("Found a peer duplicate for ", p.peers[i].GetAlias(),
					", index:", i, "/", idx)

				aliasIdx := p.peers[idx].GetAlias()
				aliasI := p.peers[i].GetAlias()

				// change alias if there is a shorter one.
				if len(aliasI) < len(aliasIdx) {
					p.peers[idx].SetAlias( aliasI )
				}

				//p.peers[idx].SetAlias( p.peers[i].GetAlias() )

				p.peers = append( p.peers[:i], p.peers[i+1:]... )
				break
			} else {
				util.DebugPrintf("WTF? \"%s\" != \"%s\"\n", p.peers[i].GetAlias(),
					p.peers[idx].GetAlias() )
			}
		}

		idx -= 1
		if idx < 0 {
			break
		}
	}
}
*/
