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

func markdownPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	md := getMarkdown(vars["url"])

	if md.Target != "" {
		http.Redirect(w, r, md.Target, 302)
	}

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

	http.Redirect(w, r, "/pages/"+newURL+"/"+md.Key, 302)

}

func reKey(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

	vars := mux.Vars(r)

	md := getMarkdown(vars["url"])

	md.Key = generateKey()

	updateMarkdown(vars["url"], &md)

	http.Redirect(w, r, "/admin", 302)
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

	checkLogin(w, r)
	if r.Method != "POST" {
		http.Redirect(w, r, "/admin", 302)
		return
	}

	title := r.PostFormValue("title")

	if title == "" {
		http.Redirect(w, r, "/admin", 302)
		return
	}

	md := newMarkdown(title, "", "", "", "", "")

	if insertMarkdown(md) == -1 {
		log.Print("Insert Failed")
		http.Redirect(w, r, "/admin", 302)
		return
	}

	http.Redirect(w, r, "/pages/"+md.URL+"/"+md.Key, 302)

}

func dashboard(w http.ResponseWriter, r *http.Request) {

	checkLogin(w, r)

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

	dashboardTemplate, err := template.ParseFiles("templates/dashboard.html")

	if err != nil {
		log.Fatal(err)
	}

	var posts []Markdown

	posts, _ = getPostsSortedByDate(page*10, 10, false, false)

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
