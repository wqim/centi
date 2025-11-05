package p2p
import (

	// general things
	"time"
	"sync"
	"context"
	"strconv"
	"strings"

	// libp2p stuff
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/crypto"
	//"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	//"github.com/libp2p/go-libp2p/core/peerstore"
	proto "github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"

	"github.com/multiformats/go-multiaddr"
	//"github.com/multiformats/go-multiaddr/net"

	// for creating a valid Connection interface
	"centi/util"
	"centi/config"
	"centi/protocol"
)

const (
	channelSize = 1024
	moduleName = "p2p"
	Delimeter = ";"
)

var (
	SupportedExt = []string{}
	// i hate the thing i have to make so many global variables
	toSend = make( chan *protocol.Message, channelSize )
	received = make( chan *protocol.Message, channelSize )
	pubKey = []byte{}
	publicKeys = []protocol.KnownPk{}
	pkMtx sync.Mutex
)

type addrList []multiaddr.Multiaddr
type P2PConfig struct {
	//Addrs			addrList
	BootstrapPeers		addrList
	ProtocolID		string
	RendezvousStrings	[]string
}

type P2PConn struct {
	hst		host.Host
	conf		*P2PConfig
	channels	[]config.Channel
}

func NewP2PConn( args map[string]string, channels []config.Channel ) (protocol.Connection, error) {

	addrs := strings.Split( args["addrs"], Delimeter )
	addresses := libp2p.ListenAddrStrings( addrs... )
	
	if len(addrs) == 0 {
		addresses = libp2p.NoListenAddrs
	}

	// generate temporary key pair, only for libp2p things...
	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519,	// key type
		-1,		// select key length when possible (i.e. RSA)
	)

	lowwater, err := strconv.Atoi( args["lowwater"] )
	if err != nil {
		return nil, err
	}

	highwater, err := strconv.Atoi( args["highwater"] )
	if err != nil {
		return nil, err
	}

	delay, err := strconv.Atoi( args["grace_period"] )
	if err != nil {
		return nil, err
	}

	rendezvousStrings := strings.Split( args["rendezvous_strings"], Delimeter )
	util.DebugPrintln("Rendezvous strings:", rendezvousStrings)
	conf := &P2PConfig{
		//nil, //addresses,
		addrList{},
		args["protocol_id"],
		rendezvousStrings,
	}

	util.DebugPrintln( "Configuration:", conf )

	connmgr, err := connmgr.NewConnManager(
		lowwater,	// Lowwater
		highwater,	// Highwater
		connmgr.WithGracePeriod( time.Millisecond * time.Duration( delay ) ),
	)

	hst, err := libp2p.New(
		libp2p.Identity( priv ),
		addresses,
		// support TLS connections
		libp2p.Security( libp2ptls.ID, libp2ptls.New ),
		// support noise connections
		libp2p.Security( noise.ID, noise.New ),
		// support any other default transports (TCP)
		libp2p.DefaultTransports,
		// prevent our peer from having too many connections
		// by attaching a connection manager
		libp2p.ConnectionManager( connmgr ),
		// attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// let this host use DHT to find other hosts
		libp2p.Routing( func(h host.Host) ( routing.PeerRouting, error ) {
			idht, err := dht.New( context.Background(), h )
			return idht, err
		}),
		// if we want to help other peers to figure out if they are behind NATs,
		// launch this server-side of AutoNAT too (AutoRelay already runs the client)
		// should not cause any performance issues.
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
	)
	if err != nil {
		return nil, err
	}
	return P2PConn {
		hst,
		conf,
		channels,
	}, nil
}

// dummies for functions which are not needed for p2p functionality
func(p P2PConn) PrepareToDelete( data []byte ) (*protocol.Message, error) { return nil, nil }
func(p P2PConn) Delete( msg *protocol.Message ) error { return nil }
func(p P2PConn) DeleteChannels() error { return nil }

// these ones are useful
func(p P2PConn) InitChannels() error {
	// this function is called when a peer connects, and
	// starts a stream with this protocol. only applies on the
	// receiving size

	var finalError error
	
	ctx := context.Background()
	bootstrapPeers := make( []peer.AddrInfo, len(p.conf.BootstrapPeers) )
	for i, addr := range p.conf.BootstrapPeers {
		peerinfo, err := peer.AddrInfoFromP2pAddr( addr )
		if err != nil {
			finalError = err
		} else {
			bootstrapPeers[i] = *peerinfo
		}
	}

	kademliaDHT, err := dht.New( ctx, p.hst, dht.BootstrapPeers(bootstrapPeers...) )
	if err != nil {
		finalError = err
	} else {
		if err = kademliaDHT.Bootstrap( ctx ); err != nil {
			finalError = err
		} else {
			// finish bootstrap
			// (really bootstrap should block until it's ready, but that isn't the case yet.)
			//util.DebugPrintln("sleeping for 1 second")
			//time.Sleep( 5 * time.Second )

			util.DebugPrintln("slept.")
			for _, rs := range p.conf.RendezvousStrings {
				routingDiscovery := drouting.NewRoutingDiscovery( kademliaDHT )
				dutil.Advertise( ctx, routingDiscovery, rs )
				util.DebugPrintln("Advertized", rs, " as ", p.hst.ID())
				go func() {
					for {
						peerChan, err := routingDiscovery.FindPeers( ctx, rs )
						if err != nil {
							finalError = err
						} else {
							if len( peerChan ) != 0 {
								util.DebugPrintln("Found", len(peerChan), "peers." )
							}
							for _peer := range peerChan {
								if _peer.ID == p.hst.ID() {
									//util.DebugPrintln( "Found ourselves:", p.hst.ID() )
									continue
								}
								util.DebugPrintln("Found peer in chain:", _peer.ID)
								stream, err := p.hst.NewStream( ctx, _peer.ID, proto.ID(p.conf.ProtocolID) )
								if err != nil {
									finalError = err
								} else {
									handleStream( string(_peer.ID), stream )
								}
							}
						}
					}
				}()
			}
		}
	}
	return finalError
}

func(p P2PConn) DistributePk( dp *config.DistributionParameters, pk []byte ) error {
	// the distribution of public key is fairy simple:
	// we just send it's content (along with hmac, of course)
	// as first message.
	pubKey = pk
	return nil
}

func(p P2PConn) CollectPks( dp *config.DistributionParameters ) ([]protocol.KnownPk, error) {
	// the way we collect public keys is the same...
	pkMtx.Lock()
	defer pkMtx.Unlock()
	return publicKeys, nil
}

func(p P2PConn) Send( msg *protocol.Message ) error {
	toSend <- msg
	return nil
}

func(p P2PConn) RecvAll() ([]*protocol.Message, error) {
	// just read the content of received messages channel
	msgs := []*protocol.Message{}
	for len(received) > 0 {
		msg, ok := <- received
		if ok {
			msgs = append( msgs, msg )
		}
	} 
	return msgs, nil
}

func(p P2PConn) MessageFromBytes( data []byte ) (*protocol.Message, error) {
	msg := &protocol.Message{
		"",
		p.Name(),
		data,
		protocol.UnknownSender,
		false,	// does not really matter here
		map[string]string{},
	}
	return msg, nil
}

func(p P2PConn) Name() string {
	return moduleName
}

func(p P2PConn) GetSupportedExtensions() []string {
	return SupportedExt
}
