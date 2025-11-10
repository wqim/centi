package local
import (
	"fmt"
	"time"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"encoding/base64"

	"centi/util"
	"centi/network"
	"centi/protocol"
	"centi/cryptography"
)

func writeJsonResponse( w http.ResponseWriter, resp []byte ) {
	w.Header().Set("Content-Type", "application/json")
	w.Write( resp )
}

func sendKeys(w http.ResponseWriter, r *http.Request, conn *protocol.ConnManagement ) {
	//util.DebugPrintln( util.CyanColor + "[GET /api/public-keys]" + util.ResetColor )
	//resp, _ := json.Marshal( *knownPks )
	packedKeys := []map[string]string{}

	for _, k := range conn.GetPeers() {
		tmp := map[string]string{
			"platform": k.GetPlatform(),
			"alias": k.GetAlias(),
			"content": base64.StdEncoding.EncodeToString( k.GetPublicKey() ),
		}
		packedKeys = append( packedKeys, tmp )
	}
	resp, err := json.Marshal( packedKeys )
	if err != nil {
		http.Error( w, "Internal Server Error", http.StatusInternalServerError )
	} else {
		//util.DebugPrintln( "[+] Sending response with public keys:", len( packedKeys ) )
		writeJsonResponse( w, resp )
	}
}

func requestKeys(w http.ResponseWriter, r *http.Request, logger *util.Logger,
		conn *protocol.ConnManagement, q *network.Queue ) {

	if r == nil || r.Body == nil {
		return
	}

	// parse the request
	var errors = []string{}
	var pkreq PkRequest
	defer r.Body.Close()

	data, err := ioutil.ReadAll( r.Body )
	if err != nil {
		errors = append( errors, err.Error() )
	} else {
		util.DebugPrintln("Data from client:", string(data) )
		if err = json.Unmarshal( data, &pkreq ); err != nil {
			errors = append( errors, err.Error() )
		} else {
			found := false
			// we should send public keys request to a specified peer if PkRequest.Peer == "*"
			for _, peer := range conn.GetPeers() {
				if pkreq.Peer == "*" || pkreq.Peer == peer.GetAlias() {
					// todo...
					found = true
					go q.RequestPublicKeys( peer )
				}
			}

			if found == false {
				util.DebugPrintln( "Tried to ask peer ", pkreq.Peer )
				errors = append( errors, "Peer " + pkreq.Peer + " not found, try to connect manually." )
			}
		}
	}
	response, err := json.Marshal( map[string][]string{
		"errors": errors,
	})
	if err != nil {
		http.Error( w, "Internal Server Error", http.StatusInternalServerError )
	} else {
		writeJsonResponse( w, response )
	}
}


func sendPeers( w http.ResponseWriter, r *http.Request, conn *protocol.ConnManagement ) {
	peersNames := []string{}
	peers := conn.GetPeers()
	for _, p := range peers {
		if p.GetKey() != nil && len(p.GetKey()) != 0 {
			peersNames = append( peersNames, p.Alias )
		}
	}
	packed, err := json.Marshal( peersNames )
	if err != nil {
		http.Error( w, "Internal Server Error", http.StatusInternalServerError )
	} else {
		writeJsonResponse( w, packed )
	}
}

func connectToPeer( w http.ResponseWriter, r *http.Request,
			logger *util.Logger, conn *protocol.ConnManagement,
			queue *network.Queue ) {
	
	if r == nil || r.Body == nil {
		return
	}
	// start key exchange with specified peer
	//util.DebugPrintln( util.CyanColor + "[POST /api/connect]" + util.ResetColor )
	errors := []string{}
	defer r.Body.Close()
	data, err := ioutil.ReadAll( r.Body )
	if err != nil {
		errors = append( errors, "Failed to read request body: " + err.Error())
	}
	msgID := ""
	timestamp := int64(0)
	if data != nil && len(data) > 0 {
		// have a valid request
		var connreq ConnectRequest
		if err := json.Unmarshal( data, &connreq ); err != nil {
			errors = append(errors, "Invalid request: JSON format required")
		} else {
			// find a corresponding key
			peer := conn.Peers.GetPeerByName( connreq.KeyAlias )
			if peer == nil {
				// there is no such key, dude
				errors = append( errors, "Key alias does not exist." )
			} else {
				// create a connection request packet
				util.DebugPrintln("Setting peer pk:",
					base64.StdEncoding.EncodeToString(peer.GetPublicKey()[:10]),
				)
				// encapsulating shared secret
				networkSubkey := []byte{}
				client := conn.CrClient
				//if conn.Config.EphemeralMode == true {
					
					// this fixes deanonymization issue
					networkSubkey, _ = cryptography.GenRandom( cryptography.SymKeySize )
					
					client, err = cryptography.NewClient()
					if err != nil {
						errors = append( errors, "Failed to generate new keys" )
					}
				/*} else {
					for k, _ := range queue.GetNetworkSubkeys() {
						if k == peer.Alias {
							networkSubkey = queue.NetworkSubkeys[ peer.Alias ]
							break
						}
					}
				}*/

				ss, encapsulated, err := peer.EncapsulateAndPack(
					client.ECCPrivateKey(),
					conn.Config.PacketSize,
					networkSubkey,
				)
				if err == nil {
					// add a peer with a new shared secret key, if we don't have a one yet.

					if peer.ValidSymKey() == false {
						util.DebugPrintln("shared secret is not set up yet, setting it...")
						peer.SetKey( ss )
						msgID = util.GenID()
						timestamp = time.Now().Unix()
						packet := &protocol.Message{
							"",
							"",
							encapsulated,
							protocol.UnknownSender,
							false,	// not a data packet, don't optimize
							map[string]string{},
						}
						queue.PushPacket( peer, packet )	// push the packet in the main thread(?)...
					}
				} else {
					errors = append( errors, "Failed to encapsulate and pack" )
				}
			}
		}
	}
	result := Result{ errors, msgID, "", timestamp }
	data, err = json.Marshal( result )
	if err != nil {
		http.Error( w, "Internal Server Error", http.StatusInternalServerError )
	} else {
		//util.DebugPrintln( "[+] Sending response (2):", string(data) )
		writeJsonResponse( w, data )
	}
}


