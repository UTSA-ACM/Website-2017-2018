package main

import (
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

	"image/png"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
)

type fileURL struct {
	Name       string
	Resizeable bool
}

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
	checkLogin(w, r)

	files, err := ioutil.ReadDir("./files")

	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "File listing failed")
	}

	var fileURLs []fileURL

	for _, name := range files {

		var item fileURL
		item.Name = name.Name()

		if strings.HasSuffix(item.Name, ".png") || strings.HasSuffix(item.Name, ".jpg") {
			item.Resizeable = true
		}

		fileURLs = append(fileURLs, item)
	}

	t, err := template.ParseFiles("templates/filemanager.html", "templates/nav.html")

	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "File listing failed")
	}

	err = t.Execute(w, fileURLs)
	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "File listing failed")
	}
}

func imageResize(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	name := vars["name"]

	if !(strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".jpg")) {
		log.Print("Not a valid ending", name)
		ajaxResponse(w, r, false, nil, "Resize failed: Not a PNG or JPG")
		return
	}

	if strings.Contains(name, "..") {
		log.Print("Tried to exit folder")
		ajaxResponse(w, r, false, nil, "Resize failed: Not a PNG or JPG")
		return
	}

	file, err := os.Open("./files/" + name)

	if err != nil {
		log.Print(err)
		ajaxResponse(w, r, false, nil, "Resize failed: Not a PNG or JPG")
		return
	}

	image, _, err := image.Decode(file)

	if err != nil {
		log.Print(err)
		ajaxResponse(w, r, false, nil, "Resize failed: Not a PNG or JPG")
		return
	}

	width, err := strconv.ParseInt(vars["width"], 10, 64)

	if err != nil {
		log.Print("Width not valid")
		ajaxResponse(w, r, false, nil, "Resize failed: Not a PNG or JPG")
		return
	}

	height, err := strconv.ParseInt(vars["height"], 10, 64)

	if err != nil {
		log.Print("Height not valid")
		ajaxResponse(w, r, false, nil, "Resize failed: Not a PNG or JPG")
		return
	}

	newImage := resize.Resize(uint(width), uint(height), image, resize.Lanczos2)

	var outFile *os.File
	var newName string

	if strings.HasSuffix(name, ".png") {

		newName = vars["newName"] + ".png"

		outFile, err = os.Create(newName)

		if err != nil {
			log.Print(err)
			ajaxResponse(w, r, false, nil, "Resize failed: Not a PNG or JPG")
			return
		}

		err = png.Encode(outFile, newImage)

	} else if strings.HasSuffix(name, ".jpg") {

		newName = vars["newName"] + ".png"

		outFile, err = os.Create(newName)

		if err != nil {
			log.Print(err)
			ajaxResponse(w, r, false, nil, "Resize failed: Not a PNG or JPG")
			return
		}

		err = png.Encode(outFile, newImage)

	}

	if err != nil {
		log.Print(err)
		ajaxResponse(w, r, false, nil, "Resize failed: Not a PNG or JPG")
		return
	}

	ajaxResponse(w, r, true, "/files/"+newName, "")

}
