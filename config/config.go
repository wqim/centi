package config

import (
	"os"
	//"strings"
	//"encoding/json"
	"gopkg.in/yaml.v3"
	
	"centi/cryptography"
	"centi/util"
)

// general structure of channel
type Channel struct {
	Name	string				`json:"name"`	// name of channel to use for communications
	Args	map[string]string		`json:"args"`	// optional arguments for this channel
}

// general information about connection
type ConnectionInfo struct {
	Platform	string			`json:"platform"`	// platform name
	Args		map[string]string	`json:"args"`		// optional arguments for Connection constructor
	Channels	[]Channel		`json:"channels"`	// a list of channels to use over this connection
}

// basic network configuration
type NetworkConfig struct {

	// delays between messages in the queue
	MinDelay	uint	`yaml:"min_delay"`
	MaxDelay	uint	`yaml:"max_delay"`
	CollectionDelay	uint	`yaml:"keys_collection_delay"` // delay between public keys collection

	// the size of messages queue
	QueueSize	uint	`yaml:"queue_size"`

	// the size of packet
	PacketSize	uint	`yaml:"packet_size"`

	// some security and privacy-related parameters
	AcceptUnknown	bool	`yaml:"accept_unknown"`              // to be or not to be a 'receiver' during key exchange
	SendKnownPeers	bool	`yaml:"send_known_peers"`            // if true, allows resending of known public keys

	// mode in which we distribute ephemerial key during key exchange instead of real one
	// pretty useful to hide us from destination
	//EphemeralMode	bool	`yaml:"ephemeral_mode"`	// bullshit, now will always be true

	// network key, required to prevent mitm attacks.
	// this one is used while distributing public key and must not be changed even in ephemeral mode.
	NetworkKey	string		`yaml:"network_key"`	// more security for closed communities.
	Peers		[]string	`yaml:"peers"`		// hashes of peer's public keys.
}

/*
 * Server configuration - configuration of local API server.
 * Usually contains a bunch of pages to serve. These pages can be
 * network applications like chats, emails, etc.
 * There are some pages which set up by default, but for developers'
 * convenience configuration is available.
 */
type ServerConfiguration struct {
	Address		string			`yaml:"address"`
	NotFoundPage	string			`yaml:"not_found_page"`
	Pages		map[string]string	`yaml:"pages"`
}

/*
 * Configuration of keys and other client-related information stored.
 */
type KeysConfig struct {
	Pk		string			`yaml:"public_key"`
	Sk		string			`yaml:"private_key"`
	//Peers		map[string][]string	`yaml:"peers"`
}

/*
 * Configuration for steganography. Currently contains only folder with files
 * but can contain steganography methods restrictions in the future.
 */
type SteganoConfig struct {
	Folder		string			`yaml:"decoy_files_folder"`
}

/*
 * Full configuration of the network. Yes, it's heavy but it allows to do
 * a lot of things like building subnetworks based on specified parameters.
 */
type FullConfig struct {
	NetworkConfig	NetworkConfig		`yaml:"network_config"`
	ServerConfig	ServerConfiguration	`yaml:"local_server_config"`
	StegConfig	SteganoConfig		`yaml:"steganography_config"`
	Logger		util.LoggerInfo		`yaml:"logger_config"`
	PlatformsData	[]ConnectionInfo	`yaml:"platforms_data"`
	DbFile		string			`yaml:"db_file"`
	DbPassword	string			`yaml:"db_password"`
	DbRowsLimit	uint			`yaml:"db_rows_limit"`
	Keys		KeysConfig		`yaml:"keys"`
}

/*
 * Functions for loading and saving configuration in YAML format.
 */
func LoadConfig(filename string, key []byte) (*FullConfig, error) {
	data, err := LoadEncrypted(filename, key)
	if err != nil {
		return nil, err
	}

	var conf FullConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func SaveConfig(filename string, key []byte, c *FullConfig) error {
	data, err := yaml.Marshal( *c ) //json.MarshalIndent(*c, "", "\t")
	if err != nil {
		return err
	}
	return SaveEncrypted(filename, key, data)
}

/*
 * Functions for saving and loading encrypted files.
 */
func LoadEncrypted(filename string, key []byte) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if key != nil && len(key) == cryptography.SymKeySize {
		return cryptography.Decrypt(data, key)
	}
	// return unencrypted data
	return data, nil
}

func SaveEncrypted(filename string, key, data []byte) error {

	var err error
	if key != nil && len(key) == cryptography.SymKeySize {
		data, err = cryptography.Encrypt(data, key)
		if err != nil {
			return err
		}
	}
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return err
	}
	return nil
}

// derive network subkey based on psk.
func ExtractNetworkSubkeys( networkSubkeys map[string]string ) map[string][]byte {
	result := map[string][]byte{}
	for k, v := range networkSubkeys {
		pass, saltBytes, err := cryptography.SplitWithSalt( v )
		if err == nil {
			//result[ k ] = cryptography.DeriveKey( pass, saltBytes )
			tmp, err := cryptography.DeriveKey( pass, saltBytes )
			if err == nil {
				result[ k ] = tmp
			}
		}
	}
	return result
}
