package web

import (
	"encoding/json"
	"github.com/browsefile/backend/src/cnst"
	"github.com/browsefile/backend/src/config"
	fb "github.com/browsefile/backend/src/lib"
	"net/http"
)

func settingsHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path != "" && r.URL.Path != "/" {
		return http.StatusNotFound, nil
	}

	switch r.Method {
	case http.MethodGet:
		return settingsGetHandler(c, w)
	case http.MethodPut:
		return settingsPutHandler(c, r)
	}

	return http.StatusMethodNotAllowed, nil
}

func settingsGetHandler(c *fb.Context, w http.ResponseWriter) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}
	return renderJSON(w, c.Config.CopyConfig())
}

func settingsPutHandler(c *fb.Context, r *http.Request) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}

	// Checks if the request body is empty.
	if r.Body == nil {
		return http.StatusForbidden, cnst.ErrEmptyRequest
	}

	// Parses the request body and checks if it's well formed.
	mod := &config.GlobalConfig{}
	err := json.NewDecoder(r.Body).Decode(mod)
	if err != nil {
		return http.StatusBadRequest, err
	}
	mod.Init()
	c.Config.UpdateConfig(mod)
	c.Config.RefreshUserRam()

	return http.StatusOK, nil
}
