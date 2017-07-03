package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"time"

	"log"

	"strconv"

	mux "github.com/gorilla/mux"
)

func index(w http.ResponseWriter, r *http.Request) {
	page := 0

	qpage := r.URL.Query().Get("page")

	if qpage != "" {
		var err error
		tpage, err := strconv.ParseInt(qpage, 0, 64)

		if err != nil {
			log.Print(err)
			http.Redirect(w, r, "/admin", 302)
			return
		}
		page = int(tpage)
	}

	dashboardTemplate, err := template.ParseFiles("templates/index.html")

	if err != nil {
		log.Fatal(err)
	}

	var posts []Markdown

	posts, _ = getPostsSortedByDate(page*10, 10, false, true)

	next := page + 1

	if getLastID() <= next*10 {
		next = page
	}

	prev := page - 1

	if page == 0 {
		prev = 0
	}

	data := struct {
		Page  int
		Next  int
		Prev  int
		Posts []Markdown
	}{
		page,
		next,
		prev,
		posts}

	err = dashboardTemplate.Execute(w, data)

	if err != nil {
		log.Print(err)
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)

	t, err := template.ParseFiles("./templates/404.html")

	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/", 302)
	}

	err = t.Execute(w, nil)

	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/", 302)
	}
}

func main() {

	rand.Seed(time.Now().UTC().UnixNano())

	start()

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", index)
	r.HandleFunc("/admin", dashboard)
	r.HandleFunc("/pages/{url}", markdownPage)
	r.HandleFunc("/pages/{url}/{key}", pageEditor)
	r.HandleFunc("/pages/{url}/{key}/update", updatePage)
	r.HandleFunc("/pages/{url}/{key}/delete", deletePage)
	r.HandleFunc("/admin/{url}/rekey", reKey)
	r.HandleFunc("/admin/new", newPage)
	r.HandleFunc("/admin/account", accountManagement)
	r.HandleFunc("/admin/password", newPassword)
	r.HandleFunc("/login", login)
	r.HandleFunc("/logout", logout)
	r.HandleFunc("/verify", verify)
	r.HandleFunc("/check", checkLogin)

	statichandler := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static", statichandler))

	r.NotFoundHandler = http.HandlerFunc(notFound)

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func renderStatic(w http.ResponseWriter, templateName string) {
	t := template.New(templateName)

	tstring := "templates/" + templateName

	t, err := t.ParseFiles(tstring)

	if err != nil {
		fmt.Print(err)
	}

	err = t.Execute(w, struct{}{})
}
