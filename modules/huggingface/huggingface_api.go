package huggingface
import (
	"fmt"
	"strings"
	"encoding/json"
	"encoding/base64"
	"io/ioutil"
	"net/http"

	"centi/util"
	"centi/modules/general"
)

func(hf HuggingfaceConn) DownloadFile( repo Repository, filename, branch string ) ([]byte, error) {
	url := "https://huggingface.co/" + repo.RepoType + "s/" + repo.Name + "/resolve/" + branch + "/" + filename

	util.DebugPrintln("[HuggingfaceConn]: Downloading file from ", url )
	
	resp, err := http.Get( url )
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll( resp.Body )
	if err != nil {
		return nil, err
	}
	return data, nil
}

func(hf HuggingfaceConn) UploadFile( repo Repository, filename string, data []byte ) error {
	contents64 := base64.StdEncoding.EncodeToString( data )
	ufiles := UFiles{
		[]UFile{
			UFile{
				filename,
				contents64,
				len( data ),
			},
		},
	}

	packed, err := json.Marshal( ufiles )
	if err != nil {
		return err
	}
	url := hf.formatURL( repo, filename, general.SendKey )
	response, err := hf.sendRequest( url, general.SendKey, packed )
	if err != nil {
		return err
	}
	util.DebugPrintln("Upload file response:")
	util.DebugPrintln( string(response) )
	return hf.Commit( repo, filename, data )
	/*_, err = general.HTTPRequest( "https://huggingface.co/api/" + repo.RepoType + "s/" + repo.Name + "/revision/" + hf.config.SendBranch + "?expand=xetEnabled", "GET", nil, hf.headers )
	if err != nil {
		return err
	}

	_, err = general.HTTPRequest( "https://huggingface.co/api/" + repo.RepoType + "s/" + repo.Name + "/revision/" + hf.config.SendBranch, "GET", nil, hf.headers )
	return err */
}

func(hf HuggingfaceConn) ListRepo( repo Repository ) ([]File, error) {
	util.DebugPrintln("Files:")
	files := []File{}
	url := hf.formatURL( repo, "", general.ListKey )
	resp, err := hf.sendRequest( url, general.ListKey, nil )
	if err != nil {
		return nil, err
	}
	var hfiles []HFFile
	if err = json.Unmarshal( resp, &hfiles ); err != nil {
		util.DebugPrintln("This is not HFFile list:")
		util.DebugPrintln( string(resp) )
		return nil, err
	}
	for _, f := range hfiles {
		util.DebugPrintln("\t-", f.Path)
		files = append( files, File{ f.Path, f.Oid, []byte{} } )
	}
	return files, nil
}

func(hf HuggingfaceConn) Commit( repo Repository, filename string, data []byte ) error {
	contents64 := base64.StdEncoding.EncodeToString( data )
	kv := KeyValue{
		"header",
		map[string]string{
			"summary": "Upload " + filename + " with huggingface_hub",
			"description": "",
		},
	}
	packed, err := json.Marshal( kv )
	if err != nil {
		return err
	}
	kv = KeyValue{
		"file",
		map[string]string{
			"content": contents64,
			"path": filename,
			"encoding": "base64",
		},
	}
	tmp, err := json.Marshal( kv )
	if err != nil {
		return err
	}
	packed = append( packed, 0x0a )
	packed = append( packed, tmp... )
	packed = append( packed, 0x0a )
	// packed - the data to post
	hf.headers["Content-Type"] = "application/x-ndjson"
	url := hf.formatURL( repo, filename, general.CommitKey )
	response, err := hf.sendRequest( url, general.CommitKey, packed )
	util.DebugPrintln("Commit response:")
	util.DebugPrintln( string(response) )
	hf.headers["Content-Type"] = "application/json"
	return err
}

func(hf HuggingfaceConn) GetNewFiles( repo Repository, allFiles []File ) []File {
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

func(hf HuggingfaceConn) sendRequest( url string, action string, data []byte ) ([]byte, error) {

	util.DebugPrintln("[HuggingfaceConn]:", hf.methods[ action ], url, string(data) )
	resp, err := general.HTTPRequest( url, hf.methods[ action ], data, hf.headers )
	if err != nil {
		return nil, err
	}
	if err = hf.isError( resp ); err != nil {
		return nil, err
	}
	return resp, nil
}

func(hf HuggingfaceConn) formatURL( repo Repository, filepath, action string ) string {
	url := hf.baseUrl + strings.Replace(
		strings.Replace(
			strings.Replace(hf.endpoints[action], "%s", repo.RepoType + "s", 1),
		"%s", repo.Name, 1 ),
		"%s", filepath, 1,
	)
	return url
}

// returns []string{ listOfDirectories, listOfAllFiles }
func(hf HuggingfaceConn) parseListing( data []byte ) ([][]string, error) {
	var files []HFFile
	if err := json.Unmarshal( data, &files ); err != nil {
		return nil, err
	}
	dirs := []string{}
	totalFiles := []string{}
	for _, f := range files {
		if f.Type == "file" {
			totalFiles = append( totalFiles, f.Path )
		} else if f.Type == "directory" {
			dirs = append( dirs, f.Path )
		}
	}

	return [][]string{dirs, totalFiles}, nil
}

func(hf HuggingfaceConn) isError( data []byte ) error {

	var tmp map[string]any
	if err := json.Unmarshal( data, &tmp ); err != nil {
		/* plaintext data or html error */
		if strings.Contains( string(data), "<p>We had to rate limit you. If you think it's an error, upgrade to a paid <a href=" ) {
			return fmt.Errorf( "RateLimitError" )
		}
		return nil
	}
	
	for k, v := range tmp {
		if k == "error" {
			val, ok := v.(string)
			if ok {	// must always work
				return fmt.Errorf( val )
			}
		}
	}
	return nil
}

func(hf HuggingfaceConn) findRepoByName( repoName string ) (Repository, error){
	for _, rep := range hf.repos {
		if rep.Name == repoName {
			return rep, nil
		}
	}
	return Repository{}, fmt.Errorf("Not found")
}
