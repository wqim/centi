package p2p
import (
	//"io"
	"time"
	"bufio"
	"io/ioutil"
	//"context"
	//"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/network"
	
	"centi/protocol"
)

/*func(p P2PConn) Addr() peer.AddrInfo {
	return peer.AddrInfo{
		ID:	p.hst.ID(),
		Addrs:	p.hst.Addrs(),
	}
}

func(p P2PConn) Connect( addrinfo peer.AddrInfo ) error {
	return p.hst.Connect( context.Background(), addrinfo )
}
*/
func handleStream( peerName string, s network.Stream ) {
	rw := bufio.NewReadWriter( bufio.NewReader(s), bufio.NewWriter(s) )
	go readData( peerName, rw )
	go writeData( rw )
}

func readData( peerName string, rw *bufio.ReadWriter ) {
	firstTime := false
	for {
		data, err := ioutil.ReadAll(rw)
		if err != nil || data == nil || len(data) == 0 {
			time.Sleep( time.Second * 2 )
			continue
		}

		if firstTime == false {
			firstTime = true
			pkMtx.Lock()
			publicKeys = append( publicKeys, protocol.KnownPk{
				moduleName,
				peerName,
				data,
			})
			pkMtx.Unlock()
		} else {
			// handle received data...
			msg := &protocol.Message{
				moduleName,
				data,
				protocol.UnknownSender,
				false,
				map[string]string{},
			}
			received <- msg
		}
	}
}

func writeData( rw *bufio.ReadWriter ) {
	
	firstTime := false
	for {
		if firstTime == false && len(pubKey) > 0 {
			firstTime = true
			// sending public key
			rw.Write( pubKey )
			rw.Flush()
		} else if len(pubKey) > 0 {
			if len(toSend) > 0 {
				msg, ok := <- toSend
				if ok {
					rw.Write( msg.Data )
					rw.Flush()
				}
			}
		}
		// sleep for some time (can be anything)...
		time.Sleep( time.Second * 1 )
	}
}
