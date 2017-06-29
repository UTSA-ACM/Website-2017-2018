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

// TODO Create cookies when verifying login, and create a function to check whether that cookie is valid etc

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

func newPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	title := vars["title"]

	md := newMarkdown(title, "", "", "", "")

	http.Redirect(w, r, "/editor/"+md.URL, 302)

}

func dashboard(w http.ResponseWriter, r *http.Request) {

	// Check login status
	cookie, err := r.Cookie("token")

	if err != nil {
		// TODO: proper error handling (redirection?)
		http.Redirect(w, r, "/login", 302)
		return
	}

	token := cookie.Value
	username := checkSession(token)

	if username == "" {
		// TODO: proper error handling (redirection?)
		http.Redirect(w, r, "/login", 302)
		return
	}

	dashboardTemplate, err := template.ParseFiles("templates/dashboard.html")

	if err != nil {
		log.Fatal(err)
	}

	dashboardTemplate.Execute(w, username)

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

	http.Redirect(w, r, "/admin", 302)

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
	r.HandleFunc("/login", login)
	r.HandleFunc("/verify", verify)
	r.HandleFunc("/check", checkLogin)

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
