package protocol
import (
	"fmt"
	"sync"
	"centi/util"
)

// messages buffer.
// stores as many messages as possible in the single []byte slice.
// useful for sending multiple messages inside one packet.
// optimizes the speed of the network
// ( do we really need this...? )
type MsgBuffer struct {
	peer		*Peer
	parts		[][]byte
	msgSize		uint
	packetSize	uint
	mtx		sync.Mutex
}

func NewMsgBuffer( peer *Peer, packetSize uint ) *MsgBuffer {
	msgSize := util.CalculateDataSize( make( []byte, packetSize ), packetSize )
	//uint( packetSize / 2 ) // as in peer struct
	return &MsgBuffer{
		peer,
		[][]byte{},
		msgSize,
		packetSize,
		sync.Mutex{},
	}
}

func(mb *MsgBuffer) Push( data []byte ) {
	mb.mtx.Lock()
	defer mb.mtx.Unlock()
	if len( mb.parts ) == 0 {
		// should we handle the size of data here...?
		mb.parts = append( mb.parts, data )
		return
	}

	lastElem := len(mb.parts) - 1
	if uint(len( mb.parts[lastElem] ) + len(data)) < mb.msgSize {
		mb.parts[ lastElem ] = append( mb.parts[ lastElem ], data... )
	} else {
		mb.parts = append( mb.parts, data )
	}
}

func(mb *MsgBuffer) Clear() {
	mb.mtx.Lock()
	defer mb.mtx.Unlock()

	mb.parts = [][]byte{}
}

func(mb *MsgBuffer) Next() ([][]byte, error) {

	mb.mtx.Lock()
	defer mb.mtx.Unlock()

	if len(mb.parts) == 0 {
		return nil, fmt.Errorf("Empty buffer")
	}
	part := mb.parts[0]
	mb.parts = mb.parts[1:]
	
	return mb.peer.PackToSend( part, mb.packetSize )
}

// manager of msgbuffer
type MsgBufManager struct {
	bufs	[]*MsgBuffer
	mtx	sync.Mutex
}

func NewMsgBufManager() *MsgBufManager {
	return &MsgBufManager{
		[]*MsgBuffer{},
		sync.Mutex{},
	}
}

func(mb *MsgBufManager) AddBuffer( msgbuf *MsgBuffer ) {
	mb.mtx.Lock()
	defer mb.mtx.Unlock()
	mb.bufs = append( mb.bufs, msgbuf )
}

func(mb *MsgBufManager) Push( peer *Peer, data []byte, packetSize uint ) {
	mb.mtx.Lock()
	defer mb.mtx.Unlock()
	for _, m := range mb.bufs {
		if m.peer.GetAlias() == peer.GetAlias() {
			m.Push( data )
			return
		}
	}
	// specified peer does not exist
	newBuffer := NewMsgBuffer( peer, packetSize )
	newBuffer.Push( data )
	mb.bufs = append( mb.bufs, newBuffer )
}

func(mb *MsgBufManager) Len() int {
	mb.mtx.Lock()
	defer mb.mtx.Unlock()
	return len(mb.bufs)
}

func(mb *MsgBufManager) Exists( peerAlias string ) bool {
	mb.mtx.Lock()
	defer mb.mtx.Unlock()
	for _, m := range mb.bufs {
		if m.peer.GetAlias() == peerAlias {
			return true
		}
	}
	return false
}

func(mb *MsgBufManager) Next() ([][]byte, error) {
	// how to pick who from peers must be the next???
	mb.mtx.Lock()
	defer mb.mtx.Unlock()

	if len( mb.bufs ) == 0 {
		return nil, fmt.Errorf("empty buffer list")
	}

	// is it a bug???
	data, err := mb.bufs[0].Next()
	mb.bufs = mb.bufs[1:]
	return data, err
}
