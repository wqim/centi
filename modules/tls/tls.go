package centi_tls
import (
	"fmt"
	"net"
	"time"
	"sync"
	"strconv"
	"strings"
	"crypto/tls"
	"encoding/json"
	"encoding/base64"

	"centi/util"
	"centi/config"
	"centi/protocol"
	"centi/cryptography"
)

/*
 * TLS module for support of direct connections over TLS.
 * Improves speed of the network and allows users not to use
 * any totalitaristic platform and be more anonymous.
 *
 * The only requirement is what nodes which has "run_server"
 * option enabled must be accessible in the global network (internet).
 *
 */
var (
	// do not use steganography because
	// our connection is already encrypted by TLS
	// and isn't stored on any servers (except secret services' ones
	// but all the packets are stored there anyway so ignore it).
	SupportedExt = []string{}
	
	// server-only
	publicKeys = make( chan *protocol.KnownPk, 100 )
	received = make( chan *protocol.Message, 100 )
	toSend = make( chan *protocol.Message, 100 )
	connectionsCount = 0
	connectionsMtx = sync.RWMutex{}

	currentMessage *protocol.Message
	sentTo = 0			// amount of server's connections to which we have sent currentMessage
	sentMtx = sync.RWMutex{}	// and of couse mutex for these things

	// both server and client ones.
	publicKey = []byte{}
	pkMtx = sync.RWMutex{}
)

const (
	delimeter = ","
	Name = "tls" // name of the module
)

type NetConfigStr struct {

	RunServer	string	`json:"run_server"`		// "true" or "false"
	PacketSize	string	`json:"packet_size"`		// the same value as for config.NetworkConfig
	Protocol	string	`json:"protocol"`		// network parameter for tls.Listen.
	Certificate	string	`json:"cert_path"`
	Key		string	`json:"key_path"`
	Address		string	`json:"net_addr"`
	
	// future security improvements
	MaxConnections	string	`json:"max_connections"`	// server-only
	//BlockList	string	`json:"block_list"`		// blacklisted users
}

type Peer struct {
	conn	*tls.Conn
	pk	*protocol.KnownPk
	hasPk	bool			// does peer has a public key of us?
	mtx	sync.RWMutex		// peer mutex
}

type NetConn struct {
	server		net.Listener
	bufferSize	uint
	peers		*[]*Peer
	mtx		sync.RWMutex	// mutex for peers
}

func NewNetConn( args map[string]string, channels []config.Channel ) (protocol.Connection, error) {
	
	// check arguments format
	var connection NetConn
	var netconf NetConfigStr
	dumped, err := json.Marshal( args )
	if err != nil {
		return connection, err
	}
	if err = json.Unmarshal( dumped, &netconf ); err != nil {
		return connection, err
	}
	
	// parse the size of packet
	packetSize, err := strconv.Atoi( netconf.PacketSize )
	if err != nil {
		return connection, err
	}

	// i don't really think someone is going to set this value too low
	if packetSize < 16 {
		return connection, fmt.Errorf("Invalid packet size")
	}

	connection.bufferSize = uint(packetSize)

	// setup connection
	if strings.ToLower( netconf.RunServer ) == "true" {

		// if we should run network relay, do this.
		cert, err := tls.LoadX509KeyPair( netconf.Certificate, netconf.Key )
		if err != nil {
			return connection, err
		}
		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		var server net.Listener
		// use any supported network: "tcp", "tcp4", "tcp6", "unix" or "unixpacket".
		server, err = tls.Listen( netconf.Protocol,
					   netconf.Address,
					   config)
		

		if err != nil {
			return connection, err
		}
		connection.server = server
		go func() {
			maxConnections, err := strconv.Atoi( netconf.MaxConnections )
			if err != nil {
				// we need to somehow send an error to user....
				return
			}

			for {
				// accept connections in the background.
				if server != nil {
					connectionsMtx.RLock()
					// if we are able to accept a connection, do this.
					if connectionsCount < maxConnections {
						conn, err := server.Accept()
						if conn != nil {
							util.DebugPrintln("Accepted connection at ", conn.RemoteAddr())
							if err == nil {

								connectionsMtx.RUnlock()
								connectionsMtx.Lock()
								connectionsCount++
								connectionsMtx.Unlock()
								connectionsMtx.RLock()

								go handleConnection( uint(packetSize), conn )
							}
						}
					}
					connectionsMtx.RUnlock()
				}
			}
		}()
	}

	// get list of peers to whom we should connect to...
	peers := channelsToPeers( uint(packetSize), channels )
	
	if len( *peers ) > 1 {
		//publicKeys = make( chan *protocol.KnownPk, len(peers) )
		received = make( chan *protocol.Message, 100 * len(*peers) )
		toSend = make( chan *protocol.Message, 100 * len(*peers) )
	}

	connection.peers = peers //channelsToPeers( channels )
	return connection, nil
}

