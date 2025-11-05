package bluetooth
import (
	bt "tinygo.org/x/bluetooth"

	"centi/util"
	"centi/protocol"
)

func writeEvent( client bt.Connection, offset int, value []byte ) {
	util.DebugPrintln("(bt::writeEvent) connection", client)
}

/* some Connection functions */
func(b BluetoothConn) InitChannels() error {
	// todo...?
	return nil
}


func(b BluetoothConn) Name() string {
	return "bluetooth"
}

func(b BluetoothConn) MessageFromBytes( data []byte ) (*protocol.Message, error) {
	msg := &protocol.Message{
		"",
		b.Name(),
		data,
		protocol.UnknownSender,
		false,	// does not really matter here
		map[string]string{},
	}
	return msg, nil
}

// some functions which are not really useful for this module
// but must be created anyway
func(b BluetoothConn) PrepareToDelete( data []byte ) (*protocol.Message, error) {
	return nil, nil
}

func(b BluetoothConn) Delete( msg *protocol.Message ) error {
	return nil
}

func(b BluetoothConn) Delete( msg *protocol.Message ) error {
	return nil
}
