package main

import (
	"bytes"
	"fmt"
	"html/template"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"strings"

	"strconv"

	"time"

	"image/png"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
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

func imageResize(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name := vars["name"]
	fmt.Print(name)

	if !(strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".jpg")) {
		log.Print("Not a valid ending", name)
		http.Redirect(w, r, "/files/"+name, 302)
		return
	}

	if strings.Contains(name, "..") {
		log.Print("Tried to exit folder")
		http.Redirect(w, r, "/files/"+name, 302)
		return
	}

	file, err := os.Open("./files/" + name)

	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/files/"+name, 302)
		return
	}

	image, _, err := image.Decode(file)

	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/files/"+name, 302)
		return
	}

	width, err := strconv.ParseInt(vars["width"], 10, 64)

	if err != nil {
		log.Print("Width not valid")
		http.Redirect(w, r, "/files/"+name, 302)
		return
	}

	height, err := strconv.ParseInt(vars["height"], 10, 64)

	if err != nil {
		log.Print("Height not valid")
		http.Redirect(w, r, "/files/"+name, 302)
		return
	}

	newImage := resize.Resize(uint(width), uint(height), image, resize.Lanczos2)

	var buf bytes.Buffer

	if strings.HasSuffix(name, ".png") {
		err = png.Encode(&buf, newImage)
	} else if strings.HasSuffix(name, ".jpg") {
		err = png.Encode(&buf, newImage)
	}

	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/files/"+name, 302)
		return
	}

	reader := bytes.NewReader(buf.Bytes())

	http.ServeContent(w, r, name, time.Now(), reader)

}
