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

// Page is a markdown post
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

func newPage(title, author, summary, body, target, meta string) *Page {

	var md Page

	md.Title = title
	md.URL = titleToUrl(title)
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

func generateKey() string {
	b := make([]byte, 7)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func titleToUrl(title string) string {

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

func renderPage(w http.ResponseWriter, page Page) {

	sanitizePage(&page)

	t, err := template.ParseFiles("templates/markdown.html", "templates/nav.html")

	err = t.Execute(w, page)

	if err != nil {
		log.Print(err)
	}

}

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
