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

	dashboardTemplate, err := template.ParseFiles("templates/index.html", "templates/nav.html")

	if err != nil {
		log.Fatal(err)
	}

	var posts []Page

	postCount := 10

	posts, _ = getPagesSortedByDate(pageID, postCount, true)

	next := pageID + 1

	if getVisibleRowCount() <= next*10 {
		next = pageID
	}

	prev := pageID - 1

	if pageID == 0 {
		prev = 0
	}

	data := struct {
		Page  int
		Next  int
		Prev  int
		Posts []Page
	}{
		pageID,
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
	r.HandleFunc("/pages/{url}", getPage)
	r.HandleFunc("/pages/{url}/{key}", pageEditor)
	r.HandleFunc("/pages/{url}/{key}/update", updatePage)
	r.HandleFunc("/pages/{url}/{key}/delete", deletePage)
	r.HandleFunc("/page/{url}/{key}/upload", receiveEditorFile)
	r.HandleFunc("/admin/{url}/rekey", reKey)
	r.HandleFunc("/admin/{url}/visibility/{visible}", changeVisible)
	r.HandleFunc("/admin/new", createPage)
	r.HandleFunc("/admin/users/active-keys", getAccountKeys)
	r.HandleFunc("/admin/users/actions/generate-account", generateAccountLink)
	r.HandleFunc("/admin/users/actions/create-account/{key}", createAccountPage)
	r.HandleFunc("/admin/users/actions/activate-account/{key}", activateAccountLink)
	r.HandleFunc("/admin/users/actions/deactivate/{key}", deactivateAccountLink)
	r.HandleFunc("/admin/account", accountManagement)
	r.HandleFunc("/admin/password", newPassword)
	r.HandleFunc("/admin/files", fileManagement)
	r.HandleFunc("/admin/files/new", receiveFile)
	r.HandleFunc("/admin/files/list", listFiles)
	r.HandleFunc("/admin/files/delete", deleteFile)
	r.HandleFunc("/admin/files/resize", imageResize)
	r.HandleFunc("/login", login)
	r.HandleFunc("/logout", logout)
	r.HandleFunc("/verify", verify)
	r.HandleFunc("/check", checkLogin)

	statichandler := http.FileServer(http.Dir("./static/"))
	fileshandler := http.FileServer(http.Dir("./files/"))
	http.Handle("/static/", http.StripPrefix("/static", statichandler))
	http.Handle("/files/", http.StripPrefix("/files", fileshandler))

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
