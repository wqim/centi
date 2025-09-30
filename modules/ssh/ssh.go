package centi_ssh
import (
	"os"
	"fmt"
	"net"
	//"sync"
	"strings"

	"golang.org/x/crypto/ssh"
	//"golang.org/x/crypto/ssh/terminal"
)

func startServer( addr, idRsa, authorizedKeys, validCredentials string ) ( net.Listener, *ssh.ServerConfig, error ) {
	// launch an ssh server on specified address
	authorizedKeysBytes, err := os.ReadFile( authorizedKeys )
	if err != nil {
		//util.DebugPrintln("(ssh::startServer) failed to read authorized keys file:", err)
		return nil, nil, err
	}

	authorizedKeysMap := map[string]bool{}
	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			//util.DebugPrintln("(ssh::startServer) failed to parse authorized keys file:", err)
			return nil, nil, err
		}

		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest
	}

	// parse all of the valid credentials in file
	validCreds := map[string]string{}
	parts := strings.Split( validCredentials, Delimeter )
	for _, part := range parts {
		// username:password
		creds := strings.Split( part, ":" )
		if len(creds) == 2 {
			validCreds[ creds[0] ] = creds[1]
		}
	}

	// setup configuration
	config := &ssh.ServerConfig{
		// Remove to disable password auth.
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in
			// a production setting.

			passwd, ok := validCreds[ c.User() ]
			if ok == true && passwd != "" {
				// such user is available
				// todo: add hashing and constant timee comparison
				if string(pass) == passwd {
					return nil, nil
				}
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},

		// Remove to disable public key auth.
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if authorizedKeysMap[string(pubKey.Marshal())] {
				return &ssh.Permissions{
					// Record the public key used for authentication.
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}

	// add a host key (nothing works without it)
	privateBytes, err := os.ReadFile( idRsa )
	if err != nil {
		//util.DebugPrintln("(ssh::startServer) failed to read private key bytes:", err)
		return nil, nil, err
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		//util.DebugPrintln("(ssh::startServer) failed to parse private key:", err)
		return nil, nil, err
	}
	config.AddHostKey(private)


	// actualy set up a server
	listener, err := net.Listen( "tcp", addr )
	if err != nil {
		//util.DebugPrintln("(ssh::startServer) failed to start server:", err)
		return nil, nil, err
	}
	return listener, config, nil
}
