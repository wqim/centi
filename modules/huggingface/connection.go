package huggingface
import (
	"fmt"
	"strings"
	"encoding/json"
	"encoding/base64"

	"centi/util"
	"centi/config"
	"centi/protocol"
	"centi/cryptography"
	"centi/modules/general"
)

// this file contains all the functions related to protocol.Connection interface

var (
	SupportedExt = []string{"h5", "zip", "csv", "xls", "tar.gz", "rar", "bin"}
)

type HuggingfaceConfig struct {
	Token		string
	UserAgent	string
	PkChannel	string		// public key repository
	PkFile		string		// public key file
	PkBranch	string
	RecvBranch	string
	SendBranch	string
	ReceiveFrom	[]string	// list of repositories to collect messages from
	SendTo		[]string	// list of repositories to send messages to
	PkRepoType	string		// repository type of public key repo: dataset, model or space
	RecvRepoType	string
	SendRepoType	string
	Authed		bool		// use authentication in order to list files (may increase rate limits)
}

type HuggingfaceConfigString struct {
	Token		string		`json:"token"`
	UserAgent	string		`json:"user_agent"`
	PkChannel	string		`json:"pk_channel"`
	PkFile		string		`json:"pk_file"`
	PkBranch	string		`json:"pk_branch"`
	RecvBranch	string		`json:"recv_branch"`
	SendBranch	string		`json:"send_branch"`
	ReceiveFrom	string		`json:"receive_from"`
	SendTo		string		`json:"send_to"`
	PkRepoType	string		`json:"pk_repo_type"`
	RecvRepoType	string		`json:"recv_repo_type"`
	SendRepoType	string		`json:"send_repo_type"`
	Authed		string		`json:"authed"`
}

type HuggingfaceConn struct {
	config		HuggingfaceConfig
	baseUrl		string
	methods		map[string]string
	endpoints	map[string]string
	headers		map[string]string
	repos		[]Repository
	channels	[]config.Channel
}

func NewHuggingfaceConn( args map[string]string, channels []config.Channel ) (protocol.Connection, error) {
	
	conn := HuggingfaceConn{}
	
	data, err := json.Marshal( args )
	if err != nil {
		return conn, err
	}

	var config HuggingfaceConfig
	var conf HuggingfaceConfigString
	if err := json.Unmarshal( data, &conf ); err != nil {
		return conn, err
	}

	config.Token = conf.Token
	config.UserAgent = conf.UserAgent
	config.PkChannel = conf.PkChannel
	config.PkFile = conf.PkFile
	config.RecvBranch = conf.RecvBranch
	config.SendBranch = conf.SendBranch
	config.ReceiveFrom = strings.Split(conf.ReceiveFrom, ",")
	config.SendTo = strings.Split( conf.SendTo, "," )
	config.PkRepoType = conf.PkRepoType
	config.RecvRepoType = conf.RecvRepoType
	config.SendRepoType = conf.SendRepoType
	config.Authed = false
	if strings.ToLower( conf.Authed ) == "true" {
		config.Authed = true
	}

	conn.config = config
	conn.channels = channels
	conn.baseUrl = "https://huggingface.co"
	conn.methods = map[string]string{
		general.ListKey: "GET",
		general.SendKey: "POST",
		general.RecvKey: "GET",
		general.DeleteKey: "POST",
		general.CreateChanKey: "POST",
		general.DeleteChanKey: "DELETE",
		general.CommitKey: "POST",
	}
	conn.endpoints = map[string]string{
		general.ListKey: "/api/%s/%s/tree/main/%s?recursive=True&expand=True",
		general.SendKey: "/api/%s/%s/preupload/main",
		general.RecvKey: "/%s/%s/resolve/main/%s",
		general.DeleteKey: "/api/%s/%s/commit/main",
		general.CreateChanKey: "/api/repos/create",
		general.DeleteChanKey: "/api/repos/delete",
		general.CommitKey: "/api/%s/%s/commit/main",
	}
	conn.headers = map[string]string{
		"Authorization": "Bearer " + conf.Token,
		"Content-Type": "application/json",
		"User-Agent": conf.UserAgent,
	}
	return conn, nil
}

