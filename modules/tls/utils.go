package centi_tls
import (
	"strings"

	//"centi/util"
	"centi/protocol"
)

/*
 * This file contains all the simple functions for
 * NetConn and some auxiliary functions.
 */

// isn't really used.
/*
func(n *NetConn) Close() {
	util.DebugPrintln("Close()")
	if n.peers != nil {
		for _, connection := range *n.peers {
			if connection != nil {
				connection.conn.Close()
			}
		}
	}
	if n.server != nil {
		n.server.Close()
	}
}
*/

func(n NetConn) MessageFromBytes(data []byte)  (*protocol.Message, error) {
	return &protocol.Message{
		"",
		Name,
		data,
		protocol.UnknownSender,
		false,
		map[string]string{},
	}, nil
}

func(n NetConn) Name() string {
	return Name
}


// are not really used because of TCP/UDP protocol nature.
func(n NetConn) Delete(msg *protocol.Message) error {
	return nil
}

// these could be useful if we had a pointer to NetConn
// but not a NetConn structure.
func(n NetConn) InitChannels() error {
	return nil
}
func(n NetConn) DeleteChannels() error {
	return nil
}

// utilities
func AddrToAlias( addr string ) string {
	parts := strings.Split( addr, ":" )
	if len(parts) > 1 {
		return Name + ":" + parts[0]
	}
	return Name + ":" + addr
}

func closePeer( peer *Peer, useRlock bool ) {
	// runlocking and rlocking mutex only because of
	// how this function is used in tcpudp.go file.
	// (such usage saves us from copy-pasting).
	if useRlock == true {
		peer.mtx.RUnlock()
	}

	peer.mtx.Lock()
	peer.conn.Close()
	peer.mtx.Unlock()

	if useRlock == true {
		peer.mtx.RLock()
	}
}
