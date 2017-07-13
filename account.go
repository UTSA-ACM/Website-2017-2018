package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
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
