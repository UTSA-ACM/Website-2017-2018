package main

import (
	"io/ioutil"
	"os"

	"database/sql"

	"log"

	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var (
	db              *sql.DB
	defaultPassword = "goadmin"
)

func start() {
	dirEntries, err := ioutil.ReadDir(".")

	if err != nil {
		log.Fatal(err)
	}

	if !contains(dirEntries, "blog.db") {
		os.Create("./blog.db")

		db, err = sql.Open("sqlite3", "./blog.db")

		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(
			"CREATE TABLE posts(id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT UNIQUE, url TEXT UNIQUE, author TEXT, summary TEXT, markdown TEXT, target TEXT, key TEXT, visible INT, created INTEGER, meta TEXT);" +
				"CREATE TABLE users(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT UNIQUE, hash TEXT);" +
				"CREATE TABLE tags(id INTEGER PRIMARY KEY AUTOINCREMENT, tag TEXT, post INT);")

		if err != nil {
			log.Fatal(err)
		}

		stmt, err := db.Prepare("INSERT INTO users(name, hash) values(?,?)")

		if err != nil {
			log.Fatal(err)
		}

		_, err = stmt.Exec("admin", getHashString(defaultPassword))

		if err != nil {
			log.Fatal(err)
		}

	} else {
		db, err = sql.Open("sqlite3", "blog.db")

		if err != nil {
			log.Fatal(err)
		}
	}

}

func insertDBPage(page *Page) int64 {

	stmt, err := db.Prepare("INSERT INTO posts(title, url, author, summary, markdown, target, key, visible, created, meta) values(?,?,?,?,?,?,?,?,?,?)")
	defer stmt.Close()
	if err != nil {
		log.Print(err)
		return -1
	}

	res, err := stmt.Exec(page.Title, page.URL, page.Author, page.Summary, page.Body, page.Target, page.Key, page.Visible, page.Datetime, page.Meta)

	if err != nil {
		log.Print(err)
		return -1
	}

	id, err := res.LastInsertId()

	if err != nil {
		log.Print(err)
		return -1
	}

	return id
}

func updateDBPage(url string, page *Page) string {

	if page.Title == "" {
		return ""
	}

	page.URL = titleToUrl(page.Title)

	if page.URL == "" {
		return ""
	}

	stmt, err := db.Prepare("UPDATE posts SET title = ?, url = ?, author = ?, summary = ?, markdown = ?, target = ?, visible = ?, meta = ?, key = ? WHERE url=?")
	defer stmt.Close()

	if err != nil {
		log.Fatal("updateDBPage:", err)
	}

	res, err := stmt.Exec(page.Title, page.URL, page.Author, page.Summary, page.Body, page.Target, page.Visible, page.Meta, page.Key, url)

	if err != nil {
		log.Fatal("updateDBPage:", err)
	}

	_, err = res.LastInsertId()

	if err != nil {
		log.Fatal("updateDBPage:", err)
	}

	return page.URL
}

func deleteDBPage(url string) error {

	stmt, err := db.Prepare("DELETE FROM posts WHERE url=?")
	defer stmt.Close()

	if err != nil {
		log.Print(err)
		return err
	}

	_, err = stmt.Exec(url)

	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func getLastID() int {

	rows, err := db.Query("SELECT seq FROM sqlite_sequence")

	if err != nil {
		log.Print(err)
		return -1
	}

	var count int

	for rows.Next() {
		err := rows.Scan(&count)

		if err != nil {
			log.Fatal(err)
		}
	}

	return count
}

func getVisibleRowCount() int {
	rows, err := db.Query("SELECT Count(*) FROM posts WHERE visible = 1")

	if err != nil {
		log.Print(err)
		return -1
	}

	var count int

	for rows.Next() {
		err := rows.Scan(&count)

		if err != nil {
			log.Print(err)
			count = -1
		}
	}

	return count
}

func getPagesSortedByDate(page, count int, visibleOnly bool) ([]Page, int) {

	var visibility string

	if visibleOnly {
		visibility = "WHERE visible = 1"
	} else {
		visibility = ""
	}

	query := fmt.Sprintf("SELECT * FROM posts %v ORDER BY created DESC LIMIT ? OFFSET ?", visibility)

	rows, err := db.Query(query, count, count*page)
	defer rows.Close()

	if err != nil {
		log.Fatal(err)
	}

	var posts []Page

	var id int

	for rows.Next() {
		var page Page
		err = rows.Scan(&id, &page.Title, &page.URL, &page.Author, &page.Summary, &page.Body, &page.Target, &page.Key, &page.Visible, &page.Datetime, &page.Meta)

		if err != nil {
			log.Fatal(err)
		}

		posts = append(posts, page)
	}

	return posts, id

}

func getDBPage(url string) Page {
	rows, err := db.Query("SELECT * FROM posts WHERE url=?", url)
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}

	var page Page
	var id int

	for rows.Next() {
		err = rows.Scan(&id, &page.Title, &page.URL, &page.Author, &page.Summary, &page.Body, &page.Target, &page.Key, &page.Visible, &page.Datetime, &page.Meta)
		if err != nil {
			log.Fatal(err)
		}
		break
	}

	return page
}

func checkUser(name, password string) bool {

	rows, err := db.Query("SELECT hash FROM users WHERE name=?", name)
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}

	var (
		hash string
	)

	for rows.Next() {
		err = rows.Scan(&hash)

		if err != nil {
			log.Fatal(err)
		}

		break

	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return false
	}

	return true
}

func updatePassword(username, old, new string) bool {

	if checkUser(username, old) {

		newhash := getHashString(new)

		stmt, err := db.Prepare("UPDATE users SET hash = ? WHERE name = ?")
		defer stmt.Close()

		if err != nil {
			log.Print(err)
			return false
		}

		_, err = stmt.Exec(newhash, username)

		if err != nil {
			log.Print(err)
			return false
		}

		return true
	}

	return false
}

func contains(fi []os.FileInfo, name string) bool {

	for _, file := range fi {
		if file.Name() == name {
			return true
		}
	}
	return false
}

func getHashString(password string) string {
	return string(getHashByte([]byte(password)))
}

func getHashByte(password []byte) []byte {
	out, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)

	if err != nil {
		log.Fatal(err)
	}

	return out
}
