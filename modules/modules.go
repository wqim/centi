package modules
import (
	"fmt"
	"centi/config"
	"centi/protocol"

	// import all the platform modules here...
	"centi/modules/email"
	"centi/modules/gitea"
	"centi/modules/github"
	"centi/modules/huggingface"
	//"centi/modules/bluetooth"
	//"centi/modules/p2p"
	"centi/modules/tls"
	ssh "centi/modules/ssh"
)

/* TODO:
 * build a platform client according to connection info presented in the structure.
 * returns a *Connection structure sample in case of success, (nil, err) in case
 * if some parameters are wrong or something strange is happening.
 */
type Module struct {
	Name		string
	SupportedExt	[]string	// list of supported files extensions
	Contructor	func( map[string]string, []config.Channel ) (protocol.Connection, error)
}

var (
	id = 0
	modules	= []Module{}
)

func FromConnectionInfo( ci *config.ConnectionInfo ) ( protocol.Connection, error ) {
	for _, m := range modules {
		if m.Name == ci.Platform {
			return m.Contructor( ci.Args, ci.Channels )
		}
	}
	return nil, fmt.Errorf("Invalid connection information")
}

func RegisterModule( m Module ) error {
	for _, m2 := range modules {
		if m.Name == m2.Name {
			return fmt.Errorf("[RegisterModule] Module is already registered")
		}
	}
	modules = append( modules, m )
	return nil
}

func UnregisterModule( name string ) error {
	idx := -1
	for i, m := range modules {
		if m.Name == name {
			idx = i
			break
		}
	}
	if idx >= 0 {
		modules = append( modules[:idx], modules[idx+1:]... )
		return nil
	}
	return fmt.Errorf("[UnregisterModule] Module not found.")
}

func InitAllModules() {
	RegisterModule( Module{"github", github.SupportedExt, github.NewGitHubConn } )
	RegisterModule( Module{"huggingface", huggingface.SupportedExt, huggingface.NewHuggingfaceConn } )
	RegisterModule( Module{"gitea", gitea.SupportedExt, gitea.NewGiteaConn } )
	RegisterModule( Module{"email", email.SupportedExt, email.NewEmailConn } )
	//RegisterModule( Module{"bluetooth", bluetooth.SupportedExt, bluetooth.NewBluetoothConn } )
	//RegisterModule( Module{"p2p", p2p.SupportedExt, p2p.NewP2PConn } )
	RegisterModule( Module{"tls", tls.SupportedExt, tls.NewNetConn } )
	RegisterModule( Module{"ssh", ssh.SupportedExt, ssh.NewSshConn} )
}
