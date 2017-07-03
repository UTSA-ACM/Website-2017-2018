package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func getPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page := getDBPage(vars["url"])

	if page.Title == "" {
		notFound(w, r)
		return
	}

	if page.Target != "" {
		http.Redirect(w, r, page.Target, 302)
	}

	renderPage(w, page)
}

func pageEditor(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	page := getDBPage(vars["url"])

	if page.Title == "" {
		notFound(w, r)
		return
	}

	if page.Key != vars["key"] {
		fmt.Fprint(w, "Access Denied")
		return
	}

	sanitizePage(&page)

	out, err := templateString("editor.html", page)

	if err != nil {
		fmt.Print(err)
	}

	fmt.Fprint(w, out)

}

func updatePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	page := getDBPage(vars["url"])

	if page.Title == "" {
		notFound(w, r)
		return
	}

	if page.Key != vars["key"] {
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

	page.Title = title
	page.Author = author
	page.Summary = summary
	page.Body = body
	page.Target = target
	page.Meta = meta

	if visible == "" {
		page.Visible = 0
	} else {
		page.Visible = 1
	}

	newURL := updateDBPage(vars["url"], &page)

	if newURL == "" {
		http.Redirect(w, r, "/pages/"+page.URL+"/"+page.Key, 302)
	}

	http.Redirect(w, r, "/pages/"+newURL+"/"+page.Key, 302)

}

func reKey(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

	vars := mux.Vars(r)

	page := getDBPage(vars["url"])

	if page.Title == "" {
		notFound(w, r)
		return
	}

	page.Key = generateKey()

	updateDBPage(vars["url"], &page)

	http.Redirect(w, r, "/admin", 302)
}

func deletePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	page := getDBPage(vars["url"])

	if page.Title == "" {
		notFound(w, r)
		return
	}

	if page.Key != vars["key"] {
		fmt.Fprint(w, "Access Denied")
		return
	}

	deleteDBPage(page.URL)

	http.Redirect(w, r, "/admin", 302)
}

func createPage(w http.ResponseWriter, r *http.Request) {

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

	page := newPage(title, "", "", "", "", "")

	if insertDBPage(page) == -1 {

		ajaxResponse(w, r, false, "", "Page creation failed")

		log.Print("Insert Failed")
		return
	}

	//http.Redirect(w, r, "/pages/"+md.URL+"/"+md.Key, 302)

	ajaxResponse(w, r, true, "/pages/"+page.URL+"/"+page.Key, "")

}
