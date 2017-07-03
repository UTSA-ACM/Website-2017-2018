package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func markdownPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	md := getMarkdown(vars["url"])

	if md.Title == "" {
		notFound(w, r)
		return
	}

	if md.Target != "" {
		http.Redirect(w, r, md.Target, 302)
	}

	renderMarkdown(w, md)
}

func pageEditor(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	md := getMarkdown(vars["url"])

	if md.Title == "" {
		notFound(w, r)
		return
	}

	if md.Key != vars["key"] {
		fmt.Fprint(w, "Access Denied")
		return
	}

	sanitizeMarkdown(&md)

	out, err := templateString("editor.html", md)

	if err != nil {
		fmt.Print(err)
	}

	fmt.Fprint(w, out)

}

func updatePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	md := getMarkdown(vars["url"])

	if md.Title == "" {
		notFound(w, r)
		return
	}

	if md.Key != vars["key"] {
		fmt.Fprint(w, "Access Denied")
		return
	}

	title := r.PostFormValue("title")
	author := r.PostFormValue("author")
	summary := r.PostFormValue("summary")
	body := r.PostFormValue("body")
	target := r.PostFormValue("target")
	visible := r.PostFormValue("visible")
	meta := r.PostFormValue("meta")

	md.Title = title
	md.Author = author
	md.Summary = summary
	md.Body = body
	md.Target = target
	md.Meta = meta

	if visible == "" {
		md.Visible = 0
	} else {
		md.Visible = 1
	}

	newURL := updateMarkdown(vars["url"], &md)

	if newURL == "" {
		http.Redirect(w, r, "/pages/"+md.URL+"/"+md.Key, 302)
	}

	http.Redirect(w, r, "/pages/"+newURL+"/"+md.Key, 302)

}

func reKey(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

	vars := mux.Vars(r)

	md := getMarkdown(vars["url"])

	if md.Title == "" {
		notFound(w, r)
		return
	}

	md.Key = generateKey()

	updateMarkdown(vars["url"], &md)

	http.Redirect(w, r, "/admin", 302)
}

func deletePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	md := getMarkdown(vars["url"])

	if md.Title == "" {
		notFound(w, r)
		return
	}

	if md.Key != vars["key"] {
		fmt.Fprint(w, "Access Denied")
		return
	}

	deleteMarkdown(md.URL)

	http.Redirect(w, r, "/admin", 302)
}

func newPage(w http.ResponseWriter, r *http.Request) {

	checkLogin(w, r)
	if r.Method != "PUT" {
		//http.Redirect(w, r, "/admin", 302)

		//t := [3]int{1, 2, 3}

		ajaxResponse(w, r, false, "", "Must be PUT Request")

		return
	}

	title := r.PostFormValue("title")

	if title == "" {

		ajaxResponse(w, r, false, "", "Title cannot be blank")
		return
	}

	md := newMarkdown(title, "", "", "", "", "")

	if insertMarkdown(md) == -1 {

		ajaxResponse(w, r, false, "", "Page creation failed")

		log.Print("Insert Failed")
		return
	}

	//http.Redirect(w, r, "/pages/"+md.URL+"/"+md.Key, 302)

	ajaxResponse(w, r, true, "/pages/"+md.URL+"/"+md.Key, "")

}
