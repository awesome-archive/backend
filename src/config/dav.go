package config

import (
	"github.com/browsefile/backend/src/cnst"
	"os"
	"path/filepath"
)

func addDavShare(shr *ShareItem, own string) {
	if shr.AllowLocal {
		for _, u := range config.Users {
			config.checkSymLinkPath(shr, u.Username, own)

		}
	} else if len(shr.AllowUsers) > 0 {
		for _, uName := range shr.AllowUsers {
			u, _ := usersRam[uName]
			config.checkSymLinkPath(shr, u.Username, own)
		}

	}
}
func delDavShare(shr *ShareItem, user string) {
	dPath := filepath.Join(config.GetDavPath(user), cnst.WEB_DAV_FOLDER, "shares", shr.Path)
	os.RemoveAll(dPath)
}
