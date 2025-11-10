package universal
import (
	"fmt"
	"encoding/json"
	"encoding/base64"
	"strings"

	"centi/util"
	"centi/config"
	"centi/protocol"
	"centi/modules/general"
)

// universal module connector
// works as a proxy for microservice
var (
	// none of extensions are supported by default...?
	SupportedExt = []string{}
)

const (
	DefaultModuleName = "universal"
)

type APIMessage struct {
	MessageType	string		`json:"message_type"`
	Args		map[string]any	`json:"args"`
}

type APIResponse struct {
	MessageType	string		`json:"message_type"`
	Status		string		`json:"status"`	// "success" or "failure"
	Args		map[string]any	`json:"args"`
}

type UniConn struct {
	name		string			// name of microservice
	addr		string			// address of service
	headers		map[string]string	// reserved for future use..?
	channels	[]config.Channel
	supExt		[]string		// supported files extensions
}

func NewUniConn( args map[string]string, channels []config.Channel ) (protocol.Connection, error) {

	supExt := []string{}

	sext, ok := args["files_extensions"]
	if ok {
		supExt = strings.Split( sext, "," )
	}

	conn := UniConn{
		args["name"],
		args["addr"],
		map[string]string{
			"Content-Type": "application/json",
		},
		channels,
		supExt,
	}
	util.DebugPrintln("New universal connection:", conn.name, "at", conn.addr)
	margs := map[string]any{}
	for k, arg := range args {
		margs[k] = arg
	}

	m := APIMessage{
		"init_microservice",
		margs,
	}
	err := conn.shortenedSend( &m )
	return conn, err
}

func(u *UniConn) send( m *APIMessage ) (*APIResponse, error) {
	data, err := json.Marshal( *m )
	if err != nil {
		return nil, err
	}
	//util.DebugPrintln("send():", string(data))
	// make a post http request.
	resp, err := general.HTTPRequest( u.addr, "POST", data, u.headers )
	if err != nil {
		return nil, err
	}
	var response APIResponse
	if err = json.Unmarshal( resp, &response ); err != nil {
		return nil, err
	}
	return &response, nil
}

// shortened version of send + anyErrorOccured for functions requiring
// only to know if any error occured during request
func(u *UniConn) shortenedSend( m *APIMessage ) error {
	resp, err := u.send( m )
	if err != nil {
		return err
	}
	return anyErrorOccured( resp )
}

func(u UniConn) InitChannels() error {
	m := APIMessage{
		"init_channels",
		map[string]any{
			"channels": u.channels,
		},
	}
	return u.shortenedSend( &m )
}

func(u UniConn) DeleteChannels() error {
	m := APIMessage{
		"delete_channels",
		map[string]any{
			"channels": u.channels,
		},
	}
	return u.shortenedSend( &m )
}

/*
func(u UniConn) DistributePk( p *config.DistributionParameters, pk []byte ) error {
	m := APIMessage{
		"distribute_pk",
		map[string]any{
			"distribution_parameters": *p,
			"public_key": base64.StdEncoding.EncodeToString( pk ),
		},
	}
	return u.shortenedSend( &m )
}

func(u UniConn) CollectPks( p *config.DistributionParameters ) ([]protocol.KnownPk, error) {
	m := APIMessage{
		"collect_pks",
		map[string]any{
			"distribution_parameters": *p,
		},
	}
	resp, err := u.send( &m )
	if err != nil {
		return nil, err
	}
	util.DebugPrintln("Response:", resp)
	util.DebugPrintln("Response keys:",resp.Args["public_keys"])
	if strings.ToLower( resp.Status ) == "failure" {
		return nil, fmt.Errorf("%s", resp.Args["error"])
	}
	keys := []protocol.KnownPk{}

	tmp, ok := resp.Args["public_keys"].([]interface{})
	util.DebugPrintln("tmp =", tmp, "; ok =", ok)
	if !ok {
		return nil, fmt.Errorf("Invalid response format: failed to get public keys.")
	} else {
		for _, item := range tmp {
			mp, ok := item.(map[string]any)
			if !ok {
				util.DebugPrintln("Failed to convert item", item, "to map[string]any")
			} else {
				var key protocol.KnownPk
				skipKey := false
				for k, v := range mp {
					val, ok := v.(string)
					if !ok {
						util.DebugPrintln("Convert to string failed", k, "=", v)
						continue
					}

					switch k {
					case "alias":
						key.Alias = val
					case "platform":
						key.Platform = val
					case "content":
						key.Content, err = base64.StdEncoding.DecodeString( val )
						if err != nil {
							skipKey = true
						}
					}
				}

				if skipKey == false {
					keys = append( keys, key )
				}
			}
		}
	}
	return keys, nil
}
*/