func handleConnection( bufferSize uint, conn net.Conn ) {
	//util.DebugPrintln("handleConnection()")
	// receive messages in the background...
	defer func() {
		connectionsMtx.Lock()
		connectionsCount--
		connectionsMtx.Unlock()
		conn.Close()
	}()

	sentPk := false
	receivedPk := false

	lastPacket := []byte{} //make([]byte, bufferSize)
	totalReceived := uint(0)

	for {
		// check if we have sent a public key...
		pkMtx.RLock()
		if len( publicKey ) > 0 {
			if sentPk == false {
				// send a public key if not done yet
				util.DebugPrintln("[SERVER] Sending public key:", base64.StdEncoding.EncodeToString(publicKey[:64]))
				conn.Write( publicKey )
				sentPk = true

			} else {
				// send an actual message from the toSend channel.
				// the pain in the ass.
				var msg *protocol.Message
				msg = nil

				sentMtx.Lock()
				connectionsMtx.RLock() 
				if sentTo == 0 || sentTo >= connectionsCount {
					if len( toSend ) > 0 {
						currentMessage = <- toSend
						msg = currentMessage
						sentTo = 0
					}
				} else {
					msg = currentMessage
				}

				if msg != nil {
					util.DebugPrintln("[SERVER] Sending message from channel:") //, string(msg.Data))
					conn.Write( msg.Data )
					sentTo++
				}
				connectionsMtx.RUnlock() 
				sentMtx.Unlock()
			}
		}
		// in another case we did not received
		// a public key yet, therefore we are unable
		// to make a secure connection.
		// do nothing in that case.
		pkMtx.RUnlock()

		//util.DebugPrintln("Original buffer size:", bufferSize)
		buf := make([]byte, bufferSize)
		n, err := conn.Read( buf )
		if err != nil {
			return
		}

		totalReceived += uint(n)
		lastPacket = append( lastPacket, buf[:n]... )

		util.DebugPrintln("[SERVER] Received n =", n, "bytes from client.")
		if receivedPk == false {
			// why do we receive public key by parts
			if totalReceived == cryptography.TotalPkSize {
				// first message is always a public key content
				util.DebugPrintln("[SERVER] Received public key:", totalReceived, "bytes.") //, string(buf))
				pubKey := &protocol.KnownPk{
					Name,
					AddrToAlias( conn.RemoteAddr().String() ),
					lastPacket,
				}
				totalReceived = 0
				lastPacket = []byte{}
				publicKeys <- pubKey
				receivedPk = true
			}
		} else if receivedPk == true && totalReceived == bufferSize {
			// received a full package.
			util.DebugPrintln("[SERVER] Received message:", totalReceived, "bytes." ) //, string(buf))
			// other messages are just normal messages
			received <- &protocol.Message{
				Name,
				lastPacket,
				AddrToAlias( conn.RemoteAddr().String() ),
				false,
				map[string]string{},
			}
			totalReceived = 0
			lastPacket = []byte{}

		} else if totalReceived > bufferSize {
			// received too many data...?
			util.DebugPrintln("Received too many data, closing connection...")
			return
		}
	}
}



// these ones are useful.
func(n NetConn) DistributePk(p *config.DistributionParameters, pk []byte ) error {
	util.DebugPrintln( "DistributePk():", len(pk), "bytes:", base64.StdEncoding.EncodeToString( pk[:64] ) )
	
	for _, peer := range *n.peers {
		if peer.hasPk == false {
			nbytes, err := peer.conn.Write( pk )
			if err == nil {
				util.DebugPrintln("Sent", nbytes, "bytes of public key.")
				peer.hasPk = true
			}
		}
	}

	pkMtx.Lock()
	publicKey = pk
	pkMtx.Unlock()
	return nil
}

