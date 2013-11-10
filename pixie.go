package main

import (
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
	"path"
	"path/filepath"
	"strings"
)

var thumbnail_dir = config.String("thumbnail_dir", "")
var tmsu_file = config.String("tmsu_file", "")

type Photo struct {
	Id    int      `json:"id"`
	Name  string   `json:"name"`
	Dir   string   `json:"dir"`
	Thumb string   `json:"thumb"`
	Tags  []string `json:"tags"`
}

func (p *Photo) String() string {
	return fmt.Sprintf("Photo {Id: %d, Name: '%s', Dir: '%s', Thumb: '%s'}", p.Id, p.Name, p.Dir, p.Thumb)
}

// tag a given fname with a given tag
func api_photo_handler(w http.ResponseWriter, r *http.Request, conn_sqlite *sql.DB) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), 503)
		return
	}
	tag := r.Form.Get("tag")
	fname := r.Form.Get("fname")
	if tag == "" || fname == "" {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), 503)
		return
	}
	// for some reason tmsu has no unique constraint on a tag name, so we have to do this racey thing:
	var tag_id int
	query := `SELECT id from tag where name = ?`
	err = conn_sqlite.QueryRow(query, tag).Scan(&tag_id)
	switch {
	case err == sql.ErrNoRows:
		tag_id = -1
	case err != nil:
		http.Error(w, fmt.Sprintf("Cannot query sqlite: for tag '%s': %s", tag, err), 503)
		return
	default:
		fmt.Println("")
	}
	if tag_id == -1 {
		query = `INSERT INTO tag (name) VALUES (?)`
		result, err := conn_sqlite.Exec(query, tag)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot insert tag '%s' into sqlite: %s", tag, err), 503)
			return
		}
		tmp, err := result.LastInsertId()
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot get id from sqlite of just inserted tag '%s'", tag), 503)
			return
		}
		tag_id = int(tmp)
	}
	var file_id int
	basename := path.Base(fname)
	dirname := path.Dir(fname)
	query = `SELECT id from file where directory = ? and name = ?`
	err = conn_sqlite.QueryRow(query, dirname, basename).Scan(&file_id)

	switch {
	case err == sql.ErrNoRows:
		file_id = -1
	case err != nil:
		http.Error(w, fmt.Sprintf("Cannot query sqlite for file '%s': %s", fname, err), 503)
		return
	default:
		fmt.Println("")
	}
	if file_id == -1 {
		// i don't really use the fingerprint and mod_time, that's more of a tmsu thing.
		query = `insert into file (directory, name, fingerprint, mod_time) values (?, ?, 'pixie', '2013-01-01')`
		result, err := conn_sqlite.Exec(query, dirname, basename)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot insert file '%s' into sqlite: %s", fname, err), 503)
			return
		}
		tmp, err := result.LastInsertId()
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot get id from sqlite of just inserted file '%s'", fname), 503)
			return
		}
		file_id = int(tmp)
	}

	// also this is a little racey because tmsu doesn't use a constraint
	var file_tag_id int
	query = `select id from file_tag where file_id = ? and tag_id = ?`
	err = conn_sqlite.QueryRow(query, file_id, tag_id).Scan(&file_tag_id)
	switch {
	case err == sql.ErrNoRows:
		file_tag_id = -1
	case err != nil:
		http.Error(w, fmt.Sprintf("Cannot query sqlite for file '%s' and tag '%s': %s", fname, tag, err), 503)
		return
	default:
		Json(w, Resp{"tag already existed"})
		return
	}
	query = `insert into file_tag (file_id, tag_id) values (?, ?)`
	_, err = conn_sqlite.Exec(query, file_id, tag_id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot insert file '%s' - tag mapping tag '%s' into sqlite: %s", fname, tag, err), 503)
		return
	}
	Json(w, Resp{"tag saved"})
}

type Resp struct {
	Msg string `json:"msg"`
}

func Json(w http.ResponseWriter, resp Resp) {
	enc := json.NewEncoder(w)
	err := enc.Encode(resp)
	if err != nil {
		fmt.Printf("WARNING: failed to encode/write json: %s\n", err)
	}
	return
}
func ErrorJson(w http.ResponseWriter, resp Resp, code int) {
	enc := json.NewEncoder(w)
	err := enc.Encode(resp)
	if err != nil {
		fmt.Printf("WARNING: failed to encode/write json: %s\n", err)
	}
	http.Error(w, "", 503)

}

// get a list of photos (with tags) for a given dir
func api_photos_handler(w http.ResponseWriter, r *http.Request, conn_sqlite *sql.DB) {
	dir := strings.Replace(r.URL.Path, "/api/photos", "", 1)
	fmt.Printf("reading dir '%s'\n", dir)
	list, err := ioutil.ReadDir(dir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot read directory: '%s': %s", dir, err), 503)
		return
	}
	photos := make([]Photo, 0, len(list))
	dir, err = filepath.Abs(dir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot figure out directory abspath: '%s': %s", dir, err), 503)
		return
	}

	filetags, err := getFileTags(dir, conn_sqlite)
	if err != nil {
		http.Error(w, fmt.Sprintf("Cannot get file tags: '%s': %s", dir, err), 503)
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
			p := Photo{id, name, dir, thumb, tags_slice}
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

func getFileTags(dir string, conn_sqlite *sql.DB) (map[string]string, error) {
	sql := `select f.name, group_concat(t.name) from file as f
    left join file_tag as ft on f.id == ft.file_id
    left join tag as t on t.id == ft.tag_id
    where directory = ?
    group by f.id`
	rows, err := conn_sqlite.Query(sql, dir)
	if err != nil {
		return nil, err
	}
	filetags := make(map[string]string)
	for rows.Next() {
		var fname string
		var tags string
		err := rows.Scan(&fname, &tags)
		if err != nil {
			return nil, err
		}
		filetags[fname] = tags
		fmt.Printf("tags: '%s'\n", tags)
	}
	return filetags, nil

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
	http.ListenAndServe(addr, nil)
}
