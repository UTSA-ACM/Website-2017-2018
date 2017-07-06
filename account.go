package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func dashboard(w http.ResponseWriter, r *http.Request) {

	checkLogin(w, r)
	username := getUsername(r)

	pageID := 0

	qpage := r.URL.Query().Get("page")

	if qpage != "" {
		var err error
		tpage, err := strconv.ParseInt(qpage, 0, 64)

		if err != nil {
			log.Print(err)
			http.Redirect(w, r, "/admin", 302)
			return
		}
		pageID = int(tpage)
	}

	dashboardTemplate, err := template.ParseFiles("templates/dashboard.html", "templates/nav.html")

	if err != nil {
		log.Fatal(err)
	}

	var posts []Page

	posts, _ = getPagesSortedByDate(pageID, 10, false)

	next := pageID + 1

	if getLastID() <= next*10 {
		next = pageID
	}

	prev := pageID - 1

	if pageID == 0 {
		prev = 0
	}

	data := struct {
		Username string
		Page     int
		Next     int
		Prev     int
		Posts    []Page
	}{
		username,
		pageID,
		next,
		prev,
		posts}

	dashboardTemplate.Execute(w, data)

}

func accountManagement(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

	renderStatic(w, "account.html")
}

func newPassword(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

	username := getUsername(r)

	old := r.PostFormValue("old")
	new := r.PostFormValue("new")

	if updatePassword(username, old, new) {
		fmt.Fprint(w, "Password Changed")
		return
	} else {
		fmt.Fprint(w, "Change Failed")
		return
	}

}

func login(w http.ResponseWriter, r *http.Request) {

	if loggedIn(r) {
		http.Redirect(w, r, "/admin", 302)
	}

	t, err := template.ParseFiles("./templates/login.html")

	if err != nil {
		log.Fatal(err)
	}

	t.Execute(w, r.FormValue("continue"))

}

func logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{Name: "token", Value: "", MaxAge: 1}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", 302)
}

func verify(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	password := r.PostFormValue("password")

	cont := r.URL.Query().Get("continue")
	if cont == "" {
		cont = "/admin"
	}

	if checkUser(name, password) {

		token := newSession(name)

		cookie := http.Cookie{Name: "token", Value: token, MaxAge: 259200}
		http.SetCookie(w, &cookie)
		//fmt.Fprint(w, "cookie should be made")
		http.Redirect(w, r, cont, 302)

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

		redirectURL := "/login?continue=" + r.RequestURI
		http.Redirect(w, r, redirectURL, 302)
		return

	}

	token := cookie.Value

	username := checkSession(token)

	if username == "" {

		redirectURL := "/login?continue=" + r.RequestURI
		http.Redirect(w, r, redirectURL, 302)
		return

	}

}

func getUsername(r *http.Request) string {

	cookie, err := r.Cookie("token")

	if err != nil {
		return ""
	}

	token := cookie.Value

	username := checkSession(token)

	return username
}