func(hf HuggingfaceConn) InitChannels() error {
	for _, c := range hf.channels {
		if err := hf.CreateChannel( &c ); err != nil {
			return err
		}
	}
	return nil
}

func(hf HuggingfaceConn) DeleteChannels() error {
	for _, c := range hf.channels {
		if err := hf.DeleteChannel( &c ); err != nil {
			return err
		}
	}
	return nil
}

func(hf HuggingfaceConn) DistributePk( p *config.DistributionParameters, pk []byte ) error {
	// possible ways of public key distribution in huggingface:
	// 1. (as in github) public key in the README file of specified repository
	// 2. embed in the model file of specified repository
	content := base64.StdEncoding.EncodeToString( pk )
	pkRepo := Repository{ hf.config.PkChannel, hf.config.PkRepoType, false, nil }
	if err := hf.UploadFile( pkRepo, hf.config.PkFile, []byte(content) ); err != nil {
		channel := &config.Channel{
			hf.config.PkChannel,
			map[string]string{
				PrivateKey: "false",
				RepoTypeKey: hf.config.PkRepoType,
			},
		}
		hf.CreateChannel( channel )
		return hf.UploadFile( pkRepo, hf.config.PkFile, []byte(content) )
	}
	return nil
}

func(hf HuggingfaceConn) CollectPks( p *config.DistributionParameters ) ([]protocol.KnownPk, error) {
	keys := []protocol.KnownPk{}
	var finalError error
	for _, repoName := range hf.config.ReceiveFrom {
		repository := Repository{
			repoName,
			hf.config.RecvRepoType,
			false,
			nil,
		}
		file, err := hf.DownloadFile( repository, hf.config.RecvBranch, hf.config.PkFile )
		if err != nil {
			finalError = err
		}

		decoded, err := base64.StdEncoding.DecodeString( string(file) )
		if err == nil {
			if len(decoded) == cryptography.PkSize {
				keys = append( keys, protocol.KnownPk{
					"huggingface",
					util.GenID(),	// generate an alias at random
					decoded,
				})
			} else {
				util.DebugPrintln("[------] Length of decoded data:", len(decoded), "/", cryptography.PkSize )
			}
		} else {
			util.DebugPrintln("[------] Failed to decode public key:", err)
			util.DebugPrintln("From server:", string(file))
		}
	}
	return keys, finalError
}

func(hf HuggingfaceConn) Send( msg *protocol.Message ) error {
	repo, err := hf.findRepoByName( msg.Args[RepoKey] )
	if err != nil {
		repo = Repository{
			msg.Args[RepoKey],
			hf.config.SendRepoType,
			false,
			nil,
		}
		hf.repos = append( hf.repos, repo )
	}
	content := base64.StdEncoding.EncodeToString( msg.Data )
	return hf.UploadFile( repo, msg.Args[FileKey], []byte(content) )
}

func(hf HuggingfaceConn) RecvAll() ( []*protocol.Message, error ) {
	
	var finalError error
	if hf.repos == nil {
		hf.repos = []Repository{}
	}

	messages := []*protocol.Message{}
	for _, repoName := range hf.config.ReceiveFrom {
		// get the repository
		repoIdx := -1
		repository := Repository{ repoName, hf.config.RecvRepoType, false, nil }	// default repository structure
		found := false
		for index, rep := range hf.repos {
			if rep.Name == repoName {
				repository = rep
				repoIdx = index
				break
			}
		}

		// list files in the repository
		util.DebugPrintln("Listing repository ", repoName)
		allFiles, err := hf.ListRepo( repository )
		if err == nil {
			
			files := hf.GetNewFiles( repository, allFiles )
			// update or append a repository
			if found == false {
				rep := Repository{
					repoName,
					hf.config.RecvRepoType,
					false,
					files,
				}
				hf.repos = append( hf.repos, rep )
			} else {
				hf.repos[ repoIdx ].Files = allFiles
			}
			
			// if there are new files, get their contents and put
			// it into messages.
			for _, f := range files {
				content, err := hf.DownloadFile( repository, f.Name, hf.config.RecvBranch )
				if err == nil {
					f.Content = content
					msg := &protocol.Message{
						hf.Name(),
						f.Content,
						protocol.UnknownSender, //???
						false,
						map[string]string{
							RepoKey: repoName,
							BranchKey: hf.config.RecvBranch,
							FileKey: f.Name,
						},
					}
					messages = append( messages, msg )
				} else {
					finalError = err
				}
			}
		} else {
			finalError = err
		}
	}
	return messages, finalError
}

