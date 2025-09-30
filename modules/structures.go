package gitea
import (

)

/*
 * auxilary structures for json requests/responses
 *
 * basically, the structures are the same as for github api.
 * but i'll copy-paste and rename them in case if gitea api format
 * will change someday.
 */
type GTLinks struct {
	Self		string		`json:"self"`
	Git		string		`json:"git"`
	Html		string		`json:"html"`
}

type GTObject struct {
	Name		string		`json:"name"`
	Path		string		`json:"path"`
	Sha		string		`json:"sha"`
	LastCommitSha	string		`json:"last_commit_sha"`
	Type		string		`json:"type"`
	Size		int		`json:"size"`
	Encoding	string		`json:"encoding"`
	Content		string		`json:"content"`
	Target		string		`json:"target"`
	URL		string		`json:"url"`
	HtmlURL		string		`json:"html_url"`
	GitURL		string		`json:"git_url"`
	DownloadURL	string		`json:"download_url"`
	SubmoduleGitURL	string		`json:"submodule_git_url"`
	Links		GTLinks		`json:"_links"`
}

type GTObj struct {
	Type		string		`json:"type"`
	Sha		string		`json:"sha"`
	URL		string		`json:"url"`
}

type GTBranch struct {
	Ref		string			`json:"ref"`
	URL		string			`json:"url"`
	Object		GTObj			`json:"object"`
}

type GTFile struct {
	Path		string			`json:"path"`
	Mode		string			`json:"mode"`
	Type		string			`json:"type"`
	Sha		string			`json:"sha"`
	Size		int			`json:"size"`
	URL		string			`json:"url"`
}

type GTList struct {
	Sha		string			`json:"sha"`
	URL		string			`json:"url"`
	Tree		[]GTFile		`json:"tree"`
	Truncated	bool			`json:"truncated"`
}
