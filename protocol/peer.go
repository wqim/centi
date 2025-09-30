// peer refactored
package protocol
import (
	"fmt"
	"sync"
	//"bytes"
	"crypto/ecdh"
	"crypto/x509"
	"encoding/json"
	"encoding/base64"
	//"encoding/hex"	// debug-only
	"github.com/cloudflare/circl/kem/kyber/kyber768"

	"centi/util"
	"centi/cryptography"
)

// structure, describing our peer
type Peer struct {
	Alias		string	// alias of peer, usually, their username
	Platform	string	// do we really need this???
	key		[]byte	// key for symmetric encryption
	// public keys of the peer
	ecPk		*ecdh.PublicKey
	pqPk		*kyber768.PublicKey
	// thread safety, of course
	pkBytes		[]byte
	mtx		sync.Mutex
}

func NewPeer( alias string ) *Peer {
	return &Peer{
		alias,
		"",
		nil,
		nil,
		nil,
		nil,
		sync.Mutex{},
	}
}

func(p *Peer) ValidSymKey() bool {
	return p.key != nil && len(p.key) == cryptography.SymKeySize
}


func(p *Peer) GetKey() []byte {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	return p.key
}

func(p *Peer) GetPublicKey() []byte {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	return p.pkBytes
}

// thread-safe alias getter
func(p *Peer) GetAlias() string {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	return p.Alias
}


// thread-safe alias setter
func(p *Peer) SetAlias( alias string ) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.Alias = alias
}

// the same things for platform
func(p *Peer) GetPlatform() string {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	return p.Platform
}

func(p *Peer) SetPlatform( platform string ) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.Platform = platform
}

// set key for symmetrical encryption
func(p *Peer) SetKey( key []byte ) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if len(key) != cryptography.SymKeySize {
		return fmt.Errorf("Invalid symmetric key size: %d", len(key))
	}
	p.key = key
	//util.DebugPrintln("[SetKey] Shared secret:", hex.EncodeToString(p.key))
	return nil
}

// unpack and set public key of peer
func(p *Peer) SetPk( pk, networkKey []byte ) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if len(pk) < cryptography.HashSize + cryptography.PkSize {
		return fmt.Errorf("invalid size of public key")
	}

	// if we are able to verify the peer's public key, do this.
	if networkKey != nil && len( networkKey ) == cryptography.SymKeySize {
		toCount := pk[ : len(pk) - cryptography.HashSize ]
		hmac := pk[ len(pk) - cryptography.HashSize : ]
		if cryptography.VerifyHMACBytes( toCount, networkKey, hmac ) == false {
			return fmt.Errorf("Invalid HMAC")
		}
		p.pkBytes = toCount
	} else {
		p.pkBytes = pk
	}

	util.DebugPrintln( "[Peer::SetPk] public key bytes:",
		//base64.StdEncoding.EncodeToString( pk[:16] ),
		base64.StdEncoding.EncodeToString( pk[ kyber768.PublicKeySize : len(pk) - cryptography.HashSize ] ),
	)
	p.pqPk = &kyber768.PublicKey{}
	p.pqPk.Unpack( pk[:kyber768.PublicKeySize] )

	ecPk, err := x509.ParsePKIXPublicKey( pk[ kyber768.PublicKeySize : len(pk) - cryptography.HashSize ] )
	if err != nil {
		return err
	}
	var ok bool
	p.ecPk, ok = ecPk.(*ecdh.PublicKey)
	if !ok {
		return fmt.Errorf("Peer::SetPk: invalid public key format")
	}
	return nil
}

// the only need for this function is fixing a bug
// with random peer alias in case our known peer
// has changed key before we noticed it and succeded in key exchange
// we are authenticating our peer by their public key (useful only if we know a person)
func(p *Peer) SetEccPk( pk []byte ) error {

	p.mtx.Lock()
	defer p.mtx.Unlock()

	ecPk, err := x509.ParsePKIXPublicKey( pk )
	if err != nil {
		return err
	}
	var ok bool
	p.ecPk, ok = ecPk.(*ecdh.PublicKey)
	if !ok {
		return fmt.Errorf("Peer::SetPk: invalid public key format")
	}
	p.pkBytes = pk

	util.DebugPrintf("[IMPORTANT] %s's pk: %s\n", p.Alias, base64.StdEncoding.EncodeToString( p.pkBytes ) )
	return nil
}