func(hf HuggingfaceConn) PrepareToDelete( data []byte ) (*protocol.Message, error) {
	// TODO
	return nil, nil
}

func(hf HuggingfaceConn) Delete( msg *protocol.Message ) error {
	if msg != nil {
		kv := KeyValue{
			"header",
			map[string]string{
				"summary": "Delete " + msg.Args[ FileKey ],
				"description": "",
			},
		}
		data, err := json.Marshal( kv )
		if err != nil {
			return err
		}
		kv = KeyValue{
			"deletedFile",
			map[string]string{
				"path": msg.Args[ FileKey ],
			},
		}
		tmp, err := json.Marshal( kv )
		if err != nil {
			return err
		}
		data = append( data, 0x0a )
		data = append( data, tmp... )
		
		repo, err := hf.findRepoByName( msg.Args[RepoKey] )
		
		if err != nil {
			repo = Repository{
				msg.Args[RepoKey],
				hf.config.SendRepoType,
				false,
				nil,
			}
			hf.repos = append( hf.repos, repo )
		}
		
		url := hf.formatURL( repo, msg.Args[FileKey], general.DeleteKey )
		hf.headers["Content-Type"] = "application/x-ndjson"
		_, err = hf.sendRequest( url, general.DeleteKey, data )
		hf.headers["Content-Type"] = "application/json"
		return err
	}
	return nil
}


func(hf HuggingfaceConn) CreateChannel( c *config.Channel ) error {

	parts := strings.Split( c.Name, "/")
	if len(parts) == 2 {
		private := false
		if util.MapContains( c.Args, PrivateKey ) == true {
			if strings.ToLower( c.Args[ PrivateKey ] ) == "true" {
				private = true
			}
		}

		repoType := hf.config.SendRepoType
		if util.MapContains( c.Args, RepoTypeKey ) == true {
			repoType = c.Args[ RepoTypeKey ]
		}

		payload := HFRepo {
			c.Name,
			repoType,
			parts[1],
			private,
		}
		data, err := json.Marshal( payload )
		if err != nil {
			return err
		}

		repo, err := hf.findRepoByName( c.Name )
		if err != nil {	// if repository not found
			// create a new repository structure
			repo = Repository{
				c.Name,
				repoType,
				private,
				nil,
			}
			hf.repos = append( hf.repos, repo )
		}
		url := hf.formatURL( repo, "", general.CreateChanKey )
		_, err = hf.sendRequest( url, general.CreateChanKey, data )
		return err
	}
	return fmt.Errorf("Invalid channel name:" + c.Name)
}


func(hf HuggingfaceConn) DeleteChannel( c *config.Channel ) error {
	parts := strings.Split( c.Name, "/" )
	if len(parts) == 2 {
		payload := map[string]string{
			"name": parts[1],
			"type": "dataset",
			"organization": parts[0],
		}
		data, err := json.Marshal( payload )
		if err != nil {
			return err
		}
		repo, err := hf.findRepoByName( c.Name )
		if err != nil {	// we know that this repository exists
			repo = Repository{
				c.Name,
				hf.config.SendRepoType,
				true,
				nil,
			}
		}
		url := hf.formatURL( repo, "", general.DeleteChanKey )
		_, err = hf.sendRequest( url, general.DeleteChanKey, data )
		return err
	}
	return fmt.Errorf("Invalid channel name: " + c.Name)
}

func(hf HuggingfaceConn) MessageFromBytes( data []byte ) (*protocol.Message, error) {
	repoName := hf.config.SendTo[ util.RandInt(len(hf.config.SendTo)) ]
	msg := &protocol.Message{
		hf.Name(),
		data,
		protocol.UnknownSender,
		true,
		map[string]string{
			FileKey: util.GenFilename("data-file-", "csv"),
			RepoKey: repoName,
		},
	}
	return msg, nil
}

func(hf HuggingfaceConn) Name() string {
	return "huggingface"
}
