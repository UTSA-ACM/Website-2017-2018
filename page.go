package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"text/template"
	"time"

	"log"

	"github.com/microcosm-cc/bluemonday"
)

// Page is a  post
type Page struct {
	Title    string
	URL      string
	Author   string
	Summary  string
	Body     string
	Key      string // Key is used to store the url key that will allow editing of the page
	Target   string // This will allow you to have a pass through to another page
	Visible  int
	Datetime int
	Meta     string
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Creates a new page object, and sets the fields to defaults and generates the key and URL
func newPage(title, author, summary, body, target, meta string) *Page {

	var md Page

	md.Title = title
	md.URL = titleToURL(title)
	md.Author = author
	md.Summary = summary
	md.Body = body
	md.Target = target
	md.Key = generateKey()
	md.Visible = 0
	md.Datetime = int(time.Now().Unix())
	md.Meta = meta

	return &md
}

// Generates the key string for editing
func generateKey() string {
	b := make([]byte, 7)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Turns a title into a URL friendly string
func titleToURL(title string) string {

	reg, err := regexp.Compile("[ ]+")

	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	title = reg.ReplaceAllString(title, "-")

	reg, err = regexp.Compile("[^a-zA-Z0-9-]+")

	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	url := reg.ReplaceAllString(title, "")

	return url

}

// Deprecated
func templateString(tstring string, data Page) (string, error) {
	var tbuf bytes.Buffer

	t := template.New(tstring)

	tstring = "templates/" + tstring

	t, err := t.ParseFiles(tstring)

	if err != nil {
		fmt.Print(err)
	}

	err = t.Execute(&tbuf, data)

	if err != nil {
		fmt.Print("template String", err)
		return "", err
	}

	return tbuf.String(), nil

}

// Renders a post
func renderPage(w http.ResponseWriter, page Page) {

	sanitizePage(&page)

	t, err := template.ParseFiles("templates/markdown.html", "templates/nav.html")

	err = t.Execute(w, page)

	if err != nil {
		log.Print(err)
	}

}

// Sanitizes the page struct
func sanitizePage(md *Page) {
	p := bluemonday.UGCPolicy()
	p.AllowDataURIImages()
	p.AllowImages()
	p.AllowAttrs("class").Globally()

	md.Body = p.Sanitize(md.Body)
	md.Author = p.Sanitize(md.Author)
	md.Summary = p.Sanitize(md.Summary)
	md.Title = p.Sanitize(md.Title)
	md.Target = p.Sanitize(md.Target)
	md.Meta = p.Sanitize(md.Meta)
}
