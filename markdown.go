package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"text/template"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

// Markdown is a markdown post
type Markdown struct {
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

func newMarkdown(title, author, summary, body, target, meta string) *Markdown {

	var md Markdown

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

func templateString(tstring string, data Markdown) (string, error) {
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

// func readMarkdown(filename string) Markdown {
// 	b, err := ioutil.ReadFile("Markdown/" + filename + ".md")
// 	if err != nil {
// 		fmt.Print(err)
// 	}

// 	var md Markdown
// 	md.Body = string(b)

// 	return md
// }

// func renderMarkdown(filename string) string {

// 	md := readMarkdown(filename)

// 	tstring, _ := templateString("Markdown.html", md)

// 	return tstring

// }

func readJson(name string) Markdown {
	b, err := ioutil.ReadFile("json/" + name + ".json")
	if err != nil {
		fmt.Print(err)
	}
	var md Markdown
	json.Unmarshal(b, &md)

	md.Body = readMD(md.Title)

	return md
}

func readMD(name string) string {
	b, err := ioutil.ReadFile("markdown/" + name + ".md")

	if err != nil {
		fmt.Print(err)
	}

	return string(b)
}

func writeJson(name string, md Markdown) {
	b, _ := json.Marshal(md)
	ioutil.WriteFile("markdown/"+name+".json", b, 0660)
}

func renderMarkdown(w http.ResponseWriter, md Markdown) {

	p := bluemonday.UGCPolicy()
	p.AllowDataURIImages()
	p.AllowAttrs("class").Globally()

	md.Body = p.Sanitize(md.Body)

	out, err := templateString("markdown.html", md)

	if err != nil {
		fmt.Print(err)
	}

	fmt.Fprint(w, out)

}

func sanitizeMarkdown(md *Markdown) {
	p := bluemonday.UGCPolicy()
	p.AllowDataURIImages()
	p.AllowAttrs("class").Globally()

	md.Body = p.Sanitize(md.Body)
	md.Author = p.Sanitize(md.Author)
	md.Summary = p.Sanitize(md.Summary)
	md.Title = p.Sanitize(md.Title)
	md.Target = p.Sanitize(md.Target)
	md.Meta = p.Sanitize(md.Meta)
}
