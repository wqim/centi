package github
import (

)

/*
 * auxilary structures for json requests/responses
 */
type GHRepo struct {
	Name		string			`json:"name"`
	Description	string			`json:"description"`
	Private		bool			`json:"private"`
	HasIssues	bool			`json:"has_issues"`
	HasProjects	bool			`json:"has_projects"`
	HasWiki		bool			`json:"has_wiki"`
	AutoInit	bool			`json:"auto_init"`	// true
}

/*
 * structures for getting list of files.
 * instead of sending request for recursively getting list of files, we
 * are sending only 2 requests: one for getting sha of repository and one
 * for recursive listing.
 */
type GHObject struct {
	Sha		string			`json:"sha"`
	Type		string			`json:"type"`
	URL		string			`json:"url"`
}

type GHBranch struct {
	Ref		string			`json:"ref"`
	NodeID		string			`json:"node_id"`
	URL		string			`json:"url"`
	Object		GHObject		`json:"object"`
}

/*
 * structure of file in github repository
 */
type GHFile struct {
	Path		string			`json:"path"`
	Mode		string			`json:"mode"`
	Type		string			`json:"type"`
	Sha		string			`json:"sha"`
	Size		int			`json:"size"`
	URL		string			`json:"url"`
}

/* structure, representing files in github repo */
type GHList struct {
	Sha		string			`json:"sha"`
	URL		string			`json:"url"`
	Tree		[]GHFile		`json:"tree"`
	Truncated	bool			`json:"truncated"`
}

/* structure for parsing data about file */
type GHFileContent struct {
	Name		string			`json:"name"`
	Path		string			`json:"path"`
	Sha		string			`json:"sha"`
	Size		int			`json:"size"`
	URL		string			`json:"url"`
	HtmlUrl		string			`json:"html_url"`
	GitUrl		string			`json:"git_url"`
	DownloadUrl	string			`json:"download_url"`
	Type		string			`json:"type"`
	Content		string			`json:"content"`	// the contents of the file
	Encoding	string			`json:"encoding"`	// base64 by default
	Links		map[string]string	`json:"_links"`
}
