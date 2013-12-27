package backend

import (
	"bitbucket.org/oniony/tmsu/src/tmsu/fingerprint"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
	"path"
)

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
	j, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("WARNING: failed to encode json: %s\n", err)
	}
	http.Error(w, string(j), code)

}

// assert a file exists (create if needed) and return id
func get_fileid(conn_sqlite *sql.DB, fname string) (file_id int, err error) {
	basename := path.Base(fname)
	dirname := path.Dir(fname)
	query := `SELECT id from file where directory = ? and name = ?`
	err = conn_sqlite.QueryRow(query, dirname, basename).Scan(&file_id)

	switch {
	case err == sql.ErrNoRows:
		file_id = -1
	case err != nil:
		return -1, errors.New(fmt.Sprintf("Cannot query sqlite for file '%s': %s", fname, err))
	default:
		fmt.Println("")
	}
	if file_id == -1 {
		stat, err := os.Stat(fname)
		if err != nil {
			return -1, errors.New(fmt.Sprintf("Cannot stat file '%s': %s", fname, err))
		}
		fingerprint, err := fingerprint.Create(fname)
		if err != nil {
			return -1, errors.New(fmt.Sprintf("Cannot create fingerprint for file '%s': %s", fname, err))
		}
		modTime := stat.ModTime()
		size := uint(stat.Size())

		query = `insert into file (directory, name, fingerprint, mod_time, size, is_dir) values (?, ?, ?, ?, ?, 0)`
		result, err := conn_sqlite.Exec(query, dirname, basename, string(fingerprint), modTime, size)
		if err != nil {
			return -1, errors.New(fmt.Sprintf("Cannot insert file '%s' into sqlite: %s", fname, err))
		}
		tmp, err := result.LastInsertId()
		if err != nil {
			return -1, errors.New(fmt.Sprintf("Cannot get id from sqlite of just inserted file '%s'", fname))
		}
		file_id = int(tmp)
	}
	return file_id, nil
}

// assert a tag exists (create if needed) and return id
func get_tagid(conn_sqlite *sql.DB, tag string) (tag_id int, err error) {
	// for some reason tmsu has no unique constraint on a tag name, so we have to do this racey thing:
	query := `SELECT id from tag where name = ?`
	err = conn_sqlite.QueryRow(query, tag).Scan(&tag_id)
	switch {
	case err == sql.ErrNoRows:
		tag_id = -1
	case err != nil:
		return -1, errors.New(fmt.Sprintf("Cannot query sqlite: for tag '%s': %s", tag, err))
	default:
		fmt.Println("read tag ", tag, "id", tag_id)
	}
	if tag_id == -1 {
		query = `INSERT INTO tag (name) VALUES (?)`
		result, err := conn_sqlite.Exec(query, tag)
		if err != nil {
			return -1, errors.New(fmt.Sprintf("Cannot insert tag '%s' into sqlite: %s", tag, err))
		}
		tmp, err := result.LastInsertId()
		if err != nil {
			return -1, errors.New(fmt.Sprintf("Cannot get id from sqlite of just inserted tag '%s'", tag))
		}
		tag_id = int(tmp)
	}
	return tag_id, nil
}

func Tag(w http.ResponseWriter, r *http.Request, conn_sqlite *sql.DB, fname string, tag string) {
	tag_id, err := get_tagid(conn_sqlite, tag)
	if err != nil {
		ErrorJson(w, Resp{err.Error()}, 503)
		return
	}
	file_id, err := get_fileid(conn_sqlite, fname)
	if err != nil {
		ErrorJson(w, Resp{err.Error()}, 503)
		return
	}
	query := `replace into file_tag (file_id, tag_id) values (?, ?)`
	_, err = conn_sqlite.Exec(query, file_id, tag_id)
	if err != nil {
		ErrorJson(w, Resp{fmt.Sprintf("Cannot create tag mapping: %s", err)}, 503)
		return
	}
	Json(w, Resp{"tag saved (or already existed)"})
}

func UnTag(w http.ResponseWriter, r *http.Request, conn_sqlite *sql.DB, fname string, tag string) {
	tag_id, err := get_tagid(conn_sqlite, tag)
	if err != nil {
		ErrorJson(w, Resp{err.Error()}, 503)
		return
	}
	file_id, err := get_fileid(conn_sqlite, fname)
	if err != nil {
		ErrorJson(w, Resp{err.Error()}, 503)
		return
	}
	query := `DELETE FROM file_tag where file_id = ? and tag_id = ?`
	result, err := conn_sqlite.Exec(query, file_id, tag_id)
	if err != nil {
		ErrorJson(w, Resp{fmt.Sprintf("tag remove failed: %s", err)}, 503)
		return
	}
	ra, err := result.RowsAffected()
	if err != nil {
		ErrorJson(w, Resp{fmt.Sprintf("Cannot check if tag remove worked: %s", err)}, 503)
		return
	}
	if ra == 0 {
		ErrorJson(w, Resp{"No tag relationship. Nothing to remove"}, 503)
		return
	}

	Json(w, Resp{"tag removed"})
}

func GetFileTags(dir string, conn_sqlite *sql.DB) (map[string]string, error) {
	query := `select f.name, group_concat(t.name) from file as f
    left join file_tag as ft on f.id == ft.file_id
    left join tag as t on t.id == ft.tag_id
    where directory = ?
    group by f.id`
	rows, err := conn_sqlite.Query(query, dir)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("query execution failed: '%s'", err))
	}
	filetags := make(map[string]string)
	for rows.Next() {
		//var fname string
		//var tags string
		// even though both values are TEXT NOT NULL, we need this, or the driver
		// errors about unsupported time. dunno why?
		var fname sql.NullString
		var tags sql.NullString
		err := rows.Scan(&fname, &tags)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("rows scan failed: '%s'", err))
		}
		filetags[fname.String] = tags.String
	}
	return filetags, nil
}
