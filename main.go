package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/stvp/go-toml-config"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

var thumbnail_dir = config.String("thumbnail_dir", "")

type Photo struct {
	Name  string `json:"name"`
	Dir   string `json:"dir"`
	Thumb string `json:"thumb"`
}

func (p *Photo) String() string {
	return fmt.Sprintf("Photo {Name: '%s', Dir: '%s', Thumb: '%s'}", p.Name, p.Dir, p.Thumb)
}

func api_handler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api/photos/dir=") {
		return
	}
	dir := strings.Replace(r.URL.Path, "/api/photos/dir=", "", 1)
	fmt.Printf("reading dir '%s'\n", dir)
	list, err := ioutil.ReadDir(dir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot read directory: '%s': %s", dir, err), 503)
	}
	photos := make([]Photo, 0, len(list))
	dir, err = filepath.Abs(dir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot figure out directory abspath: '%s': %s", dir, err), 503)
	}
	for _, f := range list {
		name := f.Name()
		ext := filepath.Ext(name)
		mime := mime.TypeByExtension(ext)
		if strings.HasPrefix(mime, "image/") {
			h := md5.New()
			io.WriteString(h, fmt.Sprintf("file://%s/%s", dir, name))
			thumb := fmt.Sprintf("%x.png", h.Sum(nil))
			p := Photo{name, dir, thumb}
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

func main() {
	addr := ":8080"
	config.Parse("config.ini")
	http.HandleFunc("/api/", api_handler)
	http.Handle("/thumbnails/", http.StripPrefix("/thumbnails/", http.FileServer(http.Dir(*thumbnail_dir))))
	http.Handle("/", http.FileServer(http.Dir(".")))
	fmt.Printf("starting up on %s\n", addr)
	http.ListenAndServe(addr, nil)
}
