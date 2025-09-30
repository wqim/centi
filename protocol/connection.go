package protocol
import (
	"centi/config"
	"centi/cryptography"
)

type KnownPk struct {
	Platform	string
	Alias		string
	Content		[]byte
}

type Connection interface{

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
}

type ConnManagement struct {
	Connections	[]Connection
	Config		*config.NetworkConfig
	Peers		*PeerManager
	CrClient	*cryptography.CryptClient	// our cryptographical parameters
}

func NewConnManagement(
	connections []Connection,
	conf *config.NetworkConfig,
	crclient *cryptography.CryptClient ) ConnManagement {

	pass, saltBytes, err := cryptography.SplitWithSalt( conf.NetworkKey )
	if err != nil {
		return ConnManagement{}
	}

	networkKey := cryptography.DeriveKey( pass, saltBytes )
	return ConnManagement{
		connections,
		conf,
		NewPeerManager( networkKey ),
		crclient,
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
			allMsgs = append( allMsgs, msgs... )
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
			if tmperr = c.Send( msg ); tmperr != nil {
				err = tmperr
			}
		}
	}
	return err
}

func( cm *ConnManagement ) Delete( msg *Message ) error {
	var err error
	for _, c := range cm.Connections {
		msg2, tmperr := c.PrepareToDelete( msg.Data )
		if tmperr != nil {
			err = tmperr
		} else {
			if tmperr = c.Delete( msg2 ); tmperr != nil {
				err = tmperr
			}
		}
	}
	return err
}
