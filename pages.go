package main

import (
	"fmt"
	"log"
	"net/http"

	"text/template"

	"github.com/gorilla/mux"
)

// Renders a page to a http response
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

// Renders the editor
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

	t, err := template.ParseFiles("editor.html", "nav.html")

	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/admin", 302)
	}

	err = t.Execute(w, page)

	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/admin", 302)
	}

}

// Takes a POST form that needs a title, author, summary, body, target (URL), visibility (This should just be a checkbox, only checks to see if it exists)
// and a meta string
// It then redirects to the created pages editor
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

// Gives a new key to the requested page
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

// Deletes the given page
func deletePage(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

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

// Creates a page with a POST form given title
func createPage(w http.ResponseWriter, r *http.Request) {

	checkLogin(w, r)
	if r.Method != "POST" {
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

		ajaxResponse(w, r, false, "", "That title/URL is already in use")

		log.Print("Insert Failed")
		return
	}

	//http.Redirect(w, r, "/pages/"+md.URL+"/"+md.Key, 302)

	ajaxResponse(w, r, true, "/pages/"+page.URL+"/"+page.Key, "")

}
