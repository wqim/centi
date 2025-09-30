package config

import (
	"os"
	//"strings"
	"encoding/json"

	"centi/cryptography"
	"centi/util"
)

// general structure of channel
type Channel struct {
	Name	string				`json:"name"`
	Args	map[string]string		`json:"args"`
}

// general information about connection
type ConnectionInfo struct {
	Platform	string			`json:"platform"`
	Args		map[string]string	`json:"args"`
	Channels	[]Channel		`json:"channels"`
}

// public key distribution parameters
type DistributionParameters struct {
	Type	uint8			`json:"type"` // type of public key distribution
	Args	map[string]string	`json:"args"` // optional arguments for platform(s)
}

// basic network configuration
type NetworkConfig struct {

	// delays between messages in the queue
	MinDelay	uint	`json:"min_delay"`
	MaxDelay	uint	`json:"max_delay"`
	CollectionDelay	uint	`json:"keys_collection_delay"` // delay between public keys collection

	// the size of messages queue
	QueueSize	uint	`json:"queue_size"`

	// the size of packet
	PacketSize	uint	`json:"packet_size"`

	// some security and privacy-related parameters
	AcceptUnknown	bool	`json:"accept_unknown"`              // to be or not to be a 'receiver' during key exchange
	SendKnownPeers	bool	`json:"send_known_peers"`            // if true, allows resending of known public keys

	// mode in which we distribute ephemerial key dufing key exchange instead of real one
	// pretty useful to hide us from destination
	EphemeralMode	bool	`json:"ephemeral_mode"`

	// network key, required to prevent mitm attacks.
	// this one is used while distributing public key and must not be chaned even in ephemeral mode.
	NetworkKey	string			`json:"network_key"`
	// subkey for every peer, to protect from identify theft attacks.
	NetworkSubkeys	map[string]string	`json:"network_subkeys"`
	DistrParams	DistributionParameters `json:"key_distribution_parameters"` // parameters of public key ditribution
}

/*
 * Server configuration - configuration of local API server.
 * Ususally contains a bunch of pages to serve. These pages can be
 * network applications like chats, emails, etc.
 * There are some pages which set up by default, but for developers'
 * convinience configuration is available.
 */
type ServerConfiguration struct {
	Address		string			`json:"address"`
	NotFoundPage	string			`json:"not_found_page"`
	Pages		map[string]string	`json:"pages"`
}

/*
 * Configuration of keys and other client-related information stored.
 */
type KeysConfig struct {
	Pk		string			`json:"public_key"`
	Sk		string			`json:"private_key"`
	Peers		map[string][]string	`json:"peers"`
}

/*
 * Configuration for steganography. Currently contains only folder with files
 * but can contain steganography methods restrictions in the future.
 */
type SteganoConfig struct {
	Folder		string			`json:"decoy_files_folder"`
}

/*
 * Full configuration of the network. Yes, it's heavy but it allows to do
 * a lot of things like building subnetworks based on specified parameters.
 */
type FullConfig struct {
	NetworkConfig	NetworkConfig		`json:"network_config"`
	ServerConfig	ServerConfiguration	`json:"local_server_config"`
	StegConfig	SteganoConfig		`json:"steganography_config"`
	Logger		util.LoggerInfo		`json:"logger_config"`
	PlatformsData	[]ConnectionInfo	`json:"platforms_data"`
	DbFile		string			`json:"db_file"`
	DbPassword	string			`json:"db_password"`
	DbRowsLimit	uint			`json:"db_rows_limit"`
	Keys		KeysConfig		`json:"keys"`
}

func LoadConfig(filename string, key []byte) (*FullConfig, error) {
	data, err := LoadEncrypted(filename, key)
	if err != nil {
		return nil, err
	}

	var conf FullConfig
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

/*
 * TODO: change format to xjson or yaml for more readability.
 */
func SaveConfig(filename string, key []byte, c *FullConfig) error {
	data, err := json.MarshalIndent(*c, "", "\t")
	if err != nil {
		return err
	}
	return SaveEncrypted(filename, key, data)
}

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
			result[ k ] = cryptography.DeriveKey( pass, saltBytes )
		}
	}
	return result
}
