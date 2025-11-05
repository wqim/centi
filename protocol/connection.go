package protocol
import (
	"strings"
	"centi/util"
	"centi/config"
	"centi/cryptography"
)

type KnownPk struct {
	Platform	string	`json:"platform"`
	Alias		string	`json:"alias"`
	Content		[]byte	`json:"content"`
}

type Connection interface{
	// general functions for real usage of module.
	Name() string
	InitChannels() error
	DeleteChannels() error
	DistributePk( p *config.DistributionParameters, pk []byte ) error
	CollectPks( p *config.DistributionParameters ) ([]KnownPk, error)
	Send( msg *Message ) error
	RecvAll() ([]*Message, error)
	PrepareToDelete( data []byte ) (*Message, error )
	Delete( msg *Message ) error
	MessageFromBytes( data []byte ) (*Message, error)

	// some functions for flexible steganography support
	GetSupportedExtensions() []string
}

type ConnManagement struct {
	Connections	[]Connection
	Config		*config.NetworkConfig
	StegConfig	*config.SteganoConfig
	Peers		*PeerManager
	CrClient	*cryptography.CryptClient	// our cryptographical parameters
	storage		*util.Storage
}

func NewConnManagement(
	connections []Connection,
	conf *config.NetworkConfig,
	sconf *config.SteganoConfig,
	crclient *cryptography.CryptClient ) ConnManagement {

	pass, saltBytes, err := cryptography.SplitWithSalt( conf.NetworkKey )
	if err != nil {
		return ConnManagement{}
	}

	networkKey := cryptography.DeriveKey( pass, saltBytes )
	return ConnManagement{
		connections,
		conf,
		sconf,
		NewPeerManager( networkKey ),
		crclient,
		util.NewStorage(),
	}
}

func( cm *ConnManagement ) GetPeers() []*Peer {
	return cm.Peers.GetPeers()
}

func( cm *ConnManagement ) AddPeer( peer *Peer ) {
	cm.Peers.AddPeer( peer )
}

// pk - packed public key of our own
func( cm *ConnManagement ) DistributePkEverywhere( pk []byte ) error {
	var err error
	for _, c := range cm.Connections {
		if tmperr := c.DistributePk( &cm.Config.DistrParams, pk ); tmperr != nil {
			err = tmperr
		}
	}
	return err
}

func( cm *ConnManagement ) CollectPks() ([]KnownPk, error) {
	var finalError error
	allKeys := []KnownPk{}
	for _, c := range cm.Connections{
		keys, err := c.CollectPks( &cm.Config.DistrParams )
		if err == nil {
			allKeys = append( allKeys, keys... )
		} else {
			finalError = err
		}
}
	return allKeys, finalError
}

func ( cm *ConnManagement ) InitChannels() error {
	var finalError error
	for _, c := range cm.Connections {
		if err := c.InitChannels(); err != nil {
			finalError = err
		}
	}
	return finalError
}

func ( cm *ConnManagement ) DeleteChannels() error {
	var finalError error
	for _, c := range cm.Connections {
		if err := c.DeleteChannels(); err != nil {
			finalError = err
		}
	}
	return finalError
}


func( cm *ConnManagement ) RecvAll() ([]*Message, error) {
	var err error
	allMsgs := []*Message{}
	for _, c := range cm.Connections {
		msgs, tmpErr := c.RecvAll()
		if msgs != nil && len(msgs) > 0 { //tmpErr == nil {
			// steganography decoding must be here...?
			exts := c.GetSupportedExtensions()
			if exts == nil || len(exts) == 0 {
				// steganography unsupported - collected encrypted messages
				allMsgs = append( allMsgs, msgs... )
			} else {
				for _, m := range msgs {
					recovered, err := RevealFromFile( m.Name, m.Data )
					if err == nil {
						allMsgs = append( allMsgs,  &Message{
								m.Name,
								m.Platform,
								recovered,
								m.Sender,
								m.SentByUs,
								m.Args,
							},
						)
					} else {
						// append the message anyway?
						allMsgs = append( allMsgs, m )
					}
				}
			}
		} else if tmpErr != nil {
			err = tmpErr
		}
	}
	return allMsgs, err
}


func( cm *ConnManagement ) SendToAll( data []byte ) error {
	var err error
	for _, c := range cm.Connections {
		msg, tmperr := c.MessageFromBytes( data )
		if tmperr != nil {
			err = tmperr
		} else {
			// steganography encoding must be here...?
			exts := c.GetSupportedExtensions()
			if exts != nil && len(exts) > 0 {
				util.DebugPrintln("[+] Module", c.Name(), "supports steganography.")
				// steganographic things are supported by module/microservice
				// hide data inside the file.
				fname, dt, err := HideInFile( cm.StegConfig.Folder, exts, msg.Data )
				if err != nil {
					// ignore for now, let other channels do the thing.
					continue
				}
				fname = util.PrepareFilename( fname )
				cm.storage.Add( c.Name() + ":" + fname, msg.Data )
				// setup a filename for message
				msg.Name = fname
				msg.Data = dt
			} else {
				util.DebugPrintln("[-] Module", c.Name(), "does not support steganography.")
			}
			if tmperr = c.Send( msg ); tmperr != nil {
				err = tmperr
			}
		}
	}
	return err
}

func( cm *ConnManagement ) Delete( msg *Message ) error {
	var err error
	fnames := cm.storage.Find( msg.Data )

	msg2 := msg
	var tmperr error
	for _, c := range cm.Connections {
		/*msg2, tmperr := c.PrepareToDelete( msg.Data )
		if tmperr != nil {
			err = tmperr
		} else {
			//if tmperr = c.Delete( msg2 ); tmperr != nil {
			//	err = tmperr
			//}
		*/	for _, fname := range fnames {
				parts := strings.Split(fname, ":")
				if len(parts) == 2 {
					msg2.Name = parts[1]
					if tmperr = c.Delete( msg2 ); tmperr != nil {
						err = tmperr
					}
				}
			}
		//}
	}
	cm.storage.Remove( msg.Data )
	return err
}