func handleRequest( w http.ResponseWriter, r *http.Request,
			logger *util.Logger, conn *protocol.ConnManagement,
			queue *network.Queue ) {

	if r == nil || r.Body == nil {
		return
	}
	// handle user's request
	//util.DebugPrintln( util.CyanColor + "[POST /api/request]" + util.ResetColor )
	var request Request
	errors := []string{}
	defer r.Body.Close()

	data, err := ioutil.ReadAll( r.Body )
	if err != nil {
		errors = append( errors, "Failed to read request body: " + err.Error() )
	}

	msgID := ""
	timestamp := int64(0)
	if data != nil && len(data) > 0 {
		if err = json.Unmarshal( data, &request ); err != nil {
			errors = append( errors, "Invalid request. JSON format required." )
		} else {
			// encrypt the data(if possible) or start a key exchange
			realData, err := base64.StdEncoding.DecodeString( request.Data )
			if err != nil {
				errors = append(errors, "Invalid request: only base64 encoding is supported")
			} else if len(realData) > 0 {
				// check to whom we are sending a message
				//found := false
				p := conn.Peers.GetPeerByName( request.Dst )
				if p != nil {
					tmpkey := p.GetKey()
					if tmpkey != nil && len(tmpkey) > 0 {
						util.DebugPrintln("Real data (got from user):")
						util.DebugPrintln( string(realData) )
						queue.PushPacket( p, &protocol.Message{
							"",
							"",
							realData,
							protocol.UnknownSender,
							true,	// sure, it's send by us and we can optimize this packet.
							map[string]string{},
						})
					} else {
						errors = append( errors, "Peer is not connected." )
					}
				} else {
					// and if we don't have a target peer, return an error
					errors = append( errors, "Specified peer does not exist." )
				}
			} else {
				errors = append( errors, "No data to send." )
			}
		}
	}

	result := Result{
		errors,
		msgID,
		"",
		timestamp,
	}
	data, err = json.Marshal( result )
	if err != nil {
		// something really bad had happended just now...
		http.Error( w, "Internal Server Error", http.StatusInternalServerError )
		logger.LogError( fmt.Errorf("Failed to marshal response: " + err.Error()) )
	} else {
		writeJsonResponse( w, data )
	}

}

func handleGetResponse( w http.ResponseWriter, r *http.Request,
			logger *util.Logger, conn *protocol.ConnManagement,
			queue *network.Queue ) {
	// just send all the messages we received and do not fuck your brain
	messages := [][]string{}
	for {	// yes, the infinite loop, but it's not really infinite because
		// queue size is fixed
		newMsg := queue.Incoming()
		if newMsg == nil {
			break
		}
		// actually, it always contains a "sender" key, but better
		// safe than sorry.
		if newMsg.Sender != "" {
			messages = append( messages, []string{
				newMsg.Sender,
				base64.StdEncoding.EncodeToString(newMsg.Data),
			})
		}
	}
	packed, err := json.Marshal( messages )
	if err != nil {
		http.Error( w, "Internal Server Error", http.StatusInternalServerError )
	} else {
		writeJsonResponse( w, packed )
	}
}

func handleResponse( w http.ResponseWriter, r *http.Request,
			logger *util.Logger, conn *protocol.ConnManagement,
			queue *network.Queue ) {
	
	if r == nil || r.Body == nil {
		return
	}
	// how it should work:
	// 1. user polls the response url until timeout is reached or the
	//    answer is received.
	// 2. we are answering with remote response data
	//util.DebugPrintln( util.CyanColor + "[POST /api/response]" + util.ResetColor )
	var request PollRequest
	response := Result{}

	errors := []string{}
	defer r.Body.Close()

	data, err := ioutil.ReadAll( r.Body )
	if err != nil {
		errors = append( errors, "Failed to read request body: " + err.Error() )
	}
	if data != nil && len(data) > 0 {
		if err = json.Unmarshal( data, &request ); err != nil {
			errors = append( errors, "Invalid request. JSON format required." )
		} else {
			// check if we have a message in queue at all
			newMsg := queue.Incoming()
			if newMsg != nil {
				// yes, we have a new message!
				util.DebugPrintln("[+] Finally got a message:", string(newMsg.Data))
				response.Data = base64.StdEncoding.EncodeToString( newMsg.Data )
				//response.MsgID = newMsg.ID
			}
		}
	}

	response.Errors = errors
	data, err = json.Marshal( response )
	if err != nil {
		http.Error( w, "Internal Server Error", http.StatusInternalServerError )
		logger.LogError( fmt.Errorf("Failed to marshal response: " + err.Error()) )
	} else {
		if response.Data != "" {
			util.DebugPrintln("[+] Really got a message:", string(data))
		}
		//w.Write( data )
		writeJsonResponse( w, data )
	}	
}
