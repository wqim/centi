package huggingface
import (
)

/*
 * auxilary functions for parsing server's response.
 */
type HFRepo struct {
	RepoID		string			`json:"repo_id"`	// repo_id
	RepoType	string			`json:"type"`
	Name		string			`json:"name"`
	Private		bool			`json:"private"`
}

type HFFile struct {
	Type		string			`json:"type"`
	Oid		string			`json:"oid"`
	Size		uint			`json:"size"`
	Path		string			`json:"path"`
}

type UFile struct {
	Path		string			`json:"path"`
	Sample		string			`json:"sample"`
	Size		int			`json:"size"`
}

type UFiles struct {
	Files		[]UFile			`json:"files"`
}

type UFileResp struct {
	Path		string			`json:"path"`
	ShouldIgnore	bool			`json:"shouldIgnore"`
	UploadMode	string			`json:"uploadMode"`
}

type UFileResponse struct {
	CommitOID	string			`json:"commitOid"`
	Files		[]UFileResp		`json:"files"`
}

type KeyValue struct {
	Key		string			`json:"key"`
	Value		map[string]string	`json:"value"`
}

// other constants and structures
const (
	FileKey = "name"
	RepoKey = "channel"
	RepoTypeKey = "channel-type"
	BranchKey = "branch"
	PrivateKey = "private"	// a key for map, not for crypto part...
)

type File struct {
	Name	string
	Sha	string
	Content	[]byte
}

type Repository struct {
	Name		string
	RepoType	string
	Private		bool
	Files		[]File
}
