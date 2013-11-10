package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"path"
)

func Tag(w http.ResponseWriter, r *http.Request, conn_sqlite *sql.DB, fname string, tag string) {
	// for some reason tmsu has no unique constraint on a tag name, so we have to do this racey thing:
	var tag_id int
	query := `SELECT id from tag where name = ?`
    err := conn_sqlite.QueryRow(query, tag).Scan(&tag_id)
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

func GetFileTags(dir string, conn_sqlite *sql.DB) (map[string]string, error) {
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
