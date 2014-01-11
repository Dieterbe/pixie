package main

import (
	"./backend"
	"./cp"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Dieterbe/gothum/workers"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stvp/go-toml-config"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

var ptr_thumbnail_dir = config.String("thumbnail_dir", "")
var ptr_exports_dir = config.String("exports_dir", "") // only used by export script, but go-toml-config insists on declaring all vars
var ptr_tmsu_file = config.String("tmsu_file", "")
var ptr_editor = config.String("editor", "gimp")
var ptr_addr = config.String("addr", "localhost:8080")

var thumbnail_dir, tmsu_file, editor, addr string

type Photo struct {
	Id    int               `json:"id"`
	Dir   string            `json:"dir"`
	Name  string            `json:"name"`
	Ext   string            `json:"-"`
	Tags  []string          `json:"tags"`
	Edits map[string]*Photo `json:"edits"`
}

func Expand(in string) (out string) {
	if strings.HasPrefix(in, "~/") {
		usr, _ := user.Current()
		home_dir := usr.HomeDir
		out := strings.Replace(in, "~/", home_dir+"/", 1)
		return out
	}
	return in
}

func (p *Photo) String() string {
	return fmt.Sprintf("Photo {Id: %d, Dir: '%s', Name: '%s', Ext: '%s', Tags: '%s'}", p.Id, p.Dir, p.Name, p.Ext, p.Tags)
}

func NewPhoto(id int, dir string, name string, ext string, filetags map[string]string, edits_dir string, edits_filetags map[string]string) (p *Photo, err error) {

	// get our tags
	var tags_slice []string
	tags_str, ok := filetags[name]
	if ok {
		tags_slice = strings.Split(tags_str, ",")
	} else {
		tags_slice = make([]string, 0, 0)
	}

	p = &Photo{id, dir, name, ext, tags_slice, nil}

	if edits_dir != "" {
		err = p.LoadEdits(edits_dir, edits_filetags)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("WARNING: failed to load edits: %s\n", err))
		}
	}

	return p, nil

}

func (p *Photo) LoadEdits(edits_dir string, edits_filetags map[string]string) error {
	// /bleh/originals/foo/bar.jpg ->
	// /bleh/edits/foo/bar.jpg (standard edit)
	// /bleh/edits/foo/bar-speficier.jpg (specified edit)
	if edits_dir == "" {
		return nil
	}
	basename := p.Name[:len(p.Name)-len(p.Ext)]
	edit_base := path.Join(edits_dir, basename)
	glob_pattern := fmt.Sprintf("%s*%s", edit_base, p.Ext)
	p.Edits = make(map[string]*Photo)
	matches, err := filepath.Glob(glob_pattern)
	if err != nil {
		return err
	}
	// note i is not really used
	for i, match_path := range matches {
		key := strings.Replace(match_path, edit_base, "", 1)
		key = strings.Replace(key, p.Ext, "", 1)
		if key == "" {
			key = "unnamed"
		} else {
			key = key[1:] // strip leading '_', '-', etc
		}
		edit, err := NewPhoto(i, edits_dir, path.Base(match_path), p.Ext, edits_filetags, "", nil)
		if err != nil {
			return errors.New(fmt.Sprintf("could not initialize edit %s for %s: %s", key, p, err))
		}
		p.Edits[key] = edit
	}
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
	dir := r.Form.Get("dir")
	name := r.Form.Get("name")
	if dir == "" || name == "" {
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
		backend.Tag(w, r, conn_sqlite, path.Join(dir, name), tag)
	} else {
		backend.UnTag(w, r, conn_sqlite, path.Join(dir, name), untag)
	}
}

