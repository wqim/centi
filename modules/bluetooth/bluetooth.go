package bluetooth
import (
	"fmt"
	"time"
	//"context"
	"sync"
	"strconv"
	bt "tinygo.org/x/bluetooth"
	"centi/util"
	"centi/config"
	"centi/protocol"
)

// Ah, I'm dreaming about getting rid of all of these global variables.
// there should be an another architect solution for modules coding.
var (
	// do not retransmit any files by bluetooth by default, only packages.
	SupportedExt = []string{}
	// basically, the same as in TLS module
	received = make( chan *protocol.Message, 1 )
	toSend = make( chan *protocol.Message, 1 )
	publicKeys = make( chan *protocol.KnownPk, 1 )
	connectionsCount = 0
	connectionsMtx = sync.RWMutex{}

	currentMessage *protocol.Message
	sentTo = 0
	sentMtx = sync.RWMutex{}

	// more universal
	publicKey = []byte{}
	pkMtx = sync.RWMutex{}
	initialized = false	
)

type Peer struct {
	conn	*bt.Device
	pk	*protocol.KnownPk
	hasPk	bool			// does the peer has our public key?
	mtx	sync.RWMutex		// mutex for this peer.
}

// perfectly we should run as both server and client
// but go bluetooth does not support it currently :(
// i'll try to implement it here...
type BluetoothConn struct {
	// things for bluetooth connection
	adapter		*bt.Adapter
	serverAdapter	*bt.Adapter
	connParams	*bt.ConnectionParams
	connections	[]*bt.Device
	connectionsMtx	*sync.RWMutex
	characteristics	[]bt.DeviceCharacteristic
	knownDevices	[]string
	peers		map[string][]bt.DeviceCharacteristic
	peersMtx	sync.RWMutex

	// things for centi connection
	pk		[]byte
	pkMtx		*sync.RWMutex
	pks		[]*protocol.KnownPk
	pksMtx		*sync.RWMutex

	//toSend		chan *protocol.Message
	// we receive all the messages in the background, so use channel to store them.
	bufferSize	uint
	doAdvertise	bool
}

// basic constructor
// todo: handle user input more gracefully...
func NewBluetoothConn( args map[string]string, channels []config.Channel ) (protocol.Connection, error) {

	if initialized == true {
		return BluetoothConn{}, fmt.Errorf("Bluetooth connection is already initialized.")
	}

	//initialized = true	// as we have a global variable `received`, allow only one bluetooth connection.

	adapter := bt.DefaultAdapter //bt.NewAdapter( args["client_adapter_id"] )
	serverAdapter := bt.DefaultAdapter //bt.NewAdapter( args["server_adapter_id"] )

	if err := adapter.Enable(); err != nil {
		return nil, err
	}
	if err := serverAdapter.Enable(); err != nil {
		return nil, err
	}

	connectionTimeout, err := strconv.Atoi( args["connection_timeout"] )
	if err != nil {
		return nil, err
	}

	minInterval, err := strconv.Atoi( args["min_interval"] )
	if err != nil {
		return nil, err
	}

	maxInterval, err := strconv.Atoi( args["max_interval"] )
	if err != nil {
		return nil, err
	}

	timeout, err := strconv.Atoi( args["timeout"] )
	if err != nil {
		return nil, err
	}

	channelCapacity, err := strconv.Atoi( args["channel_capacity"] )
	if err != nil {
		return nil, err
	}

	bufsize, err := strconv.Atoi( args["buffer_size"] )
	if err != nil {
		return nil, err
	}

	advertise := false
	if args["advertise"] == "true" {
		advertise = true
	}

	connectionParams := bt.ConnectionParams{
		ConnectionTimeout: bt.Duration( connectionTimeout ),
		MinInterval: bt.Duration( minInterval ),
		MaxInterval: bt.Duration( maxInterval ),
		Timeout: bt.Duration( timeout ),
	}

	connections := []*bt.Device{}
	knownDevices := []string{}
	// collect channels' names
	for _, ch := range channels {
		knownDevices = append( knownDevices, ch.Name )
	}

	// check if we need to increase the capacity.
	if channelCapacity > 1 {
		received = make( chan *protocol.Message, channelCapacity )
		toSend = make( chan *protocol.Message, channelCapacity )
		publicKeys = make( chan *protocol.KnownPk, channelCapacity )
	}

	// create the connection itself.
	bconn := BluetoothConn{
		adapter,
		serverAdapter,
		&connectionParams,
		connections,
		&sync.RWMutex{},
		[]bt.DeviceCharacteristic{},
		knownDevices,
		map[string][]bt.DeviceCharacteristic{},
		sync.RWMutex{},

		nil,
		&sync.RWMutex{},
		[]*protocol.KnownPk{},
		&sync.RWMutex{},

		uint(bufsize),
		advertise,
	}

	util.DebugPrintln("Setting connect handler...")
	// """this must be called before adapter.Connect()"""
	// handling client connection
	adapter.SetConnectHandler( func(device bt.Device, connected bool) {
		util.DebugPrintln("adapter.SetConnectHandler() start")
		for idx, dev := range bconn.connections {
			util.DebugPrintln("adapter.SetConnectHandler(): device address:", device.Address)
			if dev.Address == device.Address {
				if connected {
					// found a device which is already connected
					util.DebugPrintln("adapter.SetConnectHandler(): found a device which is already connected")
					return
				} else {
					// device disconnected, drop from the list
					bconn.connections = append(
						bconn.connections[:idx],
						bconn.connections[idx+1:]...)
					util.DebugPrintln("adapter.SetConnectHandler(): device disconnected, drop from the list")
					return
				}
			}
		}
		util.DebugPrintln("adapter.SetConnectHandler() end")
	})

	// initialize channels on startup because we don't have any other
	// chance (InitChannels uses object, not a pointer on object).
	if advertise == true {
		// run as both client and server
		// advertise ourselves.
		adv := serverAdapter.DefaultAdvertisement()
		adv.Configure( bt.AdvertisementOptions{
			AdvertisementType: bt.AdvertisingTypeDirectInd,
			LocalName: args["local_name"],
		})
		if err = adv.Start(); err != nil {
			return nil, err
		}
		util.DebugPrintln("Advertisement started.")
		//defer adv.Stop()
	}
	// run only as a client
	// scan network for devices
	err = bconn.adapter.Scan( func( adapter *bt.Adapter, dev bt.ScanResult ) {
		util.DebugPrintln("found device:", dev.Address.String(), dev.RSSI, dev.LocalName())
		//util.DebugPrintln("bconn.adapter.Scan() start")
		if dev.LocalName() != "" {
			util.DebugPrintln("Found device:", dev)
			for _, kn := range knownDevices {
				if kn == dev.LocalName() {
					device, err := bconn.adapter.Connect(
						dev.Address,
						*bconn.connParams )
					if err == nil {
						connections = append( connections, &device )
						// found at least one known device
						adapter.StopScan()
					}
				}
			}
		}
		util.DebugPrintln("bconn.adapter.Scan() finish")
	})
	if err != nil {
		return nil, err
	}

	err = bconn.setupConnections()
	return bconn, err
}

