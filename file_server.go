package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
)

func JsonFileServer(root http.FileSystem, pathPrefix string) http.Handler {
	return &jsonFileHandler{root, pathPrefix}
}

type jsonFileHandler struct {
	root http.FileSystem
	pathPrefix string
}

type fileContentResponse struct {
	Content string `json:"content"`
}

type dirContentResponse struct {
	Files []string `json:"files"`
}

func (f jsonFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
	}
	if !strings.HasPrefix(upath, f.pathPrefix) {
		msg, code := toHTTPError(os.ErrNotExist)
		http.Error(w, msg, code)
		return
	}
	upath = upath[len(f.pathPrefix):] // TODO ?
	log.Printf("server file %s", upath)
	serveFile(w, r, f.root, path.Clean(upath))
}

func serveFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem, name string) {

	f, err := fs.Open(name)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Still a directory? (we didn't find an index.html file)
	if d.IsDir() {
		serverDirContent(w, r, f)
		return
	}

	serveFileContent(w, r, d.Name(), f)
}

func serverDirContent(w http.ResponseWriter, r *http.Request, f http.File) {
	dirs, err := f.Readdir(-1)
	if err != nil {
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })

	files := make([]string, 0)
	for _, d := range dirs {
		name := d.Name()
		if d.IsDir() {
			name += "/"
		}
		if len(name) <= 0 {
			continue
		}
		files = append(files, name)
	}

	resp := dirContentResponse { files }
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func serveFileContent(w http.ResponseWriter, r *http.Request, name string, content io.ReadSeeker) {
	if !fileTypeSupport(name) {
		http.Error(w, "File type not support", http.StatusBadRequest)
		return
	}

	contentStr, err := ioutil.ReadAll(content)

	if err != nil {
		http.Error(w, "File content err", http.StatusBadRequest)
		return
	}
	resp := fileContentResponse { Content: string(contentStr) }

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func fileTypeSupport(name string) bool {
	if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
		return true
	}
	if strings.HasSuffix(name, ".json") {
		return true
	}
	return false
}

func toHTTPError(err error) (msg string, httpStatus int) {
	if os.IsNotExist(err) {
		return "404 page not found", http.StatusNotFound
	}
	if os.IsPermission(err) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}