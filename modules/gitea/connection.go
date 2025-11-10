package gitea
import (
	//"os"
	"fmt"
	"sync"
	"encoding/json"
	//"encoding/base64"
	"strings"

	//"centi/stegano/img"

	"centi/util"
	"centi/config"
	"centi/protocol"
	"centi/cryptography"
	"centi/modules/general"
)

/*
 * this file contains all the structures and functions related to protocol.Connection
 * functionality
 */
var (
	// add other files formats
	SupportedExt = []string{"png", "wav", "tar.gz", "zip", "rar", "bin"}
)

type File struct {
	Sha	string
	Name	string
	Content	[]byte
}
type Repository	struct {
	Name	string
	Sha	string
	Files	[]File
}

type GiteaConfig struct {
	Token		string		`json:"token"`
	AvatarPath	string		`json:"avatar_path"`	// path to avatar image in which the public key will be hidden. must be a png image.
	ReceiveBranch	string		`json:"recv_branch"`	// branch from which we collect messages
	SendBranch	string		`json:"send_branch"`	// branch to which we are committing messages
	ReceiveFrom	[]string	`json:"receive_from"`	// repositories to receive messages from
	SendTo		[]string	`json:"send_to"`	// repositories to upload messages to
}

type GiteaConfigString struct {
	Token		string		`json:"token"`		// github api token
	AvatarPath	string		`json:"avatar_path"`	// path to avatar image in which the public key will be hidden. must be a png image.
	ReceiveBranch	string		`json:"recv_branch"`	// branch from which we collect messages
	SendBranch	string		`json:"send_branch"`	// branch to which we are committing messages
	RChannels	string		`json:"rchannels"`	// list of repositories to use for network communications
	SChannels	string		`json:"schannels"`	// optional, use specified file names for communications.if nil, generates filename
	UserAgent	string		`json:"user_agent"`	// user-agent which must be set in header
}

type GiteaConn struct {
	config		GiteaConfig
	baseUrl		string
	methods		map[string]string
	endpoints	map[string]string
	headers		map[string]string
	repos		[]Repository
	channels	[]config.Channel
	
	sentMessages	map[string]string	// hash:filename
	messagesMtx	sync.Mutex		// mutex for sent messages map
}


func NewGiteaConn( args map[string]string, channels []config.Channel ) (protocol.Connection, error) {
	
	var conf GiteaConfigString
	var config GiteaConfig
	conn := GiteaConn{}

	// check if configuration is more or less corrent
	tmp, err := json.Marshal(args)
	if err != nil {
		return conn, err
	}
	if err := json.Unmarshal( tmp, &conf ); err != nil {
		return conn, err
	}

	config.Token = conf.Token
	config.AvatarPath = conf.AvatarPath
	config.ReceiveBranch = conf.ReceiveBranch
	config.SendBranch = conf.SendBranch
	config.ReceiveFrom = strings.Split( conf.RChannels, "," )
	config.SendTo = strings.Split( conf.SChannels, "," )

	conn.sentMessages = map[string]string{}
	conn.messagesMtx = sync.Mutex{}
	conn.config = config
	conn.channels = channels
	conn.baseUrl = "https://gitea.com"
	conn.methods = map[string]string{
		general.ListKey: "GET",
		general.SendKey: "POST",
		general.RecvKey: "GET",
		general.DeleteKey: "DELETE",
		general.CreateChanKey: "POST",
		general.DeleteChanKey: "DELETE",
	}
	conn.endpoints = map[string]string{
		general.ListKey: "/api/v1/repos/%s/git/trees/%s?recursive=1",
		general.SendKey: "/api/v1/repos/%s/contents/%s",
		general.RecvKey: "/api/v1/repos/%s/contents/%s",
		general.DeleteKey: "/api/v1/repos/%s/contents/%s",
		general.CreateChanKey: "/api/v1/user/repos",
		general.DeleteChanKey: "/api/v1/repos/%s",
	}
	conn.headers = map[string]string{
		"Authorization": "bearer " + conf.Token,
		"Content-Type": "application/json",
		"User-Agent": conf.UserAgent,
	}
	return conn, nil
}

func(g GiteaConn) InitChannels() error {
	for _, c := range g.channels {
		if err := g.CreateChannel( &c ); err != nil {
			return err
		}
	}
	return nil
}

func(g GiteaConn) DeleteChannels() error {
	for _, c := range g.channels {
		if err := g.DeleteChannel( &c ); err != nil {
			return err
		}
	}
	return nil
}

