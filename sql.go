package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"database/sql"

	"log"

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
		fmt.Print(err)
	}

	if !contains(dirEntries, "blog.db") {
		os.Create("./blog.db")

		db, err = sql.Open("sqlite3", "./blog.db")

		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec("CREATE TABLE posts(id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, url TEXT, author TEXT, summary TEXT, markdown TEXT, target TEXT, key TEXT, visible INT);" +
			"CREATE TABLE users(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, hash TEXT);")

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

func insertMarkdown(md *Markdown) int64 {

	stmt, err := db.Prepare("INSERT INTO posts(title, url, author, summary, markdown, target, key, visible) values(?,?,?,?,?,?,?,?)")

	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	res, err := stmt.Exec(md.Title, md.URL, md.Author, md.Summary, md.Body, md.Target, md.Key, md.Visible)

	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	id, err := res.LastInsertId()

	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	return id
}

func getMarkdown(url string) Markdown {
	rows, err := db.Query("SELECT * FROM posts WHERE url=?", url)

	if err != nil {
		log.Fatal(err)
	}

	var md Markdown
	var id int

	for rows.Next() {
		err = rows.Scan(&id, &md.Title, &md.URL, &md.Author, &md.Summary, &md.Body, &md.Target, &md.Key, &md.Visible)
		if err != nil {
			log.Fatal(err)
		}
	}

	return md
}

func checkUser(name, password string) bool {

	rows, err := db.Query("SELECT hash FROM users WHERE name=?", name)

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

func contains(fi []os.FileInfo, name string) bool {
	fmt.Print(fi[0].Name())
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
