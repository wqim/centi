package github
import (
	"fmt"
	"encoding/json"
	"encoding/base64"
	"strings"

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
	SupportedExt = []string{"png", "wav", "zip", "rar", "tar.gz", "bin", "exe"}
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

type GitHubConfig struct {
	Token		string		`json:"token"`
	Username	string		`json:"username"`
	PkBranch	string		`json:"pk_branch"`	// branch in which we are sending public key
	ReceiveBranch	string		`json:"recv_branch"`	// branch from which we collect messages
	SendBranch	string		`json:"send_branch"`	// branch to which we are committing messages
	ReceiveFrom	[]string	`json:"receive_from"`	// repositories to receive messages from
	SendTo		[]string	`json:"send_to"`	// repositories to upload messages to
	UserAgent	string		`json:"user_agent"`
	Authed		bool		`json:"authorized"`	// use authorisation for requests to get someone's repository files
								// (can raise the rate limits)
}

type GitHubConfigString struct {
	Token		string		`json:"token"`		// github api token
	Username	string		`json:"username"`	// repository in format 'username/repo_name
	PkBranch	string		`json:"pk_branch"`	// branch in which we are sending public key
	ReceiveBranch	string		`json:"recv_branch"`	// branch from which we collect messages
	SendBranch	string		`json:"send_branch"`	// branch to which we are committing messages
	RChannels	string		`json:"rchannels"`	// list of repositories to use for network communications
	SChannels	string		`json:"schannels"`	// optional, use specified file names for communications.if nil, generates filename
	UserAgent	string		`json:"user_agent"`	// user-agent which must be set in header
	Authed		string		`json:"authorized"`
}

type GitHubConn struct {
	config		GitHubConfig
	baseUrl		string
	methods		map[string]string
	endpoints	map[string]string
	headers		map[string]string
	repos		[]Repository
	channels	[]config.Channel
}


func NewGitHubConn( args map[string]string, channels []config.Channel ) (protocol.Connection, error) {
	
	var conf GitHubConfigString
	var config GitHubConfig
	conn := GitHubConn{}

	// check if configuration is more or less corrent
	tmp, err := json.Marshal(args)
	if err != nil {
		return conn, err
	}
	if err := json.Unmarshal( tmp, &conf ); err != nil {
		return conn, err
	}

	config.Token = conf.Token
	config.Username = conf.Username
	config.PkBranch = conf.PkBranch
	config.ReceiveBranch = conf.ReceiveBranch
	config.SendBranch = conf.SendBranch
	config.ReceiveFrom = strings.Split( conf.RChannels, "," )
	config.SendTo = strings.Split( conf.SChannels, "," )
	config.UserAgent = conf.UserAgent
	config.Authed = false
	if strings.ToLower( conf.Authed ) == "true" {
		config.Authed = true
	}

	conn.config = config
	conn.channels = channels
	conn.baseUrl = "https://api.github.com"
	conn.methods = map[string]string{
		general.ListKey: "GET",
		general.SendKey: "PUT",
		general.RecvKey: "GET",
		general.DeleteKey: "DELETE",
		general.CreateChanKey: "POST",
		general.DeleteChanKey: "DELETE",
		general.CommitKey: "",
	}
	conn.endpoints = map[string]string{
		general.ListKey: "/repos/%s/git/trees/%s?recursive=1",
		general.SendKey: "/repos/%s/contents/%s",
		general.RecvKey: "/repos/%s/contents/%s",
		general.DeleteKey: "/repos/%s/contents/%s",
		general.CreateChanKey: "/user/repos",
		general.DeleteChanKey: "/repos/%s",
		general.CommitKey: "",
	}
	conn.headers = map[string]string{
		"Authorization": "token " + conf.Token,
		"Content-Type": "application/json",
		"User-Agent": conf.UserAgent,
		"Accept": "application/json",
		"X-GitHub-Api-Version": "2022-11-28",
	}
	return conn, nil
}

func(g GitHubConn) InitChannels() error {
	for _, c := range g.channels {
		if err := g.CreateChannel( &c ); err != nil {
			return err
		}
	}
	return nil
}

func(g GitHubConn) DeleteChannels() error {
	for _, c := range g.channels {
		if err := g.DeleteChannel( &c ); err != nil {
			return err
		}
	}
	return nil
}

func(g GitHubConn) DistributePk( p *config.DistributionParameters, pk []byte ) error {
	// currently we are able to upload public key via following methods:
	// 1. hosting it in the /user/user repository's README file
	repoName := g.config.Username + "/" + g.config.Username
	filepath := "README.md"
	channel := &config.Channel { repoName, map[string]string{ PrivateKey: "false"} }

	// TODO: add steganography support
	readmeContents := base64.StdEncoding.EncodeToString( pk )
	if err := g.UploadFile( repoName, g.config.SendBranch, filepath, "", []byte( readmeContents) ); err != nil {
		g.CreateChannel( channel )
		return g.UploadFile( repoName, g.config.SendBranch, filepath, "", []byte( readmeContents) )
	}
	return nil
}

func(gh GitHubConn) CollectPks( p *config.DistributionParameters ) ([]protocol.KnownPk, error) {
	keys := []protocol.KnownPk{}
	var finalError error
	for _, repoName := range gh.config.ReceiveFrom {
		file, err := gh.DownloadFile( repoName, gh.config.ReceiveBranch, "README.md" )
		if err != nil {
			finalError = err
		} else if file != nil { // yes, we must check for the public key size...
			
			util.DebugPrintln("[111111111] Downloaded file")

			decoded, err := base64.StdEncoding.DecodeString( string(file) )
			if err == nil {
				if decoded != nil && len(decoded) == cryptography.PkSize {
					keys = append( keys, protocol.KnownPk{
						"github",
						util.GenID(),	// generate a random key alias
						file,
					})
				} else {
					util.DebugPrintln("[-------] Length of received public key:", len(decoded), "/", cryptography.PkSize )
				}
			} else {
				util.DebugPrintln("[------------] Failed to decode file:", err, ", length of file content:", len(file) )
			}
		}
	}
	util.DebugPrintln("[++++++] Length of received keys:", len(keys))
	return keys, finalError
}

// send message to the platform.
func(g GitHubConn) Send( msg *protocol.Message ) error {
	sha := ""
	if util.MapContains( msg.Args, ShaKey ) {
		sha = msg.Args[ ShaKey ]
	}

	//msgData := base64.StdEncoding.EncodeToString( msg.Data )
	return g.UploadFile( msg.Args[ RepoKey ], g.config.SendBranch, msg.Args[ FileKey ], sha, msg.Data )
}

/*
 * the following 2 functions do not require specification of data, so
 * we can just walk through random or known repositories and
 * collect all the files we did not download before
 */
func(g GitHubConn) RecvAll() ([]*protocol.Message, error) {

	var finalError error

	// do not have a list of repositories yet.
	if g.repos == nil {
		g.repos = []Repository{}
	}

	messages := []*protocol.Message{}
	// for eveery known Centi traffic repository

	for _, repo := range g.config.ReceiveFrom {
		// find the repository by name
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

		util.DebugPrintln("Collecting messages from ", repo)

		// get the list of all the files and their hashes (without content)
		allFiles, newSha, err := g.ListRepo( repo, g.config.ReceiveBranch )
		if err == nil {
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
						f.Content, err = base64.StdEncoding.DecodeString( string(content) )
						if err == nil {
							msg := &protocol.Message{
								f.Name,
								g.Name(),
								f.Content,
								protocol.UnknownSender, //???
								false,
								map[string]string{
									ShaKey: f.Sha,
									RepoKey: repo,
									BranchKey: g.config.ReceiveBranch,
									FileKey: f.Name,
								},
							}
							messages = append( messages, msg )
						} else {
							finalError = err
						}
					} else {
						finalError = err
					}
				}
			}
		} else {
			finalError = err
		}
	}
	return messages, finalError
}