func(b *BluetoothConn) setupConnections() error {
	var finalError error
	b.connectionsMtx.RLock()
	b.peersMtx.Lock()
	for _, dev := range b.connections {
		services, err := dev.DiscoverServices( nil )
		if err != nil {
			finalError = err
		} else {
			//uuids := []bt.UUID{}
			util.DebugPrintf("Found %d services at %s\n", len(services), dev.Address.String() )
			for _, service := range services {
				chars, err := service.DiscoverCharacteristics( nil )
				if err != nil {
					finalError = err
				} else {
					// handle characteristics
					util.DebugPrintf("\tFound %d characteristics.", len(chars))
					for _, ch := range chars {
						util.DebugPrintf("\tCharacteristic UUID: %s:%s\n",
						service.UUID(), ch.UUID())
					}
					val, ok := b.peers[ dev.Address.String() ]
					if !ok {
						b.peers[ dev.Address.String() ] = append( []bt.DeviceCharacteristic{}, chars... )
					} else {
						b.peers[ dev.Address.String() ] = append( val, chars... )
					}
				}
			}
		}
	}
	b.peersMtx.Unlock()
	b.connectionsMtx.RUnlock()

	// handle connected peers in the background
	go b.handleConnections()
	return finalError
}

// the same as tls.receiveMessagesInBackground
func(b *BluetoothConn) handleConnections() {
	// as bluetooth is faster than other platform's api
	// we use a separate thread for handling incoming messages.
	// it makes the network faster.
	for {
		b.peersMtx.RLock()
		for peer, chars := range b.peers {
			for _, char := range chars {
				// i hate this buffer size...
				buf := make( []byte, b.bufferSize ) //255 )
				n, err := char.Read( buf )
				if err == nil {
					msg := &protocol.Message{
						b.Name(),
						buf[:n],
						peer,
						false,
						map[string]string{},
					}
					received <- msg
				}
			}
		}
		b.peersMtx.RUnlock()
		// sleep a bit not to overload the pc.
		time.Sleep( time.Millisecond * 50 )
	}
}

// these are useful
func(b BluetoothConn) DeleteChannels() error {
	var finalError error

	b.connectionsMtx.RLock()
	for _, conn := range b.connections {
		if err := conn.Disconnect(); err != nil {
			finalError = err
		}
	}
	b.connectionsMtx.RUnlock()
	return finalError
}

func(b BluetoothConn) DistributePk( p *config.DistributionParameters, pk []byte ) error {
	// TODO
	var finalError error
	for _, peer := range b.peers {
		
		/*if peer.hasPk == false {
			nbytes, err := peer.conn.WriteWithoutResponse( pk ); err != nil {
				return err
			}
		}*/

		for _, ch := range peer {
			if _, err := ch.WriteWithoutResponse( pk ); err != nil {
				return err
			}
		}
	}

	b.pkMtx.Lock()
	b.pk = pk
	defer b.pkMtx.Unlock()
	return finalError
}


func(b BluetoothConn) CollectPks( p *config.DistributionParameters ) ([]protocol.KnownPk, error) {
	// TODO
	var finalError error
	pks := []protocol.KnownPk{}

	// the first message sent is always a public key content
	// collect public keys from channel
	for {
		if len( publicKeys ) == 0 {
			break
		}
		val := <- publicKeys
		pks = append( pks, *val )
	}
	// also save a copy of each public key
	for _, pubKey := range pks {
		publicKeys <- &pubKey
	}
	return pks, finalError
}

func(b BluetoothConn) Send( msg *protocol.Message ) error {
	// send message to every peer we have
	b.peersMtx.RLock()
	// if we are a server, we should send a message to everyone who is connected.
	// todo...

	// this is a client-part code
	for _, chars := range b.peers {
		for _, ch := range chars {
			if _, err := ch.WriteWithoutResponse( msg.Data ); err != nil {
				return err
			}
		}
	}
	b.peersMtx.RUnlock()
	return nil
}

func(b BluetoothConn) RecvAll() ( []*protocol.Message, error ) {
	// just move messages out of channel
	msgs := []*protocol.Message{}
	for {
		if len(received) == 0 {
			return msgs, nil
		}
		msgs = append( msgs, <- received )
	}
}

func(b BluetoothConn) GetSupportedExtensions() []string {
	return SuportedExt
}
