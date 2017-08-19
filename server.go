package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"text/template"
	"time"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"github.com/stripe/stripe-go/currency"

	"log"

	"strconv"

	mux "github.com/gorilla/mux"
)

func pageList(w http.ResponseWriter, r *http.Request) {
	pageID := 0

	qpage := r.URL.Query().Get("page")

	if qpage != "" {
		var err error
		tpage, err := strconv.ParseInt(qpage, 0, 64)

		if err != nil {
			log.Print(err)
			http.Redirect(w, r, "/", 302)
			return
		}
		pageID = int(tpage)
	}

	dashboardTemplate, err := template.ParseFiles("front-temp/pageList.html", "front-temp/nav.html", "front-temp/head.html")

	if err != nil {
		log.Fatal(err)
	}

	var posts []Page

	postCount := 10

	posts, _ = getPagesSortedByDate(pageID, postCount, true)

	for _, page := range posts {
		sanitizePage(&page)
	}

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

func dues(w http.ResponseWriter, r *http.Request) {
	duesTemplate, err := template.ParseFiles("front-temp/dues.html", "front-temp/nav.html", "front-temp/head.html")

	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/", 302)
		return
	}

	data := struct {
		PublishableKey string
	}{
		"pk_test_SQWtlmhc3dN8tDMllgP6VAiE",
	}

	err = duesTemplate.Execute(w, data)

	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/", 302)
		return
	}
}

func chargeDebit(token string, amount uint64, description string) *stripe.Charge {
	stripe.Key = "sk_test_7eht2ebDeXWThzZwU82rapko"

	params := &stripe.ChargeParams{
		Amount:   amount,
		Currency: currency.USD,
		Desc:     description,
	}

	params.SetSource(token)

	ch, err := charge.New(params)

	if err != nil {
		log.Fatalf("error while trying to charge a cc", err)
	}

	log.Printf("debit created successfully %v\n", ch.ID)

	return ch
}

func payDues(w http.ResponseWriter, r *http.Request) {
	fmt.Print(r.FormValue("stripeToken"))
	chargeDebit(r.FormValue("stripeToken"), 1500, "Testing charge")
	fmt.Fprint(w, "success")
}

func duesPaid(w http.ResponseWriter, r *http.Request) {

}

func main() {

	rand.Seed(time.Now().UTC().UnixNano())

	start()

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.HandleFunc("/", pageList)
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

	r.HandleFunc("/dues", dues)
	r.HandleFunc("/dues/pay", payDues)
	r.HandleFunc("/dues/paid", duesPaid)

	statichandler := http.FileServer(http.Dir("./static/"))
	frontStatichandler := http.FileServer(http.Dir("./front-static/"))
	fileshandler := http.FileServer(http.Dir("./files/"))
	http.Handle("/static/", http.StripPrefix("/static", statichandler))
	http.Handle("/front-static/", http.StripPrefix("/front-static", frontStatichandler))
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
