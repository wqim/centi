package main
import (
	"os"
	"fmt"
	//"flag"
	"path/filepath"

	"centi/util"
	"centi/config"
	"centi/local"
	//"centi/protocol"
	"centi/cryptography"
)

const (
	CentiFolder = ".centi"
	ConfigFilename = "config.json"
	LogFilename = "log.log"
	DbFilename = "db.db"
)

func main() {

	if len( os.Args ) < 2 {
		help()
		return
	}

	password, err := util.GetPasswd("Password: ")
	if err != nil {
		fatal("Failed to read password from stdin:", err)
	}

	key := cryptography.DeriveKey( password )

	home, err := os.UserHomeDir()
	if err != nil {
		fatal("Failed to get home directory:", err)
	}
	// check if we have configuration
	centiFolder := filepath.Join( home, CentiFolder )
	configFile := filepath.Join( centiFolder, ConfigFilename )
	// if the application is installed for the first time, create all the
	// things we need.
	if _, err := os.Stat( configFile ); err != nil {
		if _, err = os.Stat( centiFolder ); err != nil {
			// create directory and corresponding files
			if err = os.Mkdir( centiFolder, 0770 ); err != nil {
				fatal( "Failed to create centi directory:", err )
			}
		}
		conf := defaultConfig( centiFolder )
		if err = config.SaveConfig( configFile, key, conf ); err != nil {
			fatal("Failed to save default configuration:", err)
		}
	}


	switch os.Args[1] {
	case "run":
		// run the network
		if err = local.RunCentiNetwork( configFile, password ); err != nil {
			fatal( "Failed to run network:", err )
		}
	case "editconf":
		// edit configuration in secure manner
		if err = util.EditConfig( configFile, password ); err != nil {
			fatal( "Failed to edit configuration:", err )
		}

	case "readlog":
		// read network logs
		logFile := filepath.Join( centiFolder, LogFilename )
		if err := util.ReadLog( logFile, password ); err != nil {
			fatal( "Failed to read log file:", err)
		}
	default:
		help()
	}
}

func defaultConfig( centiFolder string ) *config.FullConfig {
	//filename := filepath.Join( centiFolder, ConfigFilename )
	dbfilename := filepath.Join( centiFolder, DbFilename )
	loggerFilename := filepath.Join( centiFolder, LogFilename )

	cr, err := cryptography.NewClient()
	if err != nil {
		fatal("Failed to generate public and private keys:", err)
	}
	
	sk, err := cr.SkToString()
	if err != nil {
		fatal("Failed to conver private key to string:", err)
	}

	conf := config.FullConfig{
		NetworkConfig: config.NetworkConfig{
			MinDelay: 10000,
			MaxDelay: 20000,
			CollectionDelay: 5000,
			//Timeout: 80000,
			QueueSize: 10,
			PacketSize: 2048,
			AcceptUnknown: true,
			SendKnownPeers: true,
			DistrParams: config.DistributionParameters{
				Type: 0,
				Args: map[string]string{},
			},
		},
		ServerConfig: config.ServerConfiguration{
			Address: "127.0.0.1:8080",
			NotFoundPage: "www/404.html",
			Pages: map[string]string{
				"GET /{$}": "www/index.html",
				"GET /styles.css": "www/styles.css",
				"GET /script.js": "www/script.js",
			},
		},
		Logger: util.LoggerInfo{
			Filename: loggerFilename,
			Password: "",
			IsEncrypted: false,
			IsColored: true,
			SaveTime: true,
			Mode: util.Error,
		},
		PlatformsData: []config.ConnectionInfo{
		},
		DbFile: dbfilename,
		DbPassword: util.GenID(),
		DbRowsLimit: 10000,
		Keys: config.KeysConfig{
			Pk: cr.PkToString(),
			Sk: sk,
			Peers: map[string][]string{},
		},
	}
	return &conf
}

func fatal( args ...any ) {
	fmt.Println( args... )
	os.Exit(-1)
}

func help() {
	// todo: add a pretty help menu
	line := `Usage: ./centi <command> [arguments]

The following commands are supported:
	run		run the network
	editconf	edit configuration
	readlog		read log file
`

	fmt.Printf("%s", line)
}
