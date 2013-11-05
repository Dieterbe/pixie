package main

import (
	"io/ioutil"
	"net/http"
	"encoding/json"
	"io"
	"fmt"
	"strings"
	"mime"
	"path/filepath"
	"crypto/md5"
)

type Photo struct {
	Name string `json:"name"`
	AbsPath string `json:"abspath"`
	PathMd5 string `json:"path_md5"`
}

func (p *Photo) String() string {
	return fmt.Sprintf("Photo {Name: '%s', AbsPath: '%s', PathMd5: '%s'}", p.Name, p.AbsPath, p.PathMd5)
}

func api_handler (w http.ResponseWriter, r *http.Request) {
	if ! strings.HasPrefix(r.URL.Path, "/api/photos/dir=") {
		return
	}
	dir := strings.Replace(r.URL.Path, "/api/photos/dir=", "", 1)
	fmt.Printf("reading dir '%s'\n", dir)
	list, err := ioutil.ReadDir(dir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot read directory: '%s': %s", dir, err), 503)
	}
	photos := make([]Photo, 0, len(list))
	abspath, err := filepath.Abs(dir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot figure out directory abspath: '%s': %s", dir, err), 503)
	}
	for _, f := range list {
		name := f.Name()
		ext := filepath.Ext(name)
		mime := mime.TypeByExtension(ext)
		if strings.HasPrefix(mime, "image/") {
			h := md5.New()
			io.WriteString(h, fmt.Sprintf("file://%s/%s", abspath, name))
			pathmd5 := fmt.Sprintf("%x", h.Sum(nil))
			p := Photo{name, abspath, pathmd5}
			photos = append(photos, p)
		}
	}
	fmt.Printf("%x\n", photos)
	enc := json.NewEncoder(w)
	err = enc.Encode(photos)
	if err != nil {
		fmt.Printf("WARNING: failed to encode/write json: %s\n", err)
	}
}

func main () {
	addr := ":8080"
	http.HandleFunc("/api/", api_handler)
	http.Handle("/", http.FileServer(http.Dir(".")))
	fmt.Printf("starting up on %s\n", addr)
	http.ListenAndServe(addr, nil)
}
