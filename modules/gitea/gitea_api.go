package gitea
import (
	//"os"
	"fmt"
	"encoding/hex"
	"encoding/json"
	"encoding/base64"
	//"io/ioutil"
	"strings"
	//"net/http"
	"crypto/sha1"

	"centi/util"
	"centi/modules/general"
)

/*
 * this file contains all the functions which are required to
 * make basic github api operations(read/write/delete file, etc).
 */
const (
	ShaKey = "sha"
	RepoKey = "channel"
	FileKey = "name"
	BranchKey = "branch"
	PrivateKey = "private"	// no, it's not a private key from cryptography things...
)

func(g GiteaConn) DownloadAvatar( user string ) ([]byte, error) {
	// download avatar from user's html page
	url := "https://gitea.com/" + user
	resp, err := general.HTTPRequest( url, "GET", nil, map[string]string{} )
	if err != nil {
		return nil, err
	}
	doc := string(resp)
	// [warning] this one can change in the future.
	magicString := "<img loading=\"lazy\" alt class=\"ui avatar tw-align-middle\" src=\"" 
	if strings.Contains( doc, magicString ) == false {
		return nil, fmt.Errorf("Invalid HTML Page")
	}

	// extract image link
	doc = strings.Split( doc, magicString )[1]
	url = strings.Split(doc, "\"")[0]	// extract link
	if strings.HasPrefix( url, "http" ) == false {
		url = "https://gitea.com" + url
	}
	// download the image itself.
	resp, err = general.HTTPRequest( url, "GET", nil, map[string]string{} )
	return resp, err

}

func(g GiteaConn) DownloadFile( repoName string, branch string, filepath string ) ([]byte, error) {

	// basically, download a raw file. we even don't need any headers here
	// so it's not encounted in API calls.
	url := "https://gitea.com/" + repoName + "/raw/branch/" + branch + "/" + filepath
	resp, err := general.HTTPRequest( url, "GET", nil, map[string]string{} )
	if err != nil {
		return nil, err
	}
	return resp, g.isError( resp )
}

// amount of api calls: 2 - 4 (depend on situation)
func(g GiteaConn) UploadFile( repoName, branch, filepath, sha string, data []byte ) error {

	content := base64.StdEncoding.EncodeToString( data )
	args := map[string]string{
		FileKey: filepath,
		RepoKey: repoName,
	}
	url := g.formatURL( general.SendKey, args )

	// actually upload the file
	if g.FileExists( repoName, g.config.SendBranch, filepath ) == true { // update file

		// may never happen btw...
		util.DebugPrintf( "[GiteaConn::UploadFile] found file %s:%s, trying to update it.\n", repoName, filepath )
		jsonData := map[string]string{
			"message": "Update " + filepath,
			"content": content,
		}
		data, err := json.Marshal( jsonData )
		if err != nil {
			return err
		}
		resp, err := general.HTTPRequest( url, "PUT", data, g.headers )
		if err != nil {
			return err
		}
		return g.isError( resp )

	} else {	// create file
		//util.DebugPrintf( "[GiteaConn::UploadFile] 404 file %s:%s NOT found, trying to create it.\n", repoName, filepath )
		if sha == "" {
			sha, _ = g.getFileSha( repoName, filepath )
		}
		jsonData := map[string]string{
			"message": "Upload " + filepath,
			"content": content,
			"branch": branch,
			"sha": sha,
		}
		packed, err := json.Marshal( jsonData )
		if err != nil {
			return err
		}
		_, err = g.sendRequest( url, general.SendKey, packed )
		return err
	}
}

// amount of api calls: 1 - 2
func(g GiteaConn) ListRepo( repoName, branch string ) ([]File, string, error) {
	// get a sha file of the repository
	files := []File{}
	repoSha, err := g.getRepoSha( repoName )
	if err != nil {
		return nil, "", err
	}
	//util.DebugPrintln("Repository [SHA]:", repoSha)
	args := map[string]string{
		ShaKey: repoSha,
		RepoKey: repoName,
	}
	url := g.formatURL( general.ListKey, args )
	res, err := g.sendRequest( url, general.ListKey, nil )
	if err != nil {
		return nil, "", err
	}

	// parse repository content, res - GTList structure
	var list GTList
	if err = json.Unmarshal( res, &list ); err != nil {
		return nil, "", err
	}
	//util.DebugPrintln("Repository files:")
	for _, ghFile := range list.Tree {
		if ghFile.Type == "blob" {	// an actual file
			//fmt.Println("\t+", ghFile.Path)
			files = append(files, File{ ghFile.Sha, ghFile.Path, nil })
		}
	}
	return files, repoSha, nil
}


// calculates which files were not received yet...
// is it really useful...?
func(g GiteaConn) GetNewFiles( repo Repository, allFiles []File ) []File {
	if repo.Files == nil || len(repo.Files) == 0 {
		return allFiles
	}
	files := []File{}
	for _, f := range allFiles {
		alreadyFound := false
		for _, rf := range repo.Files {
			// file was updated
			if (rf.Name == f.Name) {
				alreadyFound = true
				if (rf.Sha != f.Sha) {
					files = append( files, f )
				}
			}
		}
		if alreadyFound == false {
			// new file found
			files = append( files, f )
		}
	}
	return files
}

