package centi_ssh
import (
	"fmt"
	"net"
	"os"
	"sync"
	"bytes"
	"strconv"
	"strings"
	"golang.org/x/crypto/ssh"

	"centi/util"
	"centi/config"
	"centi/protocol"
)

var (
	// ssh connection can be used without steganography
	// so leave this empty.
	SupportedExt = []string{}
	// some global variables I really want to get rid of.
	received = make(chan *protocol.Message, 100)
	toSend = make(chan *protocol.Message, 100)
	publicKeys = make(chan *protocol.KnownPk, 100)

	publicKey []byte
	publicKeyMtx sync.RWMutex

	// amount of peers we sent message to.
	connected = uint(0)
	connectedMtx sync.RWMutex
	sentTo = 0
	sentToMtx sync.RWMutex

	// current message and a mutex for it.
	currentMessage	*protocol.Message
	currentMtx	sync.RWMutex

)

type Peer struct {
	conn		ssh.Conn
	session		ssh.Channel //*ssh.Session
	pk		*protocol.KnownPk
	hasPk		bool		// does the peer have our public key?
	tmpFolder	string		// temporary folder for messages files
	mtx		sync.RWMutex	// mutex for peer
}

type SshConn struct {
	// our connections :3
	peers		*[]*Peer
	peersMtx	sync.RWMutex
}

func NewSshConn( args map[string]string, channels []config.Channel ) (protocol.Connection, error) {

	bufferSize, err := strconv.Atoi( args["buffer_size"] )
	if err != nil {
		return nil, err
	}

	// create channels with specified channel capacity
	capacity, err := strconv.Atoi( args["channel_capacity"] )
	if err != nil {
		return nil, err
	}

	received = make( chan *protocol.Message, capacity )
	toSend = make( chan *protocol.Message, capacity )
	publicKeys = make( chan *protocol.KnownPk, capacity )

	// run as a server if we should
	if strings.ToLower( args["run_server"] ) == "true" {

		// launch a server
		listener, conf, err := startServer(
			args["address"],
			args["id_rsa"],
			args["authorized_keys"],
			args["valid_credentials"] )

		if err != nil {
			return nil, err
		}

		// check how many connections we are able to serve
		maxConnections, err := strconv.Atoi( args["max_connections"] )
		if err != nil {
			return nil, err
		}
		// handle incoming connections in the separate thread
		go handleConnections( uint(bufferSize), listener, conf, uint(maxConnections) )
	}

	// run as a client: connect to known nodes
	peers := []*Peer{}
	// read and parse known hosts file.
	raw, err := os.ReadFile( args["known_hosts"] )
	if err != nil {
		return nil, err
	}
	knownHostsMap := map[string]ssh.PublicKey{}
	for len(raw) > 0 {
		_, hosts, pubKey, _, tmp, err := ssh.ParseKnownHosts( raw )
		if err != nil {
			return nil, err
		}
		for _, host := range hosts {
			knownHostsMap[host] = pubKey
		}
		raw = tmp
	}

	util.DebugPrintln("Parsed known hosts file. Starting to connect...")
	for _, ch := range channels {

		config := &ssh.ClientConfig{
			User: ch.Args["user"],
			Auth: []ssh.AuthMethod{
				ssh.Password( ch.Args["password"] ),
			},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// check if public key is known.
				for _, pk := range knownHostsMap {
					if bytes.Equal( key.Marshal(), pk.Marshal() ) == true {
						return nil
					}
				}
				return fmt.Errorf("Failed to find public key in known hosts file!")
			},
		}

		// setup ssh connection
		conn, err := ssh.Dial( "tcp", ch.Name, config )
		if err == nil {
			// setup channel
			channel, requests, err := conn.Conn.OpenChannel("customdata", nil)
			if err != nil {
				util.DebugPrintln("Failed to open a channel:", err)
				return nil, err
			}
			go handleIncomingMessages( conn, channel, requests )
			
			// setup a session
			//session, err := conn.NewSession()
			if err == nil {
				//session.Stdout = bytes.NewBuffer([]byte{})
				//session.Stdin = bytes.NewBuffer([]byte{})
				//session.Stderr = bytes.NewBuffer([]byte{})
				// add a peer connection*/
				peers = append( peers, &Peer{
					conn,
					channel, //session,
					nil,
					false,
					"",
					sync.RWMutex{},
				})
			} else {
				util.DebugPrintln("[ssh:NewSshConn] Failed to start a new session:", err)
			}
		} else {
			util.DebugPrintln("[ssh:NewSshConn] Client failed to connect:", err)
		}
	}

	// build a connection
	sconn := SshConn{
		&peers,
		sync.RWMutex{},
	}
	util.DebugPrintln("Created a new conn.")
	// return the connection
	return sconn, nil
}

