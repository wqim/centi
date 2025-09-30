package centi_ssh
import (
	"time"
	"bytes"
	"testing"
	"centi/util"
	"centi/config"
)

func TestSsh(t *testing.T) {
	// general configuration
	args := map[string]string {
		"valid_credentials": "test:i'm_g@y",
		"run_server": "true",
		"address": "127.0.0.1:2222",
		"channel_capacity": "100",
		"max_connections": "2",
		"buffer_size": "4096",
		"id_rsa": "/home/q/.ssh/id_rsa",
		"authorized_keys": "/home/q/.ssh/known_hosts",
		"known_hosts": "/home/q/.ssh/known_hosts",
	}

	// initialize node (server + client)
	node, err := NewSshConn( args, nil )
	if err != nil {
		t.Error("Failed to initialize ssh node:", err)
		return
	}

	// change some arguments for client
	channels := []config.Channel{
		config.Channel{
			Name: "127.0.0.1:2222",
			Args: map[string]string{
				"user": "test",
				"password": "i'm_g@y",
				"id_rsa": "/home/q/.ssh/id_rsa",
			},
		},
	}
	args["run_server"] = "false"

	// create client
	cli, err := NewSshConn( args, channels )
	if err != nil {
		t.Error("Failed to initialize ssh client:", err)
		return
	}

	// now we are able to exchange keys
	pk1 := []byte("Hello from peer1")
	pk2 := []byte("Hello from peer2")
	
	m1 := []byte("i'm gay")
	m2 := []byte("i'm lesbian")
	msg1, _ := node.MessageFromBytes( m1 )
	msg2, _ := node.MessageFromBytes( m2 )

	for i := 0; i < 1; i++ {

		util.DebugPrintln("Distributing public key (2)")
		if err = cli.DistributePk( nil, pk2 ); err != nil {
			t.Error("Client failed to distribute public key:", err)
		}
		
		util.DebugPrintln("Distributing public key (1)")
		if err = node.DistributePk( nil, pk1 ); err != nil {
			t.Error("Node failed to distribute public key:", err)
		}
		
		util.DebugPrintln("Small delay to await connection...")
		time.Sleep( time.Second )

		util.DebugPrintln("Collecting public keys")
		// collect public keys
		pks1, err := node.CollectPks( nil )
		if err != nil || len(pks1) == 0 {
			t.Error("Node failed to collect public keys:", err)
		} else if bytes.Equal( pks1[0].Content, pk2 ) == false {
			t.Error("Node received a wrong public key:", string(pks1[0].Content), "!=", (pk2))
		} else {
			util.DebugPrintln("[+] PK2 is ok")
		}

		pks2, err := cli.CollectPks( nil )
		if err != nil || len(pks2) == 0 {
			t.Error("Client failed to collect public keys:", err)
		} else if bytes.Equal( pks2[1].Content, pk1 ) == false {
			util.DebugPrintln("Amount of public keys received by client:", len(pks2))
			t.Error("Client received a wrong public key:", string(pks2[1].Content), "!=", string(pk1))
		} else {
			util.DebugPrintln("[+] PK1 is ok")
		}

		util.DebugPrintln("Sending messages")
		// send and receive messages


		if err = node.Send( msg1 ); err != nil {
			t.Error("Node failed to send a message:", err)
		}

		if err = cli.Send( msg2 ); err != nil {
			t.Error("Client failed to send a message:", err)
		}

		util.DebugPrintln("Receiving messages")
		time.Sleep( time.Second )

		msgs1, err := node.RecvAll()
		if err != nil || len(msgs1) == 0 {
			t.Error("Node failed to receive messages:", err)
		} else if bytes.Equal( msgs1[0].Data, m2 ) == false {
			t.Error("Node received wrong message:", msgs1[0], "!=", m2)
		} else {
			util.DebugPrintln("[+] MSG1 is ok")
		}

		//msgs2, err := cli.RecvAll()
		if err != nil || len(msgs1) == 0 {
			t.Error("Client faild to receive messages:", err)
		} else if bytes.Equal( msgs1[1].Data, m1 ) == false {
			t.Error("Client received wrong message:", msgs1[1], "!=", m1)
		} else {
			util.DebugPrintln("[+] MSG2 is ok")
		}
	}
}
