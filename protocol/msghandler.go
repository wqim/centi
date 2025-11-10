package protocol
import (
	//"os"
	"fmt"
	"sync"
	"centi/util"
)

const (
	UnknownSender = ""
	SendChannel = uint8(0)
	RecvChannel = uint8(1)
)

type Message struct {
	// i want to refactor this
	Name		string			`json:"name"`		// filename or empty string
	Platform	string			`json:"platform"`	// may be neccessary
	Data		[]byte			`json:"data"`
	// alias of the sender (peer alias)
	// or empty string for unknown, or magic string for us.
	Sender		string			`json:"sender"`
	SentByUs	bool			`json:"sent_by_us"`	// check if the message was sent by us
	// i don't really think we need this, but may be used by modules
	Args		map[string]string	`json:"args"`
}

// channel for messages in packets, sorting them by their sequence number in packets
type MsgChannel struct {
	peerAlias	string
	total		uint64
	messages	[][]byte
	compressed	uint8
	mtx		sync.RWMutex
}

func NewMsgChannel( alias string, totalParts uint64, compressed uint8 ) *MsgChannel {
	return &MsgChannel{
		alias,
		totalParts,
		make([][]byte, totalParts),
		compressed,
		sync.RWMutex{},
	}
}

func(m *MsgChannel) IsFull() bool {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	for _, msg := range m.messages{
		if msg == nil || len(msg) == 0 {
			return false
		}
	}
	return true
}

func(m *MsgChannel) Push( msg []byte, index uint64 ) {

	m.mtx.Lock()
	defer m.mtx.Unlock()

	// basically, the checks are the same, but better safe than sorry
	// (code analysis tool warning here)
	//util.DebugPrintf("Pushing part of the message at #%d\n", index)
	if index < m.total && index < uint64(len(m.messages)) {	
		if m.messages[index] == nil || len(m.messages[index]) == 0 {
			m.messages[index] = msg
		}
	}
}

func(m *MsgChannel) Data() ([]byte, error) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	result := []byte{}

	for _, msg := range m.messages {
		if msg == nil || len(msg) == 0 {
			return nil, fmt.Errorf("not a full message")
		}
		result = append( result, msg... )
	}

	if m.compressed != 0 {
		return Decompress( result )
	}
	return result, nil
}


// general handler for all the incoming messages
type MsgHandler struct {
	msgChannels	[]*MsgChannel
	mtx		sync.RWMutex
}

func NewMsgHandler() *MsgHandler {
	return &MsgHandler {
		[]*MsgChannel{},
		sync.RWMutex{},
	}
}

// adds a MsgChannel if none exists
func(m *MsgHandler) AddChannel( alias string, total uint64, compressed uint8 ) {
	
	m.mtx.Lock()
	defer m.mtx.Unlock()
	
	if m.exists( alias, false ) == false {	// not safe because we are already holding the mutex

		ch := NewMsgChannel( alias, total, compressed )
		m.msgChannels = append( m.msgChannels, ch )
	}
}

func(m *MsgHandler) AddPacket( alias string, seq, total uint64, compressed uint8, data []byte ) {
	
	//util.DebugPrintln("MsgHandler::AddPacket. Data:", len(data))
	//util.DebugPrintln( string(data) )

	if m.exists( alias, true ) == false {
		m.AddChannel( alias, total, compressed )
	}
	m.mtx.Lock()
	defer m.mtx.Unlock()
	for _, ch := range m.msgChannels {
		if ch.peerAlias == alias {
			// really, who's gonna change the total packets
			// number?
			if ch.total == total {
				ch.Push( data, seq )
			}
		}
	}
}


// checks if there is a channel with corresponding peer alias
func(m *MsgHandler) exists( alias string, safe bool ) bool {

	if safe {
		m.mtx.RLock()
		defer m.mtx.RUnlock()
	}
	for _, ch := range m.msgChannels {
		if ch.peerAlias == alias {
			return true
		}
	}
	return false
}

// returns data of corresponding msgChannel value.
// if no data collected or there is no correct msgChannel, returns nil.
func(m *MsgHandler) ByAlias( alias string ) []byte {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	for idx, ch := range m.msgChannels {
		if ch.peerAlias == alias {
			//util.DebugPrintln("Found a messages chain for ", alias)
			if ch.IsFull() {
				//util.DebugPrintln("The chain is full")
				data, err := ch.Data()
				if err != nil {
					util.DebugPrintln("Failed to collect data from the chain:", err)
					return nil
				}
				// drop channel after we got data from it
				//util.DebugPrintf("Dropping channel at %d with alias %s\n", idx, m.msgChannels[idx].peerAlias)
				m.msgChannels = append(
					m.msgChannels[:idx],
					m.msgChannels[idx+1:]...,
				)
				//util.DebugPrintln("Collected data normally...")
				return data
			}
			break
		}
	}
	//util.DebugPrintln("Peer with alias ", alias, " not found")
	// peer not found by alias
	return nil
}