/*
func(g GiteaConn) DistributePk( p *config.DistributionParameters, pk []byte ) error {
	// gitea supports avatar upload via api
	// let's use this advantage in order to distribute public keys
	// with help of steganography.
	url := g.baseUrl + "/api/v1/user/avatar"
	imgBytes, err := os.ReadFile( g.config.AvatarPath )
	if err != nil {
		return err
	}

	//util.DebugPrintln("Distributed public key:", base64.StdEncoding.EncodeToString( pk.Content[:10] ) )
	newImg, err := img.EncodeWithLSB( img.RMode | img.GMode | img.BMode, pk, imgBytes )
	if err != nil {
		return err
	}
	//os.WriteFile("test/tmp-avatar-encoded.png", newImg, 0660) 	// for debug
	base64Image := base64.StdEncoding.EncodeToString( newImg )
	payload := map[string]string{
		"image": base64Image,
	}
	data, err := json.Marshal( payload )
	if err != nil {
		return err
	}

	resp, err := general.HTTPRequest( url, "POST", data, g.headers )
	if err != nil {
		return err
	}
	//util.DebugPrintln("Avatar upload response:", string(resp))
	return g.isError( resp )
}

func(g GiteaConn) CollectPks( p *config.DistributionParameters ) ([]protocol.KnownPk, error) {
	keys := []protocol.KnownPk{}
	var finalError error
	// get avatar of every known user and extract public key from it.
	users := g.Users()
	for _, u := range users {
		imgBytes, err := g.DownloadAvatar( u )
		if err != nil {
			finalError = err
			//util.DebugPrintln("Failed to download avatar:", finalError)
		} else {
			//os.WriteFile("test/tmp-avatar.png", imgBytes, 0660) 	// for debug
			pkBytes, err := img.DecodeFromLSB( img.RMode | img.GMode | img.BMode, imgBytes )
			if err == nil && len(pkBytes) > cryptography.PkSize {
				keys = append( keys, protocol.KnownPk{
					g.Name(),
					g.Name() + ":" + u,
					pkBytes,
				})
				//util.DebugPrintln("Collected public key:", base64.StdEncoding.EncodeToString( pkBytes[:10] ) )
			}
		}
	}
	return keys, finalError
}
*/

// send message to the platform.
func(g GiteaConn) Send( msg *protocol.Message ) error {
	g.messagesMtx.Lock()
	defer g.messagesMtx.Unlock()

	sha := ""
	if util.MapContains( msg.Args, ShaKey ) {
		sha = msg.Args[ ShaKey ]
	}

	msg.Args[ FileKey ] = msg.Name
	filename := msg.Args[ RepoKey ] + ":" + msg.Args[ FileKey ]
	hash := cryptography.Hash( msg.Data )
	g.sentMessages[ hash ] = filename

	//util.DebugPrintln("[GiteaConn::Send()] data =", string(msg.Data) )
	//msgData := base64.StdEncoding.EncodeToString( msg.Data )
	return g.UploadFile( msg.Args[ RepoKey ], g.config.SendBranch, msg.Args[ FileKey ], sha, msg.Data )
}

/*
 * the following 2 functions do not require specification of data, so
 * we can just walk through random or known repositories and
 * collect all the files we did not download before
 */
func(g GiteaConn) RecvAll() ([]*protocol.Message, error) {

	var finalError error

	// do not have a list of repositories yet.
	if g.repos == nil {
		g.repos = []Repository{}
	}

	messages := []*protocol.Message{}
	// for eveery known Centi traffic repository
	for _, repo := range g.config.ReceiveFrom {
		// find the repository by name
		//util.DebugPrintln("Listing files in the ", repo)
		repoIdx := -1
		repository := Repository{ repo, "", nil }
		found := false
		for index, rep := range g.repos {
			if rep.Name == repo {
				// update repository
				found = true
				repository = rep
				repoIdx = index
				break
			}
		}

		//util.DebugPrintln("Collecting messages from ", repo)

		// get the list of all the files and their hashes (without content)
		allFiles, newSha, err := g.ListRepo( repo, g.config.ReceiveBranch )
		if err == nil {
			//util.DebugPrintln("Length of all the files in the repository is ", len(allFiles) )
			// if no error occured, check if there are new files or changed ones
			if (newSha != repository.Sha) || (found == false) {	// repository was changed

				files := g.GetNewFiles( repository, allFiles )
				// no repository found in the connection
				if found == false {
					rep := Repository{
						repo,
						newSha,
						files,
					}
					g.repos = append( g.repos, rep )
				} else {
					// update repository info
					g.repos[ repoIdx ].Files = allFiles
					g.repos[ repoIdx ].Sha = newSha
				}

				// update file's contents and also get the list
				// of incoming messages
				for _, f := range files {
					
					content, err := g.DownloadFile( repo, g.config.ReceiveBranch, f.Name )
					if err == nil {
						//f.Content, err = base64.StdEncoding.DecodeString( string(content) )
						//if err == nil {
							msg := &protocol.Message{
								f.Name,
								g.Name(),
								content, //f.Content,
								protocol.UnknownSender, //???, everything is ok, we detemine sender by key
								false,
								map[string]string{
									ShaKey: f.Sha,
									RepoKey: repo,
									BranchKey: g.config.ReceiveBranch,
									FileKey: f.Name,
								},
							}
							messages = append( messages, msg )
						/*else {
							util.DebugPrintln("Failed to decode file content:", err)
							util.DebugPrintln("File content:", string(content))
							// finalError = err
						}*/
					} else {
						finalError = err
					}
				}
			}
		} else {
			finalError = err
		}
	}
	//util.DebugPrintln("Final length of messages:", len(messages) )
	return messages, finalError
}


