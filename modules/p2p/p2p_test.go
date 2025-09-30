package p2p
import (
	"fmt"
	"time"
	"bytes"
	"testing"
)

func TestP2P(t *testing.T) {
	args1 := map[string]string{
		"addrs": "/ip4/127.0.0.1/tcp/9000",
		"lowwater": "100",
		"highwater": "400",
		"grace_period": "600",	// 1 minute
		"protocol_id": "/chat/1.1.0",
		"rendezvous_strings": "hello",
	}
	args2 := map[string]string{
		"addrs": "/ip4/127.0.0.1/tcp/7000",
		"lowwater": "100",
		"highwater": "400",
		"grace_period": "600",	// 1 minute
		"protocol_id": "/chat/1.1.0",
		"rendezvous_strings": "hello",
	}
	
	p1, err := NewP2PConn( args1, nil )
	checkError( "Failed to create p1:", err, t )
	p2, err := NewP2PConn( args2, nil )
	checkError( "Failed to create p2:", err, t )

	// check if we are able to initialize channels
	if err = p2.InitChannels(); err != nil {
		t.Error("Failed to initialize channels (2): ", err)
	}
	if err = p1.InitChannels(); err != nil {
		t.Error("Failed to initialize channels (1): ", err)
	}

	// distribute public keys
	pk1 := []byte("This is a test message for connectivity testing.")
	pk2 := []byte("I hate everything I do and I feel like I'm useless")

	p1.DistributePk( nil, pk1 )
	p2.DistributePk( nil, pk2 )

	time.Sleep( time.Second * 10 )

	pk1r, _ := p2.CollectPks( nil )
	pk2r, _ := p1.CollectPks( nil )

	if bytes.Equal( pk1r[0].Content, pk1 ) == false {
		t.Errorf("Failed to retransmit public key: %v != %v", string(pk1), string(pk1r[0].Content))
	}

	if bytes.Equal( pk2r[0].Content, pk2 ) == false {
		t.Errorf("Failed to retransmit public key: %v != %v", string(pk2), string(pk2r[0].Content))
	}

	msg1, _ := p1.MessageFromBytes( []byte("Hello from peer 1!!!") )
	msg2, _ := p2.MessageFromBytes( []byte("Hi there from Peer 2?") )

	p1.Send( msg1 )
	m1, _ := p2.RecvAll()
	for _, m := range m1 {
		fmt.Println( string(m.Data) )
	}
	p2.Send( msg2 )
	m2, _ := p1.RecvAll()
	for _, m := range m2 {
		fmt.Println( string(m.Data) )
	}
}

func checkError( opt string, err error, t *testing.T ) {
	if err != nil {
		t.Error( opt, err )
	}
}
