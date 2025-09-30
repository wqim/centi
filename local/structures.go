package local
import (
)

type Result struct {
	Errors		[]string	`json:"errors"`
	MsgID		string		`json:"message_id"`
	Data		string		`json:"data"`
	Timestamp	int64		`json:"timestamp"`
}

type Request struct {
	Dst		string		`json:"dst"`		// destination - alias of receiver's public key
	Data		string		`json:"data"`		// base64-encoded data to send
}

type PollRequest struct {
	MsgID		string		`json:"message_id"`
	Timestamp	int64		`json:"timestamp"`	// i really don't know there to store it,,,
}

type Response struct {
	Ok		bool		`json:"ok"`		// if no error occured
	Message		string		`json:"message"`	// error message, if any
	Data		string		`json:"data"`		// base64-encoded response from receiver
}

type ConnectRequest struct {
	KeyAlias	string		`json:"key_alias"`	// alias of key to connect
}

type PkRequest struct {
	Peer		string		`json:"peer_alias"`	// alias of peer to request known public keys from
}
