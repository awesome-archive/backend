package web

import "C"
import (
	fb "github.com/browsefile/backend/src/lib"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// set router, and all other params to the context, returns true in case request are about shares
func processParams(c *fb.Context, r *http.Request) (isShares bool) {

	if c.Query == nil {
		c.Query = r.URL.Query()
	}
	c.Sort = c.Query.Get("sort")
	c.Order = c.Query.Get("order")
	c.PreviewType = c.Query.Get("previewType")
	c.Inline, _ = strconv.ParseBool(c.Query.Get("inline"))
	c.RootHash, _ = url.QueryUnescape(c.Query.Get("rootHash"))
	c.Checksum = c.Query.Get("checksum")
	c.ShareType = c.Query.Get("share")
	c.IsRecursive, _ = strconv.ParseBool(c.Query.Get("recursive"))
	c.Override, _ = strconv.ParseBool(c.Query.Get("override"))
	c.Algo = c.Query.Get("algo")

	//search request
	q := c.Query.Get("query")
	if len(q) > 0 {
		if strings.Contains(q, "type") {
			arr := strings.Split(q, ":")
			arr = strings.Split(arr[1], " ")
			c.SearchString = arr[1]
			c.SearchType = arr[0]
		} else {
			c.SearchString = q
		}
		//c.Query.Del("query")
	}
	if len(c.Algo) > 0 && !strings.HasPrefix(c.Algo, "z") {
		arr := strings.Split(c.Algo, "_")
		c.Algo = arr[0]
		if len(arr) > 1 {
			c.Image = strings.Contains(arr[1], "i")
			c.Video = strings.Contains(arr[1], "v")
			c.Audio = strings.Contains(arr[1], "a")
		}
	}

	r.URL.RawQuery = ""

	isShares = setRouter(c, r)
	//in case download, might be multiple files
	if strings.HasPrefix(c.Router, "dow") {
		f := c.Query.Get("files")
		if len(f) > 0 {
			c.FilePaths = strings.Split(f, ",")
		}

	} else if r.Method == http.MethodPatch {
		c.Destination = r.Header.Get("Destination")
		if len(c.Destination) == 0 {
			c.Destination = r.URL.Query().Get("destination")
		}
		c.Action = r.Header.Get("action")
		if len(c.Action) == 0 {
			c.Action = r.URL.Query().Get("action")
		}
	}

	return

}

func setRouter(c *fb.Context, r *http.Request) (isShares bool) {
	c.Router, r.URL.Path = splitURL(r.URL.Path)
	isShares = strings.HasPrefix(c.Router, "shares")

	//redirect to the real handler in shares case
	if isShares {
		//possibility to process shares view/download
		if !(strings.EqualFold(c.ShareType, "my-list") ||
			strings.EqualFold(c.ShareType, "my") ||
			strings.EqualFold(c.ShareType, "list") ||
			strings.EqualFold(c.ShareType, "gen-ex")) {
			c.Router, r.URL.Path = splitURL(r.URL.Path)
		}

		if c.Router == "download" {
			c.Router = "download-share"
		} else if c.Router == "resource" || c.Router == "external" {
			c.Router = "shares"
		}
	}

	return
}

// splitURL splits the path and returns everything that stands
// before the first slash and everything that goes after.
func splitURL(path string) (string, string) {
	if path == "" {
		return "", ""
	}

	path = strings.TrimPrefix(path, "/")

	i := strings.Index(path, "/")
	if i == -1 {
		return "", path
	}

	return path[0:i], path[i:]
}
