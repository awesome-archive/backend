package lib

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"github.com/browsefile/backend/src/errors"
	"github.com/browsefile/backend/src/lib/fileutils"
	"github.com/maruel/natural"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// File contains the information about a particular file or directory.
type File struct {
	// Indicates the Kind of view on the front-end (Listing, editor or preview).
	Kind string `json:"kind"`
	// The name of the file.
	Name string `json:"name"`
	// The Size of the file.
	Size int64 `json:"size"`
	// The absolute URL.
	URL string `json:"url"`
	// The extension of the file.
	Extension string `json:"extension"`
	// The last modified time.
	ModTime time.Time `json:"modified"`
	// The File Mode.
	Mode os.FileMode `json:"mode"`
	// Indicates if this file is a directory.
	IsDir bool `json:"isDir"`
	// Absolute path.
	Path string `json:"path"`
	// Relative path to user's virtual File System.
	VirtualPath string `json:"virtualPath"`
	// Indicates the file content type: video, text, image, music or blob.
	Type string `json:"type"`
	// Stores the content of a text file.
	Content string `json:"content,omitempty"`

	Checksums map[string]string `json:"checksums,omitempty"`
	*Listing  `json:",omitempty"`

	Language string `json:"language,omitempty"`
}

// A Listing is the context used to fill out a template.
type Listing struct {
	// The items (files and folders) in the path.
	Items []*File `json:"items"`
	// The number of directories in the Listing.
	NumDirs int `json:"numDirs"`
	// The number of files (items that aren't directories) in the Listing.
	NumFiles int `json:"numFiles"`
	// Which sorting order is used.
	Sort string `json:"sort"`
	// And which order.
	Order string `json:"order"`
	//indicator to the frontend, to prevent request previews
	AllowGeneratePreview bool `json:"allowGeneratePreview"`
}

// GetInfo gets the file information and, in case of error, returns the
// respective HTTP error code
func GetInfo(url *url.URL, c *Context) (*File, error) {
	var err error
	info, err, path, t := fileutils.GetFileInfo(c.User.Scope, url.Path)

	i := &File{
		URL:         "/files" + url.String(),
		VirtualPath: url.Path,
		Path:        path,
	}

	if err != nil {
		return i, err
	}

	i.Name = info.Name()
	i.ModTime = info.ModTime()
	i.Mode = info.Mode()
	i.IsDir = info.IsDir()
	i.Size = info.Size()
	i.Extension = filepath.Ext(i.Name)
	i.Type = t

	if i.IsDir && !strings.HasSuffix(i.URL, "/") {
		i.URL += "/"
	}

	return i, nil
}

// GetListing gets the information about a specific directory and its files.
func (i *File) GetListing(u *UserModel, isRecursive bool) error {
	// Gets the directory information using the Virtual File System of
	// the user configuration.
	var files []os.FileInfo
	var paths []string
	files = make([]os.FileInfo, 0, 1000)
	paths = make([]string, 0, 1000)
	var (
		fileinfos           []*File
		dirCount, fileCount int
	)
	// Absolute URL
	var fUrl url.URL

	if isRecursive {
		err := filepath.Walk(filepath.Join(u.Scope, i.VirtualPath),
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				files = append(files, info)
				path = strings.Replace(path, u.Scope, "", -1)
				paths = append(paths, path)

				return nil
			})
		if err != nil {
			log.Println(err)
		}
	} else {
		f, err := u.FileSystem.OpenFile(i.VirtualPath, os.O_RDONLY, 0)
		if err != nil {
			return err
		}
		defer f.Close()
		// Reads the directory and gets the information about the files.
		files, err = f.Readdir(-1)
		if err != nil {
			return err
		}
	}

	baseurl, err := url.PathUnescape(i.URL)
	if err != nil {
		return err
	}
	for ind, f := range files {
		name := f.Name()

		if strings.HasPrefix(f.Mode().String(), "L") {
			// It's a symbolic link. We try to follow it. If it doesn't work,
			// we stay with the link information instead if the target's.
			info, err := os.Stat(f.Name())
			if err == nil {
				f = info
			}
		}

		if f.IsDir() {
			name += "/"
			dirCount++
		} else {
			fileCount++
		}
		var vPath, path string
		if isRecursive {
			path = filepath.Join(i.Path, paths[ind])
			vPath = filepath.Dir(paths[ind])
			if f.IsDir() {
				fUrl = url.URL{Path: baseurl}
			} else {
				fUrl = url.URL{Path: "/files" + paths[ind]}
			}

		} else {
			path = filepath.Join(i.Path, name)
			vPath = filepath.Join(i.VirtualPath, name)
			fUrl = url.URL{Path: baseurl + name}
		}

		i := &File{
			Name:        f.Name(),
			Size:        f.Size(),
			ModTime:     f.ModTime(),
			Mode:        f.Mode(),
			IsDir:       f.IsDir(),
			URL:         fUrl.String(),
			Extension:   filepath.Ext(name),
			Path:        path,
			VirtualPath: vPath,
		}

		i.SetFileType(false)
		fileinfos = append(fileinfos, i)
	}

	i.Listing = &Listing{
		Items:    fileinfos,
		NumDirs:  dirCount,
		NumFiles: fileCount,
	}

	return nil
}