// check if supplied public key equals to peer's public key
// this function compares only ecdh public keys because
// we are not able to always know the kyber768 key of our peer.
func(p *Peer) Equals( pk []byte ) bool {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if p.ecPk == nil {	// nothing to do
		//util.DebugPrintln("ecPk = nil")
		return false
	}

	ecPk, err := x509.ParsePKIXPublicKey( pk )
	if err != nil {
		//util.DebugPrintln("ParsePKIX failed")
		return false
	}
	ecKey, ok := ecPk.(*ecdh.PublicKey)
	if !ok {
		//util.DebugPrintln("pointer convertation failed")
		return false
	}
	if p.ecPk.Equal( ecKey ) == false {
		//util.DebugPrintln("...this is really failed...?")
		return false
	}
	// todo: check the hash of public key
	util.DebugPrintln("[Peer::Equals] Length of public key supplied:", len(pk), "/", len(p.pkBytes) )
	return true
}

func(p *Peer) EncapsulateAndPack(
	ourEcSk *ecdh.PrivateKey,
	packetSize uint,
	networkSubkey []byte ) ([]byte, []byte, error) {

	p.mtx.Lock()
	defer p.mtx.Unlock()

	finalSS := []byte{}
	if p.pqPk != nil {
		// encapsulate shared secret
		ct := make([]byte, kyber768.CiphertextSize)
		ss := make([]byte, kyber768.SharedKeySize)
		p.pqPk.EncapsulateTo( ct, ss, nil )
		// pack
		ecPk, err := x509.MarshalPKIXPublicKey( ourEcSk.Public() )
		if err != nil {
			return nil, nil, err
		}
		// encrypt the marshalled version of public key
		ctPk, err := cryptography.Encrypt( ecPk, ss )
		if err != nil {
			return nil, nil, err
		}
		// check if we are able to generate final shared secret
		// right now
		if p.ecPk != nil {	// should always be true
			ss2, err := ourEcSk.ECDH( p.ecPk )
			if err != nil {
				return nil, nil, err
			}
			finalSS = cryptography.DeriveSharedSecret( ss, ss2 )
			/*util.DebugPrintln("[EncapsulateAndPack] Shared secret:",
				hex.EncodeToString( finalSS )) */
		}
		hmac := cryptography.HMACBytes( ecPk, networkSubkey )
		packet := append( ct, ctPk... )
		packet = append( packet, hmac... )
		//util.DebugPrintln( "TOTAL SIZE OF ENCAPSULATED PACKET:", len(packet) )
		packed, err := cryptography.PackData( packet, packetSize )
		return finalSS, packed, err
	}
	return nil, nil, fmt.Errorf("Where is PQS public key?")
}

// cozy wrappers
func(p *Peer)Encrypt( data []byte ) ([]byte, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	return cryptography.Encrypt( data, p.key )
}

func(p *Peer)Decrypt( data []byte ) ([]byte, error) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	return cryptography.Decrypt( data, p.key )
}

func(p *Peer)pack( data []byte, packetSize uint, packetType uint8, seqNumber uint64 ) ([][]byte, error) {
	// returns the slice/array of packets to send
	// if size of data is enought to fit into one packet, returns just one packet's content
	// or error if anything went wrong

	if packetSize < MinPacketSize {
		return nil, fmt.Errorf("invalid packet size")
	}

	p.mtx.Lock()
	defer p.mtx.Unlock()

	isCompressed, data, err := Compress( data )
	if err != nil {
		return nil, err
	}

	result := [][]byte{}
	// split data onto smaller packets just for sure
	//step := uint(packetSize / 2)
	step := util.CalculateDataSize( data, packetSize )
	amountOfPackets := ( uint(len(data)) + step - 1 ) / step

	util.DebugPrintln("Amount of packets:", amountOfPackets)
	util.DebugPrintln("Size of data:", len(data) )

	for i := uint(0); i < uint(len(data)); i += step {
		// handle out of range slice
		border := i + step
		if border > uint(len(data)) {
			border = uint(len(data))
		}

		toEncrypt := make([]byte, border - i) //data[i:border]
		//util.DebugPrintln("Made a buffer")
		copy( toEncrypt, data[i:border] )

		//util.DebugPrintln("Len(toEncrypt) = ", len(toEncrypt), "len(data) =", len(data[i:border]) )
		//util.DebugPrintln("Copied into buffer")
		// pack data
		//util.DebugPrintln("Data in the start:", data[i:border])
		packed, err := PackData(
			packetType,
			isCompressed,
			seqNumber,
			uint64(amountOfPackets),
			toEncrypt,
			p.key )

		if err != nil {
			return nil, err
		}
		seqNumber++

		//util.DebugPrintln("protocol::PackData is fine")
		// add some offsets
		packed, err = cryptography.PackData(
			packed,
			packetSize - cryptography.TagSize - cryptography.NonceSize,
		)
		if err != nil {
			return nil, err
		}

		//util.DebugPrintln("cryptography::PackData is fine")
		// encrypt data
		/*p.mtx.Unlock()
		ct, err := p.Encrypt( packed ) */
		ct, err := cryptography.Encrypt( packed, p.key )
		if err != nil {
			return nil, err
		}

		//util.DebugPrintln("Data in the finish:", data[i:border])
		//p.mtx.Lock()
		//util.DebugPrintln("Encrypt was fine")
		result = append( result, ct )
	}

	util.DebugPrintln("Total packets:", len(result))
	return result, nil
}