func(n NetConn) CollectPks(p *config.DistributionParameters) ([]protocol.KnownPk, error) {
	// just move public keys out of channel
	var finalError error
	pks := []protocol.KnownPk{}

	// the first message sent is always a public key content.
	for _, peer := range *n.peers {
		if peer == nil || peer.conn == nil {
			// todo: remove disconnected peers.
			continue
		}

		//util.DebugPrintln("[CLIENT] Peer:", peer)
		peer.mtx.RLock()

		if peer.pk == nil {

			buf := make([]byte, cryptography.TotalPkSize)
			nbytes := 0

			for {
				m, err := peer.conn.Read( buf[nbytes:] )
				if err != nil {
					closePeer( peer, true )
					finalError = err
				} else {
					nbytes += m
					if nbytes > cryptography.TotalPkSize {
						// invalid case(?)
						util.DebugPrintln("Peer is trying to flood us!")
						break
					}

					if nbytes == cryptography.TotalPkSize {
						util.DebugPrintln("[CLIENT] received public key!") //, string(buf))
						pubKey := &protocol.KnownPk{
							n.Name(),
							AddrToAlias( peer.conn.RemoteAddr().String() ),
							buf,
						}
						peer.pk = pubKey
						publicKeys <- pubKey
						break
					}
				}
			}
		}
		peer.mtx.RUnlock()
	}


	// collect public keys from channel in any case
	util.DebugPrintln("CollectPks()")
	for {
		if len(publicKeys) == 0 {
			break
		}
		val := <- publicKeys
		pks = append( pks, *val )
	}
	// also push a copy of each public key we collected
	// in the channel so we could fetch them again...(?)
	for _, pubKey := range pks {
		publicKeys <- &pubKey
	}
	return pks, finalError
}

func receiveMessagesInBackground( bufferSize uint, peer *Peer ) {
	for {
		if peer == nil {
			return
		}
		peer.mtx.RLock()
		if peer.conn == nil {
			return
		}
		if peer.pk == nil {
			// not a neccessary delay, just not to
			// overwarm the computer.
			peer.mtx.RUnlock()
			time.Sleep( 1 * time.Second )
			continue
		}

		buf := make([]byte, bufferSize)
		nbytes, err := peer.conn.Read( buf )
		if err == nil {
			util.DebugPrintln("Got message from ", peer.conn.RemoteAddr().String())
			msg := &protocol.Message{
				Name,
				buf[:nbytes],
				AddrToAlias( peer.conn.RemoteAddr().String() ),
				false,
				map[string]string{},
			}
			// add message into shared channel
			received <- msg
		} else {
			// release mutex and don't handle any i/o
			// from connection anymore
			closePeer( peer, true )
			return
		}
		peer.mtx.RUnlock()
	}
}

// send message to everyone
func(n NetConn) Send( msg *protocol.Message ) error {

	if len(*n.peers) > 0 {
		util.DebugPrintln("Send(): length =", len(msg.Data), "bytes." )
	}

	// for server-part
	if n.server != nil && len(publicKey) > 0 {
		// public key is set.
		// this thing messes up the server-client tests...
		// also check what there is someone we are able to send
		// data to...
		connectionsMtx.RLock()
		connCount := connectionsCount
		connectionsMtx.RUnlock()

		//util.DebugPrintln("Amount of connections:", connCount)
		if connCount > 0 {
			toSend <- msg
		}
	}

	for _, peer := range *n.peers {
		if peer.hasPk {	// already sent our public key
			peer.conn.Write( msg.Data )
		}
	}
	return nil
}

func(n NetConn) RecvAll() ([]*protocol.Message, error) {
	// just move messages out of channel
	//util.DebugPrintln("RecvAll()")
	msgs := []*protocol.Message{}
	// retrieve all the messages from the channel
	// (both server-side and client-side)
	for {
		if len(received) == 0 {
			return msgs, nil
		}
		msgs = append( msgs, <- received )
	}
}

func channelsToPeers( bufferSize uint, channels []config.Channel ) *[]*Peer {
	peers := []*Peer{}
	for _, peer := range channels {
		parts := strings.Split( peer.Name, ":" )
		if len(parts) == 2 {
			var conf *tls.Config
			// debug-only feature.
			if util.DebugMode == true {
				conf = &tls.Config{
					InsecureSkipVerify: true,
				}
			} else {
				conf = &tls.Config{}
			}

			conn, err := tls.Dial( "tcp", parts[0] + ":" + parts[1], conf )
			if err == nil {
				newPeer := &Peer{
					conn,
					nil,
					false,
					sync.RWMutex{},
				}
				go receiveMessagesInBackground( bufferSize, newPeer )
				peers = append( peers, newPeer )
			} else {
				util.DebugPrintln("[!] Failed to connect to peer:", err)
			}
		}
	}
	util.DebugPrintln("[+] Connected to", len(peers), "peers.")
	return &peers
}
