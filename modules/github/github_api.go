package github
import (
	"fmt"
	"encoding/hex"
	"encoding/json"
	"encoding/base64"
	"io/ioutil"
	"strings"
	"net/http"
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

func(g GitHubConn) DownloadFile( repoName string, branch string, filepath string ) ([]byte, error) {
	formattedUrl := "https://raw.githubusercontent.com/" + repoName + "/refs/heads/" + branch + "/" + filepath
	fmt.Println("[GitHubConn] Downloading file from", formattedUrl)
	resp, err := http.Get( formattedUrl )
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll( resp.Body )
	return data, err
}

func(g GitHubConn) UploadFile( repoName, branch, filepath, sha string, data []byte ) error {
	var err error
	content := base64.StdEncoding.EncodeToString( data )
	if sha == "" {
		sha, err = g.getFileSha( repoName, filepath )
	}
	jsonData := map[string]string{
		"message": "Update " + filepath,
		"content": content,
		"branch": branch,
	}
	if sha != "" {
		jsonData["sha"] = sha
	}
	packed, err := json.Marshal( jsonData )
	if err != nil {
		return err
	}
	args := map[string]string{
		FileKey: filepath,
		RepoKey: repoName,
	}
	url := g.formatURL( general.SendKey, args )
	_, err = g.sendRequest( url, general.SendKey, packed )
	return err
}

func(g GitHubConn) ListRepo( repoName string, branch string ) ([]File, string, error) {
	files := []File{}
	repoSha, err := g.getRepoSha( repoName )
	if err != nil {
		return nil, "", err
	}
	args := map[string]string{
		ShaKey: repoSha,
		RepoKey: repoName,
	}
	url := g.formatURL( general.ListKey, args )
	res, err := g.sendRequest( url, general.ListKey, nil )
	if err != nil {
		return nil, "", err
	}

	// res - GHList structure
	var list GHList
	if err = json.Unmarshal( res, &list ); err != nil {
		return nil, "", err
	}
	fmt.Println("Repository files:")
	for _, ghFile := range list.Tree {
		if ghFile.Type == "blob" {	// an actual file
			fmt.Println("\t+", ghFile.Path)
			files = append(files, File{ ghFile.Sha, ghFile.Path, nil })
		}
	}
	return files, repoSha, nil
}


func(g GitHubConn) GetNewFiles( repo Repository, allFiles []File ) []File {
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

func(g GitHubConn) sendRequest( url, action string, data []byte ) ([]byte, error) {
	/*headers := map[string]string{}
	if (action == general.ListKey) && (g.config.Authed == true) {
		headers = g.headers
	}*/
	fmt.Println("[GitHubConn::sendRequest]: ", g.methods[action], url, string(data) )
	resp, err := general.HTTPRequest( url, g.methods[ action ], data, g.headers )
	if err != nil {
		return nil, err
	}
	
	//fmt.Println("Create channel error:")
	//fmt.Println( string(resp) )

	if err = g.isError( resp ); err != nil {
		return nil, err
	}
	return resp, nil
}

func(g GitHubConn) formatURL( action string, args map[string]string ) string {
	// format URL before the request
	url := strings.Replace( g.endpoints[ action ], "%s", args[ RepoKey ], 1 )
	switch action {
	case general.ListKey:
		if args != nil {
			if util.MapContains(args, ShaKey) {
				url = strings.Replace( url, "%s", args[ ShaKey ], 1 )
			} else {
				sha, err := g.getRepoSha( args[ RepoKey ] )
				if err != nil {
					panic("Failed to get repository sha: " + err.Error())
				} else {
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

func(g GitHubConn) parseData( data []byte ) (*GHFileContent, error) {
	var fc GHFileContent
	if err := json.Unmarshal( data, &fc ); err != nil {
		return nil, err
	}

	if fc.Encoding == "base64" {
		_, err := base64.StdEncoding.DecodeString( fc.Content )
		if err != nil {
			return nil, err
		}
		return &fc, nil
	}
	return nil, fmt.Errorf("[parseData] unknown encoding")
}

func(g GitHubConn) getRepoSha( channel string ) (string, error) {
	url := g.baseUrl + "/repos/" + channel + "/git/refs/heads/main"
	res, err := general.HTTPRequest( url, g.methods[ general.RecvKey ], nil, g.headers )
	if err != nil {
		fmt.Println("Failed to get repository sha:")
		fmt.Println( string(res) )
		return "", err
	}
	var fc GHBranch
	if err := json.Unmarshal( res, &fc ); err != nil {
		return "", err
	}
	return fc.Object.Sha, nil
}

func(g GitHubConn) getFileSha( channel string, filename string ) (string, error) {
	// list file in the repository
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

func(g GitHubConn) parseListing( data []byte ) ([][]string, error) {

	// function which parses the listing of files in the repository
	// returns list of []string containing filename at 0 index and sha at 1 index
	var filesTree GHList
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

func(g GitHubConn) isError( resp []byte ) error {
	// function which checks if the server response contains
	// an error (hit rate limits, token expired, etc.)
	var tmp map[string]any
	if err := json.Unmarshal( resp, &tmp ); err != nil {
		if strings.Contains( strings.ToLower( string(resp)), "rate limit" ) {
			return fmt.Errorf("rate limit")
		}
		return nil
	}
	for k, v := range tmp {
		if k == "error" {
			val, ok := v.(string)
			if ok {
				return fmt.Errorf( val )
			}
		}
		if k == "message" {
			val, ok := v.(string)
			if ok {
				if strings.HasPrefix( val, "API rate limit exceeded for user ID" ) {
					return fmt.Errorf("RateLimitError")
				}
				if val == "404 Not Found" {
					return fmt.Errorf( "NotFoundError" )
				}
				return fmt.Errorf( val )
			}
		}
	}
	return nil
}

func(g GitHubConn) computeGitBlobSHA(s string) string {
	p := fmt.Sprintf("blob %d\x00", len(s))
	h := sha1.New()
	h.Write([]byte(p))
	h.Write([]byte(s))
	return hex.EncodeToString( h.Sum([]byte(nil)) )
}