// unused things
func(s SshConn) PrepareToDelete( data []byte ) (*protocol.Message, error ) {
	return nil, nil
}

func(s SshConn) Delete( msg *protocol.Message ) error {
	return nil
}

func(s SshConn) InitChannels() error {
	return nil
}

func(s SshConn) DeleteChannels() error {
	return nil
}

// used things
func(s SshConn) Name() string {
	return Name
}

func(s SshConn) MessageFromBytes( data []byte ) (*protocol.Message, error) {
	msg := &protocol.Message{
		"",
		Name,
		data,
		protocol.UnknownSender,
		false,	// does not really matter here
		map[string]string{},
	}
	return msg, nil
}

func(s SshConn) DistributePk( p *config.DistributionParameters, pk []byte ) error {
	// first message is always a public key

	publicKeyMtx.Lock()
	publicKey = pk
	publicKeyMtx.Unlock()

	for _, peer := range *s.peers {
		if peer != nil {
			if peer.hasPk == false {
				// TODO: send our public key to peer
				peer.hasPk = true
				n, err := peer.session.SendRequest( PkRequestType, false, pk )
				util.DebugPrintln("SendRequest returned", n, err)
			}
		} else {
			util.DebugPrintln("[DistributePk] peer.session stdout is nil :(")
		}
	}
	return nil
}

func(s SshConn) CollectPks( p *config.DistributionParameters ) ([]protocol.KnownPk, error) {
	// first message is always a public key
	s.peersMtx.RLock()
	defer s.peersMtx.RUnlock()
	
	pks := []protocol.KnownPk{}

	for len( publicKeys ) > 0 {
		pks = append( pks, *(<- publicKeys) )
	}
	// backup the keys.
	for _, pk := range pks {
		publicKeys <- &pk
	}

	for _, peer := range *s.peers {
		if peer.pk != nil {
			pks = append( pks, *peer.pk )
		}
	}
	return pks, nil
}

func(s SshConn) Send ( msg *protocol.Message ) error {

	//buf := bytes.NewBuffer( msg.Data )
	var finalError error
	
	// send a message as server
	connectedMtx.RLock()
	// check if we have at least one active connection at this time.
	if connected > 0 {
		toSend <- msg
	}
	connectedMtx.RUnlock()

	// send message to nodes we are connected to.
	s.peersMtx.RLock()
	defer s.peersMtx.RUnlock()
	for _, peer := range *s.peers {
		// sending data over request payload.
		if peer != nil {
			// do not require it to be replyed
			peer.session.SendRequest( MsgRequestType, false, msg.Data )
		} else {
			util.DebugPrintln("[ssh:Send] peer session stdout is nil :C")
		}
	}

	return finalError
}

func(s SshConn) RecvAll() ([]*protocol.Message, error) {
	// get all the messages from the channel
	msgs := []*protocol.Message{}

	util.DebugPrintln("Len(s.received):", len(received))
	for len(received) > 0 {
		msg := <- received
		msgs = append( msgs, msg )
	}
	return msgs, nil
}

func(s SshConn) GetSupportedExtensions() []string {
	return SupportedExt
}