// delete message (file) from channel (repository)
func(g GiteaConn) Delete( msg *protocol.Message ) error {
	var err error
	if msg != nil {
		
		for _, schan := range g.config.SendTo {
			if g.FileExists( schan, g.config.SendBranch, msg.Name ) {
				msg.Args[ RepoKey ] = schan
				break
			}
		}
		
		sha := ""
		if util.MapContains( msg.Args, ShaKey ) == true {
			sha = msg.Args[ ShaKey ]
		} else {
			sha, err = g.getFileSha( msg.Args[ RepoKey ], msg.Args[ FileKey ] )
			if err != nil {
				return err
			}
		}
		data := map[string]string{
			"message": "Delete " + msg.Args[ FileKey ],
			"sha": sha,
		}
		packed, err := json.Marshal( data )
		if err != nil {
			return err
		}
		args := map[string]string{FileKey: msg.Args[ FileKey ], ShaKey: sha, RepoKey: msg.Args[ RepoKey ] }
		url := g.formatURL( general.DeleteKey, args )
		//util.DebugPrintln("[GiteaConn]: url =", url, "; args =", args)
		_, err = g.sendRequest( url, general.DeleteKey, packed )
	}
	return err
}


func(g GiteaConn) CreateChannel( c *config.Channel ) error {
	parts := strings.Split( c.Name, "/" )
	if len(parts) == 2 && util.MapContains( c.Args, PrivateKey ) == true {

		private := (strings.ToLower( c.Args[ PrivateKey ] ) == "true")
		repo := map[string]any {
			"auto_init": true,
			"name": parts[1],
			"private": private,
		}
		data, err := json.Marshal( repo )
		if err != nil {
			return err
		}
		args := map[string]string{FileKey: c.Name}	// ?
		url := g.formatURL( general.CreateChanKey, args )
		_, err = g.sendRequest( url, general.CreateChanKey, data )
		//util.DebugPrintln("[CreateChannel]:", string(resp))
		return err
	}
	return fmt.Errorf("invalid channel name")
}


func(g GiteaConn) DeleteChannel( c *config.Channel ) error {

	var err error
	sha := ""
	if util.MapContains( c.Args, ShaKey ) {
		sha = c.Args[ ShaKey ]
	} else {
		sha, err = g.getRepoSha( c.Name )
		if err != nil {
			return err
		}
	}
	args := map[string]string{ RepoKey: c.Name, ShaKey: sha, FileKey: "" }
	url := g.formatURL( general.DeleteChanKey, args )

	parts := strings.Split( c.Name, "/" )
	if len(parts) == 2 {
		mdata := map[string]string{
			"owner": parts[0],
			"repo": parts[1],
		}
		data, err := json.Marshal( mdata )
		if err != nil {
			return err
		}
		// does this request really requires data?
		_, err = g.sendRequest( url, general.DeleteChanKey, data )
		return err
	}
	return fmt.Errorf("[GiteaConn::DeleteChannel] Invalid channel name.")
}

func(g GiteaConn) MessageFromBytes( data []byte ) (*protocol.Message, error) {
	repoName := g.config.SendTo[ util.RandInt(len(g.config.SendTo)) ]
	msg := &protocol.Message{
		"",
		g.Name(),
		data,
		protocol.UnknownSender,
		true,	// does not really matter?
		map[string]string{
			FileKey: util.GenFilename( "test", "base64" ),
			RepoKey: repoName,
		},
	}
	return msg, nil
}

func(g GiteaConn) Name() string {
	return "gitea"
}

func(g GiteaConn) GetSupportedExtensions() []string {
	return SupportedExt
}