func(u UniConn) Send( msg *protocol.Message ) error {
	m := APIMessage{
		"send",
		map[string]any{
			"message": *msg,
		},
	}
	return u.shortenedSend( &m )
}

func(u UniConn) RecvAll() ([]*protocol.Message, error) {
	m := APIMessage{
		"recv_messages",
		map[string]any{},
	}
	resp, err := u.send( &m )
	if err != nil {
		return nil, err
	}
	if strings.ToLower( resp.Status ) == "failure" {
		return nil, fmt.Errorf("%s", resp.Args["error"])
	}
	messages := []*protocol.Message{}
	// get a list of messages received by microservice
	tmp, ok := resp.Args["messages"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("[??????????] Invalid response format: failed to get incoming messages")
	}
	for _, item := range tmp {
		mp, ok := item.(map[string]any)
		if !ok {
			util.DebugPrintln("[!!!!!!!!!!] Failed to convert Item:", item)
		} else {
			var msg protocol.Message
			skipMessage := false
			for k, v := range mp {
				switch k {
				case "platform":
					msg.Platform, ok = v.(string)
					if !ok {
						continue
					}
				case "sender":
					msg.Sender, ok = v.(string)
					if !ok {
						continue
					}
				case "sent_by_us":
					msg.SentByUs, ok = v.(bool)
					if !ok {
						continue
					}
				case "args":
					msg.Args, ok = v.(map[string]string)
					if !ok {
						continue
					}
				case "data":
					val, ok := v.(string)
					if !ok {
						continue
					}
					msg.Data, err = base64.StdEncoding.DecodeString(val)
					if err != nil {
						skipMessage = true
					}
				}
			}

			if skipMessage == false {
				messages = append( messages, &msg )
			}
		}
	}

	return messages, nil
}

func(u UniConn) Delete( msg *protocol.Message ) error {
	m := APIMessage{
		"delete",
		map[string]any{
			"message": *msg,
		},
	}
	return u.shortenedSend( &m )
}

func(u UniConn) MessageFromBytes( data []byte ) (*protocol.Message, error) {
	m := APIMessage{
		"message_from_bytes",
		map[string]any{
			"data": base64.StdEncoding.EncodeToString( data ),
		},
	}

	resp, err := u.send( &m )
	if err != nil {
		return nil, err
	}
	if strings.ToLower( resp.Status ) == "failure" {
		return nil, fmt.Errorf("%s", resp.Args["error"])
	}

	args := map[string]string{}
	for k, v := range resp.Args {
		args[k] = fmt.Sprintf("%v", v)
	}

	msg := &protocol.Message{
		args["msg_name"],
		u.Name(),
		data,
		protocol.UnknownSender,
		true,	// yes, it's sent from us.
		args,
	}
	return msg, nil
}

func(u UniConn) Name() string {
	if u.name != "" {
		return u.name
	}
	return DefaultModuleName
}

func(u UniConn) GetSupportedExtensions() []string {
	return u.supExt
}

func anyErrorOccured( resp *APIResponse ) error {
	if strings.ToLower( resp.Status ) == "failure" {
		// yes, occured
		// get an error
		return fmt.Errorf( "%s", resp.Args["error"] )
	}
	// no error occured
	return nil
}
