package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"time"

	"log"

	mux "github.com/gorilla/mux"
)

func markdownPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	md := getMarkdown(vars["url"])
	renderMarkdown(w, md)
}

func pageEditor(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	md := getMarkdown(vars["url"])

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

	md.Title = title
	md.Author = author
	md.Summary = summary
	md.Body = body
	md.Target = target

	if visible == "" {
		md.Visible = 0
	} else {
		md.Visible = 1
	}

	newURL := updateMarkdown(vars["url"], &md)

	http.Redirect(w, r, "/page/"+newURL+"/"+md.Key, 302)

}

func deletePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	md := getMarkdown(vars["url"])

	if md.Key != vars["key"] {
		fmt.Fprint(w, "Access Denied")
		return
	}

	deleteMarkdown(md.URL)

	http.Redirect(w, r, "/admin", 302)
}

func newPage(w http.ResponseWriter, r *http.Request) {

	title := r.PostFormValue("title")

	md := newMarkdown(title, "", "", "", "")

	if insertMarkdown(md) == -1 {
		log.Fatal("Insert Failed")
	}

	http.Redirect(w, r, "/page/"+md.URL+"/"+md.Key, 302)

}

func dashboard(w http.ResponseWriter, r *http.Request) {

	checkLogin(w, r)

	dashboardTemplate, err := template.ParseFiles("templates/dashboard.html")

	if err != nil {
		log.Fatal(err)
	}

	var posts []Markdown

	posts, _ = getPostsSortedByDate(0, 10, true)

	dashboardTemplate.Execute(w, posts)

}

func login(w http.ResponseWriter, r *http.Request) {
	renderStatic(w, "login.html")
}

func verify(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	password := r.PostFormValue("password")

	if checkUser(name, password) {

		token := newSession(name)

		cookie := http.Cookie{Name: "token", Value: token, MaxAge: 259200}
		http.SetCookie(w, &cookie)
		//fmt.Fprint(w, "cookie should be made")
		http.Redirect(w, r, "/admin", 302)

	} else {
		cookie := http.Cookie{Name: "token", Value: "", MaxAge: 0}
		http.SetCookie(w, &cookie)
		fmt.Fprint(w, "failure")
	}
}

func loggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("token")

	if err != nil {
		return false
	}

	token := cookie.Value

	username := checkSession(token)

	if username == "" {

		return false
	}

	return true
}

func checkLogin(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")

	if err != nil {

		http.Redirect(w, r, "/login", 302)
		return

	}

	token := cookie.Value

	username := checkSession(token)

	if username == "" {
		http.Redirect(w, r, "/login", 302)
		return
	}

	//http.Redirect(w, r, "/admin", 302)

}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to Epos!")
}

func main() {

	rand.Seed(time.Now().UTC().UnixNano())

	start()

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", index)
	r.HandleFunc("/admin", dashboard)
	r.HandleFunc("/page/{url}", markdownPage)
	r.HandleFunc("/page/{url}/{key}", pageEditor)
	r.HandleFunc("/page/{url}/{key}/update", updatePage)
	r.HandleFunc("/page/{url}/{key}/delete", deletePage)
	r.HandleFunc("/admin/new", newPage)
	r.HandleFunc("/login", login)
	r.HandleFunc("/verify", verify)
	r.HandleFunc("/check", checkLogin)

	statichandler := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static", statichandler))

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
