package universal
import (
	"bytes"
	"testing"
	"centi/config"
)

func TestUniversalMicroservice(t *testing.T) {
	args := map[string]string{
		"name": "test",
		"addr": "http://127.0.0.1:8001/",
	}
	channels := []config.Channel{}
	uniconn, _ := NewUniConn( args, channels )
	if err := uniconn.InitChannels(); err != nil {
		t.Error("Failed to initialize channels:", err)
	}

	params := config.DistributionParameters{}
	pk := []byte("Hello World!")
	data := []byte("Right Left Wrong")

	// distribute public key and receive it (as we are the only user of this microservice)
	if err := uniconn.DistributePk( &params, pk ); err != nil {
		t.Error("Failed to distribute public key:", err)
	}

	pks, err := uniconn.CollectPks( &params )
	if err != nil || len(pks) == 0 {
		t.Error("Failed to collect public keys:", err)
	}
	if !bytes.Equal( pks[0].Content, pk ) {
		t.Error("Failed to correctly get public key:", string(pks[0].Content), "!=", string(pk))
	}
	
	msg, _ := uniconn.MessageFromBytes( data )
	if err := uniconn.Send( msg ); err != nil {
		t.Error("Failed to send a message:", err)
	}

	msgs, err := uniconn.RecvAll()
	if err != nil || len(msgs) == 0 {
		t.Error("Failed to receive messages:", err)
	}
	if !bytes.Equal( msgs[0].Data, data ) {
		t.Error("Failed to correctly get messages:", string(msgs[0].Data), "!=", string(data))
	}


	if err := uniconn.DeleteChannels(); err != nil {
		t.Error("Failed to delete channels:", err)
	}
}
