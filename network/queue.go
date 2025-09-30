package network
import (
	"centi/util"
	"centi/protocol"
)

// the queue which rules the whole network
type Queue struct {
	// refactor this too
	queue		chan *protocol.Message		// data in the jsoned packet
	incoming	chan *protocol.Message		// channel of incoming messages
	Logger		*util.Logger			// logger
	db		*util.DB			// database connection
	conn		*protocol.ConnManagement	// actual connection manager
	packetsBuf	*protocol.MsgBufManager		// handler for packets we are sending
	msgHandler	*protocol.MsgHandler		// handler for incoming data packets
	pkHandler	*protocol.MsgHandler		// handler for incoming public keys
	NetworkSubkeys	map[string][]byte		// subkeys for network
	closed		bool
}

// auxilary function for outer things.
func(q *Queue) PushPacket( receiver *protocol.Peer, packet *protocol.Message ) {
	if q.closed {
		return
	}
	
	if packet.SentByUs {
		util.DebugPrintln( util.GreenColor + "We are going to optimize this packet" + util.ResetColor )
		q.packetsBuf.Push(
			receiver,
			packet.Data,
			q.conn.Config.PacketSize,
		)
	} else {
		util.DebugPrintln( util.YellowColor + "We won't optimize this packet." + util.ResetColor )
		q.queue <- packet
	}
}

func(q *Queue) Incoming() *protocol.Message {
	if len(q.incoming) == 0 || q.closed {	// handle empty channel case
		return nil
	}
	res, ok := <- q.incoming
	if ok {
		return res
	}
	return nil
}

func NewQueue(  dbfilepath, dbpassword string, dbRowsLimit,
		queueSize uint, logger *util.Logger,
		conn *protocol.ConnManagement,
		networkSubkeys map[string][]byte ) (*Queue, error) {

	db, err := util.ConnectDB( dbfilepath, dbpassword, dbRowsLimit )
	if err != nil {
		util.DebugPrintln("Failed to ConnectDB")
		return nil, err
	}

	if err = db.InitDB(); err != nil {
		util.DebugPrintln("Failed to db.InitDB")
		return nil, err
	}

	return 	&Queue{
		make( chan *protocol.Message, queueSize ),
		make( chan *protocol.Message, queueSize ),
		logger,
		db,
		conn,
		protocol.NewMsgBufManager(),
		protocol.NewMsgHandler(),
		protocol.NewMsgHandler(),
		networkSubkeys,
		false,
	}, nil

}

func(q *Queue) Close() {
	q.closed = true
	q.db.Close()
	close(q.queue)
	close(q.incoming)
}
