package centi_ssh
import (
	"net"
	"strings"
	"golang.org/x/crypto/ssh"

	"centi/util"
	"centi/protocol"
)

func AddrToAlias( addr string ) string {
	parts := strings.Split( addr, ":" )
	if len(parts) > 1 {
		return parts[0]
	}
	return addr
}

/*
 * server-side functions for handling peers i/o
 */
func handleConnections(
	bufferSize uint,
	listener net.Listener,
	config *ssh.ServerConfig,
	maxConnections uint ) {


	for {
		//if listener != nil {
			// if we are able to accept connection
			connectedMtx.RLock()

			if connected < maxConnections {
				conn, err := listener.Accept()
				if err == nil {
					// handle connection in the separate thread
					connectedMtx.RUnlock()
					connectedMtx.Lock()
					connected++
					connectedMtx.Unlock()
					connectedMtx.RLock()

					util.DebugPrintln("[ssh:handleConnections] Got connection:", conn.RemoteAddr().String())
					//util.DebugPrintln("[ssh:handleConnections] mutex ok")
					conn, chans, reqs, err := ssh.NewServerConn( conn, config )
					if err == nil {
						go handleConnection(
							bufferSize,
							conn,
							chans,
							reqs,
						)
					} else {
						util.DebugPrintln("ssh.NewServerConn() failed:", err)
					}
				} else {
					util.DebugPrintln("listener.Accept() failed:", err)
				}
			}

			connectedMtx.RUnlock()
		//}
	}
}

func handleConnection(
	bufferSize uint,
	conn *ssh.ServerConn,
	chans <- chan ssh.NewChannel,
	reqs <- chan *ssh.Request ) {

	util.DebugPrintln("[ssh::handleConnection] Handling an incoming connection...")
	// cleanup on exit
	defer func() {
		conn.Close()
		connectedMtx.Lock()
		connected--
		connectedMtx.Unlock()
	}()

	sentPk := false

	// hande connected peer.
	for newChannel := range chans {
		// channels have a type, depending on the application level protocol intended.
		util.DebugPrintln("[ssh::handleConnection] Handling a channel:", newChannel )
		if newChannel.ChannelType() != "customdata" {
			newChannel.Reject( ssh.UnknownChannelType, "unknown channel type" )
			continue
		}

		util.DebugPrintln("Accepted a new channel")
		channel, requests, err := newChannel.Accept()
		if err != nil {
			return
		}

		util.DebugPrintln("Answering requests.", len(requests))
		
		var lastMsg *protocol.Message
		for req := range requests {
			// check if we should send a public key
			if sentPk == false {
				publicKeyMtx.RLock()
				if publicKey != nil {
					sentPk = true
					channel.SendRequest( PkRequestType, false, publicKey )
				}
				publicKeyMtx.RUnlock()
			}

			// pick the message to send (if any)
			currentMtx.Lock()
			sentToMtx.Lock()
			connectedMtx.RLock()
			if currentMessage == nil || sentTo == 0 || uint(sentTo) == connected {
				if len( toSend ) > 0 {
					currentMessage = <- toSend
					sentTo = 0
				}
			}

			if lastMsg != currentMessage {
				lastMsg = currentMessage
				channel.SendRequest( MsgRequestType, false, lastMsg.Data )
				sentTo++
			}
			connectedMtx.RUnlock()
			sentToMtx.Unlock()
			currentMtx.Unlock()


			//req.Reply( req.Type == "shell", nil )
			if req.Type == MsgRequestType {
				util.DebugPrintln("(ssh::handleConnection): got message from client:", len(req.Payload))
				msg := &protocol.Message{
					"",
					Name,
					req.Payload,
					AddrToAlias( conn.RemoteAddr().String() ),
					false,
					map[string]string{},
				}
				received <- msg
			} else if req.Type == PkRequestType {
				pk := &protocol.KnownPk{
					Name,
					AddrToAlias( conn.RemoteAddr().String() ),
					req.Payload,
				}
				publicKeys <- pk
				util.DebugPrintln("[ssh::handleConnction] Got public key:", len(req.Payload))
			}
		}

		util.DebugPrintln("Closing a channel.")
		channel.Close()
	}
}

/*
 * client-side function for handling incoming messages in the background
 */
 func handleIncomingMessages( conn ssh.Conn, channel ssh.Channel, requests <- chan *ssh.Request ) {
	 defer channel.Close()

	 util.DebugPrintln( "[", conn.RemoteAddr().String(), "] Amount of requests:", len(requests))
	 for req := range requests {
		 util.DebugPrintln( "[!!!!] Type of request:", req.Type )

		 if req.Type == MsgRequestType {
			 util.DebugPrintln("[ssh::handleIncomingMessages]: got message from server:", len(req.Payload))
			 msg := &protocol.Message{
				 "",
				 Name,
				 req.Payload,
				 AddrToAlias( conn.RemoteAddr().String() ),
				 false,
				 map[string]string{},
			 }
			 received <- msg
		 } else if req.Type == PkRequestType {
			 pk := &protocol.KnownPk{
				 Name,
				 AddrToAlias( conn.RemoteAddr().String() ),
				 req.Payload,
			 }
			 publicKeys <- pk
			 util.DebugPrintln("[ssh::handleIncomingMessages] Got public key:", len(req.Payload))
		 }
	}
	util.DebugPrintln("Exitting handleIncomingMessages...")
 }