// SetFileType obtains the mimetype and converts it to a simple
// type nomenclature.
func (f *File) SetFileType(checkContent bool) error {
	if len(f.Type) > 0 || f.IsDir {
		return nil
	}
	var content []byte
	var err error
	isOk, mimetype := fileutils.GetBasedOnExtensions(f.Extension)
	// Tries to get the file mimetype using its extension.
	if !isOk && checkContent {
		return nil
		log.Println("Can't detect file type, based on extension ", f.Name)
		/*content, mimetype, err = fileutils.GetBasedOnContent(f.Path)
		if err != nil {
			return err
		}*/
	}

	f.Type = mimetype

	// If the file type is text, save its content.
	if f.Type == "text" {
		if len(content) == 0 {
			//todo: fix me, what if file too big ?
			content, err = ioutil.ReadFile(f.Path)
			if err != nil {
				return err
			}
		}

		f.Content = string(content)
	}

	return nil
}

// Checksum retrieves the checksum of a file.
func (i *File) Checksum(algo string) error {
	if i.IsDir {
		return errors.ErrIsDirectory
	}

	if i.Checksums == nil {
		i.Checksums = make(map[string]string)
	}

	file, err := os.Open(i.Path)
	if err != nil {
		return err
	}

	defer file.Close()

	var h hash.Hash

	switch algo {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		return ErrInvalidOption
	}

	_, err = io.Copy(h, file)
	if err != nil {
		return err
	}

	i.Checksums[algo] = hex.EncodeToString(h.Sum(nil))
	return nil
}

// CanBeEdited checks if the extension of a file is supported by the editor
func (i File) CanBeEdited() bool {
	return i.Type == "text"
}

// ApplySort applies the sort order using .Order and .Sort
func (l Listing) ApplySort() {
	// Check '.Order' to know how to sort
	if l.Order == "desc" {
		switch l.Sort {
		case "name":
			sort.Sort(sort.Reverse(byName(l)))
		case "size":
			sort.Sort(sort.Reverse(bySize(l)))
		case "modified":
			sort.Sort(sort.Reverse(byModified(l)))
		default:
			// If not one of the above, do nothing
			return
		}
	} else { // If we had more Orderings we could add them here
		switch l.Sort {
		case "name":
			sort.Sort(byName(l))
		case "size":
			sort.Sort(bySize(l))
		case "modified":
			sort.Sort(byModified(l))
		default:
			sort.Sort(byName(l))
			return
		}
	}
}

// Implement sorting for Listing
type byName Listing
type bySize Listing
type byModified Listing

// By Name
func (l byName) Len() int {
	return len(l.Items)
}

func (l byName) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}

// Treat upper and lower case equally
func (l byName) Less(i, j int) bool {
	if l.Items[i].IsDir && !l.Items[j].IsDir {
		return true
	}

	if !l.Items[i].IsDir && l.Items[j].IsDir {
		return false
	}

	return natural.Less(strings.ToLower(l.Items[j].Name), strings.ToLower(l.Items[i].Name))
}

// By Size
func (l bySize) Len() int {
	return len(l.Items)
}

func (l bySize) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}

const directoryOffset = -1 << 31 // = math.MinInt32
func (l bySize) Less(i, j int) bool {
	iSize, jSize := l.Items[i].Size, l.Items[j].Size
	if l.Items[i].IsDir {
		iSize = directoryOffset + iSize
	}
	if l.Items[j].IsDir {
		jSize = directoryOffset + jSize
	}
	return iSize < jSize
}

// By Modified
func (l byModified) Len() int {
	return len(l.Items)
}

func (l byModified) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}

func (l byModified) Less(i, j int) bool {
	iModified, jModified := l.Items[i].ModTime, l.Items[j].ModTime
	return iModified.Sub(jModified) < 0
}
