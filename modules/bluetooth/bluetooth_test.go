package bluetooth
import (
	"bytes"
	"testing"
	"centi/config"
)

func TestBluetooth(t *testing.T) {
	args := map[string]string{
		"connection_timeout": "5000",	// duration in 0.625µs units
		"min_interval": "100",
		"max_interval": "3000",
		"timeout": "7000",
		"channel_capacity": "100",
		"buffer_size": "4096",
		"advertise": "true",
		"local_name": "peer1",
	}

	chans := []config.Channel{
		config.Channel{
			"peer1",
			map[string]string{},
		},
	}
	b1, err := NewBluetoothConn( args, chans )
	if err != nil {
		t.Error("Failed to initialize bluetooth connection (1): ", err)
		return
	}

	//args["local_name"] = "peer2"
	args["advertise"] = "false"

	b2, err := NewBluetoothConn( args, chans )
	if err != nil {
		t.Error("Failed to initialize bluetooth connection (2): ", err)
		return
	}

	pk1 := []byte("The Pretty Reckless - Kill Me")
	pk2 := []byte("Evanescence - Going Under")

	if err = b1.DistributePk( nil, pk1 ); err != nil {
		t.Error("Failed to distribute first public key:", err)
		return
	}

	if err = b2.DistributePk( nil, pk2 ); err != nil {
		t.Error("Failed to distribute second public key:", err)
		return
	}
	

	b1pks, err := b2.CollectPks( nil )
	if err != nil {
	}

	b2pks, err := b1.CollectPks( nil )
	if err != nil {
	}

	if b1pks == nil || b2pks == nil || len(b1pks) == 0 || len(b2pks) == 0 {
		t.Error("Failed to collect public keys:", err)
		return
	}

	if bytes.Equal( b2pks[0].Content, pk1 ) == false {
		t.Error("Failed to receive first public key correctly:", string(pk1), "!=", string(b2pks[0].Content))
	}

	if bytes.Equal( b1pks[0].Content, pk2 ) == false {
		t.Error("Failed to receive second public key correctly:", string(pk2), "!=", string(b1pks[0].Content))
	}

	msg1, _ := b1.MessageFromBytes( []byte("from peer1: hello") )
	msg2, _ := b1.MessageFromBytes( []byte("from peer2: hi there!") )

	if err = b1.Send( msg1 ); err != nil {
		t.Error("Failed to send message1:", err)
	}
	if err = b2.Send( msg2 ); err != nil {
		t.Error("Failed to send message2:", err)
	}

	msgs1, err := b1.RecvAll()
	if err != nil {
		t.Error("b1 failed to receive messages:", err)
	} else if msgs1 == nil || len(msgs1) == 0 {
		t.Error("did not receive messages (1)")
	}

	msgs2, err := b2.RecvAll()
	if err != nil {
		t.Error("b2 failed to receive messages:", err)
	} else if msgs1 == nil || len(msgs1) == 0 {
		t.Error("did not receive messages (2)")
	}

	if bytes.Equal( msgs1[0].Data, msg2.Data ) == false {
		t.Error("Invalid message:", string(msg2.Data), "!=", string(msgs1[0].Data))
	}
	if bytes.Equal( msgs2[0].Data, msg1.Data ) == false {
		t.Error("Invalid message:", string(msg1.Data), "!=", string(msgs2[0].Data))
	}

}
