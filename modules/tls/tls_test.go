package centi_tls
import (
	"fmt"
	//"os"
	//"os/exec"
	//"time"
	"bytes"
	"testing"
	"centi/util"
	"centi/config"
	"centi/protocol"
)

func TestTCPUDP(t *testing.T) {

	/*if err := createCertificate("cert1.pem", "key1.pem"); err != nil {
		t.Error("Failed to create certificate and key:", err)
		return
	}
	if err := createCertificate("cert2.pem", "key2.pem"); err != nil {
		t.Error("Failed to create certificate and key (2):", err)
		return
	}*/


		fmt.Println("Testing first connection...")
		conf1 := map[string]string{
			"run_server": "true",
			"packet_size": "4096",
			"cert_path": "cert1.pem",
			"key_path": "key1.pem",
			"net_addr": "127.0.0.1:9000",
		}
		conf2 := map[string]string{
			"run_server": "false",
			"packet_size": "4096",
			"cert_path": "cert2.pem",
			"key_path": "key2.pem",
			"net_addr": "127.0.0.1:8000",
		}

		chans := []config.Channel{
			config.Channel{
				"127.0.0.1:9000",
				nil,
			},
		}

		// not an actual public keys, just dummies for check.
		pk1 := []byte("public key no 1")
		pk2 := []byte("public key no 2")

		util.DebugPrintln("Public key of server:", string(pk1))
		util.DebugPrintln("Public key of client:", string(pk2))
		
		conn1, err := NewNetConn( conf1, nil )
		if err != nil {
			t.Error("Failed to initialize first connection: ",
			"(" + "tcp" + ")", err)
		}

		conn2, err := NewNetConn( conf2, chans )
		if err != nil {
			t.Error("Failed to initialize second connection: ",
			"(" + "tcp" + ")", err)
		}

		fmt.Println("Initialized connections, starting public key distribution")
		// check public keys distribution
		// the order here does not really matter.
		/*if err = conn2.DistributePk( nil, pk2 ); err != nil {
			t.Error("Failed to send second pk:", err)
		}
		if err = conn1.DistributePk( nil, pk1 ); err != nil {
			t.Error("Failed to send first pk:", err)
		}


		time.Sleep(1 * time.Second)

		fmt.Println("Collecting public keys...")
		pks1, err := conn1.CollectPks( nil )
		if err != nil {
			t.Error("Failed to collect public keys(1):", err)
		}
		util.DebugPrintln("Public key received by server:", string(pks1[0].Content) )

		pks2, err := conn2.CollectPks( nil )
		if err != nil {
			t.Error("Failed to collect public keys(2):", err)
		}

		util.DebugPrintln("Public key received by client:", string(pks2[0].Content) )
		fmt.Println("Checking public keys content...")
		// server-side check
		if bytes.Equal( pks1[0].Content, pk2 ) == false {
			t.Error("Public key 2 incorrectly received:",
			string(pk2), "!=", string(pks1[0].Content))
		}

		// client-side check
		if bytes.Equal( pks2[1].Content, pk1 ) == false {
			t.Error("Public key 1 incorrectly received:",
			string(pk1), "!=", string(pks2[0].Content))
		}

		fmt.Println("Sending messages")
		// check messages sending....
		msg1 := []byte("Hello world!")
		msg2 := []byte("Goodbye world!")
	
		m1, _ := conn1.MessageFromBytes( msg1 )
		m2, _ := conn2.MessageFromBytes( msg2 )
		

		// client sends message, server receives it.
		if err = conn2.Send( m2 ); err != nil {
			t.Error("Failed to send m2:", err)
		}
		time.Sleep( 1 * time.Second )

		fmt.Println("Receiving messages from server...")
		messages1, err := conn1.RecvAll()
		if len(messages1) != 1 {
			t.Error("Failed to get messages from conn1:", err)
		}

		// server sends the message, client receives it.
		if err = conn1.Send( m1 ); err != nil {
			t.Error("Failed to send m1:", err)
		}
		time.Sleep( 1 * time.Second )

		fmt.Println("Receiving messages from client...")
		messages2, err := conn2.RecvAll()
		if len(messages2) != 1 {
			t.Error("Failed to receive messages from conn2:", err)
		}

		fmt.Println("Comparing messages")
		//fmt.Println("Messages1:", messages1)
		//fmt.Println("Messages2:", messages2)
		// check if messages content was damaged.
		if compareMessages( msg1, messages2[0] ) == false {
			t.Error("[1]:", string(msg1), "!=", string(messages2[0].Data))
		}
		if compareMessages( msg2, messages1[0] ) == false {
			t.Error("[2]:", string(msg2), "!=", string(messages1[0].Data))
		}
		*/
		fmt.Println("Closing connections...")
		// close connections
		conn1.DeleteChannels()
		conn2.DeleteChannels()
}

func compareMessages( data []byte, msg *protocol.Message ) bool {
	return bytes.Equal( data, msg.Data )
}

/*
openssl req -x509 -newkey rsa:4096 -keyout key1.pem -out cert1.pem -sha256 -days 365 -nodes -subj "/C=XX/ST=StateName/L=CityName/O=CompanyName/OU=CompanySectionName/CN=CommonNameOrHostname"

func createCertificate( certName, keyName string ) error {
	cmd := exec.Command(
		"openssl",
		"genrsa",
		"-out",
		keyName,
		"4096",
	)
	fmt.Println( cmd )

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command(
		"openssl",
		"req",
		"-x509",
		"-new",
		"-nodes",
		"-key",
		keyName,
		"-sha256",
		"-days", "65",
		"-out",
		certName,
	)
	fmt.Println( cmd )
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}*/
