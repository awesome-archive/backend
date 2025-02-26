package lib

import (
	"github.com/browsefile/backend/src/cnst"
	"github.com/browsefile/backend/src/config"
	"github.com/browsefile/backend/src/lib/utils"
	"net/http"
	"net/url"
	"strings"
)

//used to filter out specific files by name and path
type FitFilter func(name, p string) bool

// Context contains the needed information to make handlers work.
type Context struct {
	*FileBrowser
	User *UserModel
	File *File
	// On API handlers, Router is the APi handler we want.
	Router int
	*Params
	FitFilter
}

//params in URL request
type Params struct {
	//indicate that requested preview
	PreviewType string
	//return files list by recursion
	IsRecursive bool
	//indicate about share request, nformation about share metadata like list, my-meta.
	ShareType    string
	SearchString string
	/*
		external share item root dir hash,
		note: as a result, cut owner from files paths, and replaces with rootHash(hash value of specific shareItem)
	*/
	RootHash string
	//download type, zip or playlist m3u8
	Algo string
	//download multiple files
	FilePaths []string

	Auth string

	Checksum string

	Inline bool
	// playlist & search file mime types, true if any was specified at request url, uses in FitFilter type
	Audio bool
	Image bool
	Video bool
	Pdf   bool
	//search query
	Query url.Values
	//override existing file
	Override bool
	// used in resource patch requests type
	Destination string
	Action      string

	Sort  string
	Order string
	//is share request
	IsShare bool
	//requestURL
	URL    string
	Method string
	RESP   http.ResponseWriter
	REQ    *http.Request
}

//cut user home path from for non download routes
func (c *Context) CutPath(path string) string {
	if c.Router != cnst.R_DOWNLOAD {
		if c.IsShare {
			if c.IsExternalShare() {
				path = strings.TrimPrefix(path, c.GetUserHomePath())
				path = "/" + strings.SplitN(path, "/", 3)[2]
			} else if c.Router == cnst.R_PLAYLIST {
				path = strings.TrimPrefix(path, c.GetUserSharesPath())
			} else {
				path = strings.TrimPrefix(path, c.GetUserSharesPath())
			}
		} else {
			path = strings.TrimPrefix(path, c.GetUserHomePath())
		}
	}
	return path

}
func (c *Context) GetUserHomePath() string {
	return c.Config.GetUserHomePath(c.User.Username)
}
func (c *Context) GetUserPreviewPath() string {
	return c.Config.GetUserPreviewPath(c.User.Username)
}
func (c *Context) GetUserSharesPath() string {
	return c.Config.GetUserSharesPath(c.User.Username)
}

func (c *Context) GenPreview(out string) {
	if len(c.Config.ScriptPath) > 0 {
		_, t := utils.GetBasedOnExtensions(c.File.Name)
		if t == cnst.IMAGE || t == cnst.VIDEO {
			c.Pgen.Process(c.Pgen.GetDefaultData(c.File.Path, out, t))
		}
	}
}

func (c *Context) GenSharesPreview(out string) {
	if len(c.Config.ScriptPath) > 0 {

		_, t := utils.GetBasedOnExtensions(c.File.Name)
		if t == cnst.IMAGE || t == cnst.VIDEO {
			c.Pgen.Process(c.Pgen.GetDefaultData(c.File.Path, out, t))

		}

	}
}

//true if request contains rootHash param
func (c *Context) IsExternalShare() (r bool) {
	return len(c.RootHash) > 0
}
func (c *Context) GetAuthConfig() *config.ListenConf {
	isTls := c.REQ.TLS != nil
	var cfgM *config.ListenConf
	if isTls {
		cfgM = c.Config.Tls
	} else {
		cfgM = c.Config.Http

	}
	return cfgM
}