func find_edits_dir(dir string) (string, error) {
	edits_dir := strings.Replace(dir, "originals", "edits", 1)
	edits_dir = strings.Replace(edits_dir, "originals-generated", "edits", 1)
	if edits_dir == dir {
		fmt.Fprintf(os.Stderr, "WARNING: '%s' not in a format that allows finding an edits dir (needs 'originals' subdir)\n", dir)
		return "", nil
	}
	_, err := os.Stat(edits_dir)
	if err == nil {
		return edits_dir, nil
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	err = os.MkdirAll(edits_dir, os.ModeDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Failed to (recursively) create edits_dir '%s': %s\n", edits_dir, err)
		return "", nil
	}
	return edits_dir, nil

}

func IsPhoto(f os.FileInfo) (isPhoto bool, name string, ext string) {
	name = f.Name()
	ext = filepath.Ext(name)
	isPhoto = strings.HasPrefix(mime.TypeByExtension(strings.ToLower(ext)), "image/")
	return
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
	photos := make([]*Photo, 0, len(list))
	dir, err = filepath.Abs(dir)
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Cannot figure out directory abspath: '%s': %s", dir, err)}, 503)
		return
	}

	filetags, err := backend.GetFileTags(dir, conn_sqlite)
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Cannot get file tags for main dir '%s': %s", dir, err)}, 503)
		return
	}
	edits_filetags := make(map[string]string)
	if edits_dir != "" {
		edits_filetags, err = backend.GetFileTags(edits_dir, conn_sqlite)
		if err != nil {
			backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Cannot get file tags for edits_dir '%s': %s", dir, err)}, 503)
			return
		}
	}

	id := 0
	for _, f := range list {
		if isPhoto, name, ext := IsPhoto(f); isPhoto {
			p, err := NewPhoto(id, dir, name, ext, filetags, edits_dir, edits_filetags)
			if err != nil {
				fmt.Printf("WARNING: failed to create Photo instance: %s\n", err)
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

func api_edit_handler(w http.ResponseWriter, r *http.Request, conn_sqlite *sql.DB) {
	// this could be more elegant by getting entire photo as json and just converting json to go object.
	err := r.ParseForm()
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Invalid request: %s", err)}, 503)
		return
	}
	id, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Cannot parse id: %s", err)}, 503)
	}
	dir := r.Form.Get("dir")
	name := r.Form.Get("name")
	start_from := r.Form.Get("start_from")
	if dir == "" || name == "" {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Invalid request: %s", err)}, 503)
		return
	}
	edits_dir, err := find_edits_dir(dir)
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Edits directory '%s' seems to exist but unable to read: %s", edits_dir, err)}, 503)
		return
	}
	if edits_dir == "" {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("dir is not in a format that supports an edits dir")}, 409)
		return
	}

	var start_from_name string
	if start_from != "" {
		start_from_name = path.Base(start_from)
	} else {
		start_from = path.Join(dir, name)
		start_from_name = name
	}
	start_from_ext := path.Ext(start_from_name)
	tmp := path.Join(edits_dir, start_from_name[:len(start_from_name)-len(start_from_ext)]+"-save-your-edit-to-a-new-file"+start_from_ext)

	err = cp.Cp(tmp, start_from)
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Could not copy %s to %s because: %s", start_from, tmp, err)}, 503)
		return
	}
	// this doesn't actually help with preventing write. but I guess that doesn't really matter anyway.
	// the filename speaks for itself.
	err = os.Chmod(tmp, 0000)
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Could not make %s unwriteable: %s", tmp, err)}, 503)
		return
	}
	err = exec.Command(editor, tmp).Run()
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Edit did not complete successfully: %s", err)}, 503)
		return
	}
	err = os.Remove(tmp)
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Could not remove tmp file %s: %s", tmp, err)}, 503)
		return
	}

	filetags, err := backend.GetFileTags(dir, conn_sqlite)
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Cannot get file tags: '%s': %s", dir, err)}, 503)
		return
	}
	edits_filetags, err := backend.GetFileTags(edits_dir, conn_sqlite)
	if err != nil {
		backend.ErrorJson(w, backend.Resp{fmt.Sprintf("Cannot get file tags: '%s': %s", edits_dir, err)}, 503)
		return
	}

	p, err := NewPhoto(id, dir, name, path.Ext(name), filetags, edits_dir, edits_filetags)
	if err != nil {
		fmt.Printf("WARNING: failed to create Photo instance: %s\n", err)
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(p)
	if err != nil {
		fmt.Printf("WARNING: failed to encode/write json: %s\n", err)
	}
}

func serveThumb() http.Handler {
	h := http.FileServer(http.Dir(thumbnail_dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := workers.Resize("pixie-gothum", r.URL.Path, thumbnail_dir)
		if err != nil {
			http.Error(w, err.Error(), 500)
		} else {
			fmt.Println(r.URL.Path)
			// generate the thumb path again. we can optimize this later.
			fileUri_in := fmt.Sprintf("file://%s", r.URL.Path)
			hash := md5.New()
			io.WriteString(hash, fileUri_in)
			pathmd5 := fmt.Sprintf("%x", hash.Sum(nil))
			path_out := pathmd5 + ".png"
			r.URL.Path = path_out
			fmt.Println(r.URL.Path)
			h.ServeHTTP(w, r)
		}
	})
}
func main() {
	fmt.Println("Starting...")
	err := config.Parse(Expand("~/.pixie/config.ini"))
	if err != nil {
		log.Fatal("could not load config '~/.pixie/config.ini': " + err.Error() + " (did you set up as per the README?)")
	}
	thumbnail_dir = Expand(*ptr_thumbnail_dir)
	tmsu_file = Expand(*ptr_tmsu_file)
	addr = *ptr_addr
	editor = *ptr_editor
	conn_sqlite, err := sql.Open("sqlite3", tmsu_file)
	if err != nil {
		log.Fatal("could not open database: ", err.Error())
	}
	http.HandleFunc("/api/photos/", func(w http.ResponseWriter, r *http.Request) {
		api_photos_handler(w, r, conn_sqlite)
	})
	http.HandleFunc("/api/photo", func(w http.ResponseWriter, r *http.Request) {
		api_photo_handler(w, r, conn_sqlite)
	})
	http.HandleFunc("/api/edit", func(w http.ResponseWriter, r *http.Request) {
		api_edit_handler(w, r, conn_sqlite)
	})
	http.Handle("/api/config/binds", http.StripPrefix("/api/config", http.FileServer(http.Dir(Expand("~/.pixie")))))

	http.Handle("/thumbnails/", http.StripPrefix("/thumbnails", serveThumb()))
	http.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Printf("starting up on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
