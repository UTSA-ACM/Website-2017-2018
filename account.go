package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var (
	keyManager userKeyManager
)

type userKeyManager struct {
	lock   sync.RWMutex
	keyMap map[string]*userKey
}

type userKey struct {
	Key         string
	Valid       bool
	Created     int
	Activated   bool
	ActivatedBy string
	ActivatedOn int
}

func init() {
	keyManager.keyMap = make(map[string]*userKey)
}

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

	if getRowCount() <= (pageID+1)*10 {
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

func changeVisible(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

	vars := mux.Vars(r)

	url := vars["url"]

	vt, err := strconv.ParseInt(vars["visible"], 10, 64)

	if err != nil {
		log.Print(err)
		return
	}

	visible := int(vt)

	setVisible(url, visible)
}

func accountManagement(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

	renderStatic(w, "account.html")
}

func generateAccountLink(w http.ResponseWriter, r *http.Request) {

	checkLogin(w, r)

	b := make([]byte, 24)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	generatedKey := string(b)

	var keyEntry userKey

	keyEntry.Key = generatedKey
	keyEntry.Valid = true
	keyEntry.Created = int(time.Now().Unix())
	keyEntry.Activated = false
	keyEntry.ActivatedBy = ""
	keyEntry.ActivatedOn = -1

	keyManager.lock.Lock()

	keyManager.keyMap[generatedKey] = &keyEntry

	keyManager.lock.Unlock()

	// Send successful response with generated account key as data payload
	ajaxResponse(w, r, true, generatedKey, "")

}

func createAccountPage(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	key := vars["key"]

	keyEntry, ok := keyManager.keyMap[key]

	if !ok || !keyEntry.Valid {
		http.Redirect(w, r, "/", 302)
		return
	}

	t, err := template.ParseFiles("./templates/createaccount.html")

	if err != nil {
		log.Print(err)
	}

	t.Execute(w, struct{}{})

}

func activateAccountLink(w http.ResponseWriter, r *http.Request) {

	/*
		Takes form parameters, activates the given key
		username:
		password:
	*/
	vars := mux.Vars(r)

	activatedKey := vars["key"]
	formUsername := r.PostFormValue("username")
	formPassword := r.PostFormValue("password")

	keyManager.lock.Lock()

	keyEntry, ok := keyManager.keyMap[activatedKey]

	if !ok || !keyEntry.Valid {
		ajaxResponse(w, r, false, "", "Invalid account creation key")
		return
	}

	if len(formUsername) < 2 {
		ajaxResponse(w, r, false, "", "Username must be at least 2 characters long")
	}

	if insertUser(formUsername, formPassword) < 0 {
		ajaxResponse(w, r, false, "", "Internal error while activating account")
		return
	}

	keyEntry.ActivatedOn = int(time.Now().Unix())
	keyEntry.ActivatedBy = formUsername
	keyEntry.Activated = true
	keyEntry.Valid = false

	keyManager.lock.Unlock()

	ajaxResponse(w, r, true, "", "")

}

func getAccountKeys(w http.ResponseWriter, r *http.Request) {

	checkLogin(w, r)

	keyManager.lock.RLock()

	var keys []string

	for k := range keyManager.keyMap {
		if keyManager.keyMap[k].Valid {
			keys = append(keys, k)
		}
	}

	keyManager.lock.RUnlock()

	ajaxResponse(w, r, true, keys, "")

}

func newPassword(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

	username := getUsername(r)

	old := r.PostFormValue("old")
	new := r.PostFormValue("new")

	if updatePassword(username, old, new) {
		fmt.Fprint(w, "Password Changed")
		return
	}

	fmt.Fprint(w, "Change Failed")
	return

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
