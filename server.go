package main

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"time"

	mux "github.com/gorilla/mux"
)

// TODO Create cookies when verifying login, and create a function to check whether that cookie is valid etc

func markdownPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	md := getMarkdown(vars["url"])
	renderMarkdown(w, md)
}

func insertMd(w http.ResponseWriter, r *http.Request) {

	md := newMarkdown("hello world and", "me", "no sum", "# no life\nmy life", "")

	insertMarkdown(md)
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
		http.Redirect(w, r, "/check", 302)

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
		fmt.Fprint(w, "Not logged in")
	} else {

		token := cookie.Value

		username := checkSession(token)

		if username == "" {
			fmt.Fprint(w, "Not logged in")
			return
		}

		fmt.Fprint(w, "Logged in as "+username)
	}

}

func main() {

	rand.Seed(time.Now().UTC().UnixNano())

	start()

	r := mux.NewRouter()
	r.HandleFunc("/page/{url}", markdownPage)
	r.HandleFunc("/c", insertMd)
	r.HandleFunc("/login", login)
	r.HandleFunc("/verify", verify)
	r.HandleFunc("/check", checkLogin)
	http.Handle("/", r)
	http.ListenAndServe(":80", nil)

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
