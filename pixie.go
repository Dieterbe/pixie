package main

import (
	"./backend"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stvp/go-toml-config"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var thumbnail_dir = config.String("thumbnail_dir", "")
var tmsu_file = config.String("tmsu_file", "")

type Photo struct {
	Id    int               `json:"id"`
	Name  string            `json:"name"`
	Ext   string            `json:"-"`
	Dir   string            `json:"dir"`
	Thumb string            `json:"thumb"`
	Tags  []string          `json:"tags"`
	Edits map[string]string `json:"edits"`
}

func (p *Photo) String() string {
	return fmt.Sprintf("Photo {Id: %d, Name: '%s', Dir: '%s', Thumb: '%s'}", p.Id, p.Name, p.Dir, p.Thumb)
}

func (p *Photo) LoadEdits(edits_dir string) error {
	// /bleh/originals/foo/bar.jpg ->
	// /bleh/edits/foo/bar.jpg (standard edit)
	// /bleh/edits/foo/bar-speficier.jpg (specified edit)
	if edits_dir == "" {
		return nil
	}
	basename := p.Name[:len(p.Name)-len(p.Ext)]
	edit_base := path.Join(edits_dir, basename)
	glob_pattern := fmt.Sprintf("%s*%s", edit_base, p.Ext)
	fmt.Println("looking for edits ", glob_pattern)
	out := make(map[string]string)
	matches, err := filepath.Glob(glob_pattern)
	if err != nil {
		return err
	}
	for _, match_path := range matches {
		key := strings.Replace(match_path, edit_base, "", 1)
		key = strings.Replace(key, p.Ext, "", 1)
		if key == "" {
			key = "standard"
		} else {
			key = key[1:] // strip leading '_', '-', etc
		}
		out[key] = match_path
	}
	p.Edits = out
	return nil
}

// [un]tag a given fname with a given tag
func api_photo_handler(w http.ResponseWriter, r *http.Request, conn_sqlite *sql.DB) {
	err := r.ParseForm()
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Invalid request: %s", err)}, 503)
		return
	}
	tag := r.Form.Get("tag")
	untag := r.Form.Get("untag")
	fname := r.Form.Get("fname")
	if fname == "" {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Invalid request: %s", err)}, 503)
		return
	}
	if tag == "" && untag == "" {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Invalid request: %s", err)}, 503)
		return
	}
	if tag != "" && untag != "" {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Invalid request: %s", err)}, 503)
		return
	}
	if tag != "" {
		backend.Tag(w, r, conn_sqlite, fname, tag)
	} else {
		backend.UnTag(w, r, conn_sqlite, fname, untag)
	}
}

func find_edits_dir(dir string) (string, error) {
	edits_dir := strings.Replace(dir, "originals", "edits", 1)
	edits_dir = strings.Replace(edits_dir, "originals-generated", "edits", 1)
	_, err := os.Stat(edits_dir)
	if err != nil {
		return edits_dir, nil
	}
	if os.IsNotExist(err) {
		return "", nil
	}
	return edits_dir, err
}

// get a list of photos (with tags) for a given dir
func api_photos_handler(w http.ResponseWriter, r *http.Request, conn_sqlite *sql.DB) {
	dir := strings.Replace(r.URL.Path, "/api/photos", "", 1)
	fmt.Printf("reading dir '%s'\n", dir)
	list, err := ioutil.ReadDir(dir)
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Cannot read directory: '%s': %s", dir, err)}, 503)
		return
	}
	edits_dir, err := find_edits_dir(dir)
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Edits directory '%s' seems to exist but unable to read: %s", edits_dir, err)}, 503)
		return
	}
	photos := make([]Photo, 0, len(list))
	dir, err = filepath.Abs(dir)
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Cannot figure out directory abspath: '%s': %s", dir, err)}, 503)
		return
	}

	filetags, err := backend.GetFileTags(dir, conn_sqlite)
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Cannot get file tags: '%s': %s", dir, err)}, 503)
		return
	}

	id := 0
	for _, f := range list {
		name := f.Name()
		ext := filepath.Ext(name)
		mime := mime.TypeByExtension(ext)
		if strings.HasPrefix(mime, "image/") {
			h := md5.New()
			io.WriteString(h, fmt.Sprintf("file://%s/%s", dir, name))
			thumb := fmt.Sprintf("%x.png", h.Sum(nil))
			var tags_slice []string
			tags_str, ok := filetags[name]
			if ok {
				fmt.Printf("%s: '%s'\n", name, tags_str)
				tags_slice = strings.Split(tags_str, ",")
			} else {
				tags_slice = make([]string, 0, 0)
			}
			p := Photo{id, name, ext, dir, thumb, tags_slice, nil}
			err = p.LoadEdits(edits_dir)
			if err != nil {
				fmt.Printf("WARNING: failed to load edits: %s\n", err)
			}
			id++
			photos = append(photos, p)
		}
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(photos)
	if err != nil {
		fmt.Printf("WARNING: failed to encode/write json: %s\n", err)
	}
}

func main() {
	addr := ":8080"
	config.Parse("config.ini")
	conn_sqlite, err := sql.Open("sqlite3", *tmsu_file)
	if err != nil {
		log.Fatal("could not open database: ", err.Error())
	}
	http.HandleFunc("/api/photos/", func(w http.ResponseWriter, r *http.Request) {
		api_photos_handler(w, r, conn_sqlite)
	})
	http.HandleFunc("/api/photo", func(w http.ResponseWriter, r *http.Request) {
		api_photo_handler(w, r, conn_sqlite)
	})
	http.Handle("/thumbnails/", http.StripPrefix("/thumbnails/", http.FileServer(http.Dir(*thumbnail_dir))))
	http.Handle("/", http.FileServer(http.Dir(".")))
	fmt.Printf("starting up on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