// amount of api calls: 0 (because of direct file download) // 1 - 2 ( because of ListRepo )
func(g GiteaConn) FileExists( repo, branch, path string ) bool {
	// pick the way of checking file existense
	//if util.RandInt(2) == 1 {
		// check if file is in the repo
		//util.DebugPrintln("Listing repository to check if file exists")
		// THIS::::: WAS CORRECT::::
		/*files, _, err := g.ListRepo( repo, branch )
		if err == nil {
			for _, f := range files {
				if f.Name == path {
					return true
				}
			}
		}*/
	//} else {
		// direct file download
		//util.DebugPrintln("Trying direct file download...")
		resp, err := g.DownloadFile( repo, branch, path )
		if err == nil && resp != nil && len(resp) > 0 {
			// check response content
			/*if strings.Contains( string(resp), ">404 Not Found" ) == true {
				return false
			}*/
			//util.DebugPrintln(string( resp ) )
			//util.DebugPrintln( resp )
			if string(resp) == "Not found.\n" {
				//util.DebugPrintln("[File not found!!!]")
				return false
			}
			//os.WriteFile("test/tmp.html", resp, 0660)
			return true
		}
	//} */
	return false
}

// amount of api calls: 1
func(g GiteaConn) sendRequest( url, action string, data []byte ) ([]byte, error) {

	//fmt.Println("[GiteaConn::sendRequest]: ", g.methods[action], url )
	resp, err := general.HTTPRequest( url, g.methods[ action ], data, g.headers )
	if err != nil {
		return nil, err
	}

	if err = g.isError( resp ); err != nil {
		return nil, err
	}
	return resp, nil
}

// amount of api calls: 0-1
func(g GiteaConn) formatURL( action string, args map[string]string ) string {
	// format URL before the request
	url := strings.Replace( g.endpoints[ action ], "%s", args[ RepoKey ], 1 )
	switch action {
	case general.ListKey:
		if args != nil {
			if util.MapContains(args, ShaKey) {
				url = strings.Replace( url, "%s", args[ ShaKey ], 1 )
			} else {
				sha, err := g.getRepoSha( args[ RepoKey ] )
				if err == nil {
					url = strings.Replace( url, "%s", sha, 1 )
				}
			}
		}
	default:
		if util.MapContains( args, FileKey ) {
			url = strings.Replace( url, "%s", args[ FileKey ], 1 )
		}
	}
	return g.baseUrl + url
}

// amount of api calls: 1
func(g GiteaConn) getRepoSha( channel string ) (string, error) {
	// just gen the sha of the repostory via api.
	url := g.baseUrl + "/api/v1/repos/" + channel + "/git/refs/heads/main"
	res, err := general.HTTPRequest( url, g.methods[ general.RecvKey ], nil, g.headers )
	if err != nil {
		util.DebugPrintln("[GiteaConn::getRepoSha] Failed to get repository sha:", err)
		util.DebugPrintln( string(res) )
		return "", err
	}
	var fc []GTBranch
	if err := json.Unmarshal( res, &fc ); err != nil {
		util.DebugPrintln("[GiteaConn::genRepoSha] Result of request:", string( res ) )
		return "", err
	}
	return fc[0].Object.Sha, nil
}

// amount of api calls: 1
func(g GiteaConn) getFileSha( channel string, filename string ) (string, error) {
	// list file in the repository
	// todo: fix this (???)
	args := map[string]string{
		FileKey: filename,
		RepoKey: channel,
	}
	url := g.formatURL( general.ListKey, args )
	res, err := general.HTTPRequest( url, g.methods[ general.ListKey ], nil, g.headers )
	if err != nil {
		return "", err
	}
	if err = g.isError( res ); err != nil {
		return "", err
	}
	listing, err := g.parseListing( res )
	if err != nil {
		return "", err
	}
	for _, f := range listing {
		if len(f) == 2 { // already true according to data supplied, but still better safe than sorry
			if f[0] == filename {
				return f[1], nil
			}
		}
	}
	return "", fmt.Errorf("[getFileSha]: sha not found.")
}

// parse listing of files
func(g GiteaConn) parseListing( data []byte ) ([][]string, error) {
	// function which parses the listing of files in the repository
	// returns list of []string containing filename at 0 index and sha at 1 index
	var filesTree GTList
	if err := json.Unmarshal( data, &filesTree ); err != nil {
		return nil, err
	}

	res := [][]string{}
	for _, f := range filesTree.Tree {
		if f.Type == "blob" {
			res = append( res, []string{ f.Path, f.Sha } )
		}
	}
	return res, nil
}

// check if the server's response contains an error
func(g GiteaConn) isError( resp []byte ) error {
	var x map[string]any
	if err := json.Unmarshal( resp, &x ); err != nil {
		//fmt.Println("(gitea) Plaintext data:", string(data))
		return nil // normal text/binary data
	}
	for k, v := range x {
		if (k == "error") || (k == "message") {
			m, ok := v.(string)
			if ok {
				if m == "The target couldn't be found." {
					return fmt.Errorf( "NotFoundError" )
				} else if strings.HasPrefix( m, "repository file does not exist") {
					return fmt.Errorf( "NotFoundError" )
				}

				fmt.Println("(gitea::isError) Error:", string(resp))
				return fmt.Errorf( m )
			}
		}
	}
	//fmt.Println("(gitea): no errors detected")
	return nil
}

func(g GiteaConn) Users() []string {
	// returns usernames of all the users known
	users := []string{}
	for _, c := range g.config.ReceiveFrom {
		lst := strings.Split( c, "/" )
		if len(lst) == 2 {
			users = append( users, lst[0] )
		}
	}
	return users
}


func(g GiteaConn) computeGitBlobSHA(s string) string {
	p := fmt.Sprintf("blob %d\x00", len(s))
	h := sha1.New()
	h.Write([]byte(p))
	h.Write([]byte(s))
	return hex.EncodeToString( h.Sum([]byte(nil)) )
}