func(g GitHubConn) PrepareToDelete( data []byte ) (*protocol.Message, error) {
	// TODO:
	return nil, nil
}

// delete message (file) from channel (repository)
func(g GitHubConn) Delete( msg *protocol.Message ) error {
	var err error
	sha := ""
	if msg != nil {
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

		g.headers["Accept"] = "application/vnd.github+json"
		_, err = g.sendRequest( url, general.DeleteKey, packed )
		g.headers["Accept"] = "application/json"
	}
	return err
}


func(g GitHubConn) CreateChannel( c *config.Channel ) error {
	parts := strings.Split( c.Name, "/" )
	if len(parts) == 2 && util.MapContains( c.Args, PrivateKey ) == true {

		private := (strings.ToLower( c.Args[ PrivateKey ] ) == "true")
		repo := GHRepo {
			Name:		parts[1],
			Description:	parts[1],
			Private: 	private,
			HasIssues:	false,
			HasProjects:	false,
			HasWiki:	false,
			AutoInit:	true,
		}
		data, err := json.Marshal( repo )
		if err != nil {
			return err
		}
		args := map[string]string{FileKey: c.Name}
		url := g.formatURL( general.CreateChanKey, args )
		_, err = g.sendRequest( url, general.CreateChanKey, data )
		return err
	}
	return fmt.Errorf("invalid channel name")
}


func(g GitHubConn) DeleteChannel( c *config.Channel ) error {

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
	args := map[string]string{ RepoKey: c.Name, ShaKey: sha }
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
		_, err = g.sendRequest( url, general.DeleteChanKey, data )
		return err
	}
	return fmt.Errorf("[GitHubConn::DeleteChannel] Invalid channel name.")
}

func(gh GitHubConn) MessageFromBytes( data []byte ) (*protocol.Message, error) {
	repoName := gh.config.SendTo[ util.RandInt(len(gh.config.SendTo)) ]
	msg := &protocol.Message{
		"",
		gh.Name(),
		data,
		protocol.UnknownSender,
		true,
		map[string]string{
			FileKey: util.GenFilename( "test", "go" ),
			RepoKey: repoName,
		},
	}
	return msg, nil
}

func(gh GitHubConn) Name() string {
	return "github"
}

func(gh GitHubConn) GetSupportedExtensions() []string {
	return SupportedExt
}
