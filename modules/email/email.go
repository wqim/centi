package email
import (
	"io"
	"fmt"
	"bytes"
	"strings"
	"net/smtp"
	"io/ioutil"
	"encoding/json"
	"encoding/base64"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	//"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"

	"centi/util"
	"centi/config"
	"centi/protocol"
	"centi/cryptography"
)

var (
	// add other files, which are commonly sent by email
	SupportedExt = []string{"jpg", "jpeg", "png", "pdf", "xls", "doc", "docx"}
)

type EmailConfig struct {
	fromAddr	string		`json:"email"`
	password	string		`json:"password"`
	subject		string		`json:"subject"`
	to		string		`json:"recipients"`
	smtpServer	string		`json:"smtp_addr"`	// host:port
	imapServer	string		`json:"imap_addr"`
}

type EmailConn struct {
	fromAddr	string
	password	string
	subject		string
	to		[]string
	smtpServer	string
	imapServer	string
	cli		*client.Client	// imap mail client
}

func NewEmailConn( args map[string]string, channels []config.Channel ) (protocol.Connection, error) {
	data, err := json.Marshal( args )
	if err != nil {
		return nil, err
	}
	var emailConf EmailConfig
	if err := json.Unmarshal( data, &emailConf ); err != nil {
		return nil, err
	}

	cli, err := client.DialTLS( emailConf.imapServer, nil )
	if err != nil {
		return nil, err
	}
	if err = cli.Login( emailConf.fromAddr, emailConf.password ); err != nil {
		return nil, err
	}

	var EmailConn = EmailConn{
		emailConf.fromAddr,
		emailConf.password,
		emailConf.subject,
		strings.Split( emailConf.to, "," ),
		emailConf.smtpServer,
		emailConf.imapServer,
		cli,
	}

	return EmailConn, nil
}

// probably does not need any realisation
func(e EmailConn) InitChannels() error {
	return nil
}

func(e EmailConn) DeleteChannels() error {
	return nil
}

func(e EmailConn) CreateChannel( c *config.Channel ) error {
	return nil
}

func(e EmailConn) DeleteChannel( c *config.Channel ) error {
	return nil
}

// do we really need this?
func(e EmailConn) PrepareToDelete( data []byte ) (*protocol.Message, error) {
	return nil, nil
}

func(e EmailConn) Delete( msg *protocol.Message ) error {
	return nil
}

// actually need realisation
func(e EmailConn) DistributePk( p *config.DistributionParameters, pk []byte ) error {
	tmpMsg := &protocol.Message{
		"",
		e.Name(),
		pk,
		protocol.UnknownSender,
		false,
		nil,
	}
	// must send it as a casual email
	return e.Send( tmpMsg )
}

func(e EmailConn) CollectPks( p *config.DistributionParameters ) ([]protocol.KnownPk, error) {
	msgs, err := e.RecvAll()
	if err != nil {
		return nil, err
	}
	// todo: parse messages into public keys
	keys := []protocol.KnownPk{}
	for _, m := range msgs {
		if len(m.Data) > cryptography.PkSize {
			keys = append( keys, protocol.KnownPk{
				e.Name(),
				m.Args["sender"],
				m.Data,
			})
		}
	}
	return keys, nil
}

func(e EmailConn) Send( msg *protocol.Message ) error {
	// TODO: add steganography
	parts := strings.Split( e.smtpServer, ":" )
	if len(parts) != 2 {
		return fmt.Errorf("Invalid server name format:", e.smtpServer)
	}
	var finalError error
	auth := smtp.PlainAuth("", e.fromAddr, e.password, parts[0] )
	for _, to := range e.to {
		msg :=  "From: " + e.fromAddr + "\r\n" + 
			"To: " + to + "\r\n" +
			"Subject: " + e.subject + "\r\n\r\n" +
			base64.StdEncoding.EncodeToString( msg.Data )

		err := smtp.SendMail( e.smtpServer, auth, e.fromAddr, []string{to}, []byte(msg) )
		if err != nil {
			finalError = err
		}
	}
	return finalError
}


func(e EmailConn) RecvAll() ([]*protocol.Message, error) {
	// TODO
	mailbox, err := e.cli.Select("INBOX", false)
	if err != nil {
		return nil, err
	}

	seqset := new( imap.SeqSet )
	seqset.AddRange( 1, mailbox.Messages )

	messages := make( chan *imap.Message, mailbox.Messages )

	// fetch messages
	section := imap.BodySectionName{}
	err = e.cli.Fetch( seqset, []imap.FetchItem{
		imap.FetchEnvelope, section.FetchItem(),
	}, messages)
	
	if err != nil {
		return nil, err
	}

	// parse them into protocol.Message structure
	incoming := []*protocol.Message{}
	for {
		if len(messages) == 0 {
			break
		}
		msg := <- messages
		if msg == nil {
			continue
		}

		// get message body
		r := msg.GetBody( &section )
		if r == nil {
			continue
		}
		// parse message
		mr, err := mail.CreateReader(r)
		if err == nil {
			// get the sender
			header := mr.Header
			from, err := header.AddressList("From")
			if err == nil {
				util.DebugPrintln("From:", from)
				// get the subject
				subject, err := header.Subject()
				if err == nil && subject == e.subject {
					// the message is related to centi network
					// iterate over message parts
					msgData := []byte{}
					for {
						p, err := mr.NextPart()
						if err == io.EOF {
							break
						} else if err == nil {
							switch h := p.Header.(type) {
							case *mail.InlineHeader:
								buf := new( bytes.Buffer )
								buf.ReadFrom( p.Body )
								bodyContent := buf.Bytes()
								msgData = append( msgData, bodyContent... )

							case *mail.AttachmentHeader:
								filename, _ := h.Filename()
								util.DebugPrintln("[email] Attachment:", filename)
								tmpbuf, err := ioutil.ReadAll( p.Body )
								if err == nil {
									// TODO: extract data from steganoed file
									msgData = append( msgData, tmpbuf... )
								}
							}
						}
					}

					sender := from[0].Address
					tmpMsg := protocol.Message{
						"",
						e.Name(),
						msgData,
						sender,
						false,
						map[string]string{},
					}
					incoming = append( incoming, &tmpMsg )
				}
			}
		}
	}
	return incoming, nil
}

func(e EmailConn) MessageFromBytes( data []byte ) (*protocol.Message, error) {
	msg := &protocol.Message{
		"",
		e.Name(),
		data,
		protocol.UnknownSender,
		false,
		map[string]string{},
	}
	return msg, nil
}

func(e EmailConn) Name() string {
	return "email"
}

func(e EmailConn) GetSupportedExtensions() []string {
	return SupportedExt
}