func(p *Peer)PackToSend( data []byte, packetSize uint ) ([][]byte, error) {
	return p.pack(data, packetSize, DataPct, 0 )
}

func(p *Peer) PackToResend( data []byte, packetSize uint, amountOfResend uint8, peers []*Peer ) ([][]byte, error) {
	// pack data for final receiver
	packets, err := p.PackToSend( data, packetSize )
	if err != nil {
		return nil, err
	}
	for i := uint8(0); i < amountOfResend; i++ {
		seqNumber := uint64(0)
		tmpPackets := [][]byte{}
		peerIdx := util.RandInt( len(peers) )
		peer := peers[peerIdx]
		
		for _, packet := range packets {
			newPackets, err := peer.pack( packet, packetSize, RetransmitPct, seqNumber )
			if err != nil {
				return nil, err
			}
			seqNumber += uint64(len(newPackets))
			tmpPackets = append( tmpPackets, newPackets... )
		}

		packets = tmpPackets
	}
	return packets, nil
}


func(p *Peer) Unpack( data []byte, packetSize uint ) (*Packet, error) {
	// returns data, should we resend packet or not, and an error, if any occured
	if packetSize < MinPacketSize {
		return nil, fmt.Errorf("invalid packet size")
	}

	pt, err := p.Decrypt( data )
	if err != nil {
		// packet is not meant for us, resend it
		return nil, err
	}

	pt, err = cryptography.UnpackData( pt, packetSize - cryptography.TagSize - cryptography.NonceSize )
	if err != nil {
		return nil, err
	}

	// try to unpack data
	return UnpackDataToPacket( pt, p.key )
}

func(p *Peer) PackPkRequest( packetSize uint ) ([]byte, error) {
	data, err := cryptography.GenRandom(10)	// any dummy here
	if err != nil {	// i don't really think it can happen, but still...
		data = []byte("public-key-request")
	}

	packets, err := p.pack( data, packetSize, PkReqPct, 0 )
	if err != nil {
		return nil, err
	}
	// we really need only 1 packet because we don't handle any data here...
	// maybe I'll change it later
	return packets[0], nil
}

func(p *Peer) PackPublicKeys( peers []*Peer, packetSize uint ) ([][]byte, error) {
	//util.DebugPrintln("Peer::PackPublicKeys")
	// pack cleartext data and encrypt it, than pack into single packets
	knownPks := []map[string]string{}
	for _, peer := range peers {
		
		peersPk := peer.GetPublicKey()
		if len( peersPk ) > 0 {
			knownPks = append( knownPks, map[string]string{
				//"platform": peer.Platform,
				"alias": peer.Alias,
				"content": base64.StdEncoding.EncodeToString( peersPk ),
			})
		}
	}
	data, err := json.Marshal( knownPks )
	if err != nil {
		util.DebugPrintln("Failed to marshal public keys:", err)
		return nil, err
	}
	return p.pack( data, packetSize, PkPct, 0 )
}

func(p *Peer) UnpackPublicKeys( data []byte, packetSize uint ) ([]KnownPk, error) {
	// unpacking cleartext data with public keys
	//util.DebugPrintln("Peer::UnpackPublicKeys")
	var knownPks []KnownPk
	if err := json.Unmarshal( data, &knownPks ); err != nil {
		util.DebugPrintln("Failed to unmarshal public keys:", err)
		return nil, err
	}
	return knownPks, nil
}
