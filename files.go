package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func receiveFile(w http.ResponseWriter, r *http.Request) {

	file, header, err := r.FormFile("file")

	if err != nil {
		log.Print(err)
		ajaxResponse(w, r, false, "", "File upload failed")
	}

	defer file.Close()

	disk, err := os.OpenFile("files/"+header.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	defer disk.Close()

	if err != nil {
		log.Print(err)
		ajaxResponse(w, r, false, "", "File upload failed")
	}

	io.Copy(disk, file)

	ajaxResponse(w, r, true, "/files/"+header.Filename, "")

}

func fileManagement(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("./files")

	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "File listing failed")
	}

	t, err := template.ParseFiles("templates/filemanager.html", "templates/nav.html")

	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "File listing failed")
	}

	err = t.Execute(w, files)
	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "File listing failed")
	}
}
