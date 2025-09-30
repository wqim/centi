package local
import (
	"fmt"
	"time"
	"centi/util"
	"centi/config"
	"centi/modules"
	"centi/protocol"
	"centi/network"
	"centi/cryptography"
)

/*
 * package local contains all the functions and structures which are required
 * in order to run centi network and being able to send data to/from it.
 */
func RunCentiNetwork( configFile string, password, saltBytes []byte ) error {

	// 1. read all the things we need
	key := cryptography.DeriveKey( password, saltBytes )
	fullConfig, err := config.LoadConfig( configFile, key )
	if err != nil {
		return err
	}

	// 2. create all the things we need
	logger := util.NewLogger( &fullConfig.Logger )
	modules.InitAllModules()
	conns := buildConnections( fullConfig.PlatformsData )
	cryptClient, err := cryptography.ClientFromKeys( fullConfig.Keys.Pk, fullConfig.Keys.Sk )
	if err != nil {
		// try to build a new cryptographic client and save it in the configuration
		cryptClient, err = cryptography.NewClient()
		if err != nil {
			return err
		}
		fullConfig.Keys.Pk = cryptClient.PkToString()
		fullConfig.Keys.Sk, err = cryptClient.SkToString()
		if err != nil {
			return err
		}

		if err = config.SaveConfig( configFile, key, fullConfig ); err != nil {
			logger.LogError( fmt.Errorf("Failed to save configuration: " + err.Error()) )
			return err
		}
	}
	connections := protocol.NewConnManagement(
		conns,
		&fullConfig.NetworkConfig,
		cryptClient,
	)

	subkeys := config.ExtractNetworkSubkeys( fullConfig.NetworkConfig.NetworkSubkeys )

	// 3. initialize channels
	connections.DeleteChannels()
	//return nil
	connections.InitChannels()

	// 4. run the queue in the separate threads.
	queue, err := network.NewQueue(
		fullConfig.DbFile,
		fullConfig.DbPassword,
		fullConfig.DbRowsLimit,
		fullConfig.NetworkConfig.QueueSize,
		logger,
		&connections,
		subkeys,
	)
	
	if err != nil {
		return err
	}

	pk, err := cryptClient.GetPublicKey( connections.Peers.NetworkSubkey() )
	if err != nil {
		return err
	}

	//util.DebugPrintln("Our public key is ", base64.StdEncoding.EncodeToString(pk[:10]))
	err = connections.DistributePkEverywhere( pk )
	if err != nil {
		//util.DebugPrintln( util.RedColor + "[ERROR]:" + util.ResetColor, err )
		logger.LogError(err)
	}

	go collectPublicKeys(
		connections,
		logger,
		fullConfig.NetworkConfig.CollectionDelay,
	)

	go queue.RunNetworkBackground()
	go queue.RunNetwork()

	return RunCentiApiServer( &fullConfig.ServerConfig, logger, &connections, queue )
}

func buildConnections( conf []config.ConnectionInfo ) []protocol.Connection {
	conns := []protocol.Connection{}
	for _, ci := range conf {
		conn, err := modules.FromConnectionInfo( &ci )
		if err == nil {
			conns = append( conns, conn )
		}
	}
	return conns
}

func buildCryptClients( amount int, keys *config.KeysConfig ) []*cryptography.CryptClient {
	// initializing one client per module
	clients := []*cryptography.CryptClient{}
	for i := 0; i < amount; i++ {
		cli, err := cryptography.NewClient()
		if err == nil {
			clients = append( clients, cli )
		}
	}
	return clients
}

func collectPublicKeys(
	connections protocol.ConnManagement,
	logger *util.Logger,
	delay uint ) {
	for {
		util.DebugPrintln("[local::collectPublicKeys] collecting public keys...")
		pks, err := connections.CollectPks()
		if err == nil {
			//*publicKeysList.KnownPks = pks
			for _, pk := range pks {
				// fix public key alias/content:
				// 1. get public key by it's content
				// 
				// drop pqs key because it does not participates in
				// keys comparison
				if len( pk.Content ) < cryptography.PkSize {
					// not a public key at all
					util.DebugPrintln("Not a public key at all (< cryptography.PkSize)")
					continue
				}

				pkContent := pk.Content[ cryptography.PkSize :]

				_, peer := connections.Peers.GetPeerByPublicKey( pkContent )
				// also verify alias in order not to allow inpersonalization of existing peer.
				if peer != nil && peer.GetAlias() == pk.Alias {
					// already exists
					util.DebugPrintln("Got peer by public key:", peer.GetAlias(), "/", pk.Alias)
					peer.SetAlias( pk.Alias )
				} else {
					// peer with specified public key not found,
					// trying to find them by alias
					peer = connections.Peers.GetPeerByName( pk.Alias )
					if peer != nil {
						util.DebugPrintln("Got peer by name:", pk.Alias)
						// peer has updated their public key
						// set full version of public key
						peer.SetPk( pk.Content, connections.Peers.NetworkSubkey() )	
					} else {
						// found a new peer
						util.DebugPrintln("Did not found a peer ", pk.Alias, "; a new one?")
						newPeer := protocol.NewPeer( pk.Alias )
						if err := newPeer.SetPk( pk.Content, connections.Peers.NetworkSubkey() ); err == nil {
							connections.AddPeer( newPeer )
						}
					}
				}
			}
			//util.DebugPrintln("Dropping duplicate peers...")
			connections.Peers.DropDuplicates()
		} else {
			logger.LogError( err )
		}
		//if len(pks) > 0 {
			util.DebugPrintf("[local::collectPublicKeys] collected %d public keys.\n", len(pks))
		//}
		time.Sleep( time.Duration(delay) * time.Millisecond )
	}
}
