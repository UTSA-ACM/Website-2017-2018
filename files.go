package main

import (
	"errors"
	"fmt"
	"html/template"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"strings"

	"image/jpeg"
	"image/png"

	"strconv"

	"github.com/nfnt/resize"
)

type fileURL struct {
	Name       string
	Resizeable bool
}

func receiveFile(w http.ResponseWriter, r *http.Request) {

	checkLogin(w, r)

	file, header, err := r.FormFile("file")

	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/admin/files", 302)
	}
	defer file.Close()

	disk, err := os.OpenFile("files/"+header.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	defer disk.Close()

	if err != nil {
		log.Print(err)
		http.Redirect(w, r, "/admin/files", 302)
	}

	io.Copy(disk, file)

	//ajaxResponse(w, r, true, "/files/"+header.Filename, "")
	http.Redirect(w, r, "/admin/files", 302)

}

func fileManagement(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

	files, err := ioutil.ReadDir("./files")

	if err != nil {
		log.Print(err)
		fmt.Fprint(w, "File listing failed")
		return
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

func listFiles(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

	files, err := ioutil.ReadDir("./files")

	if err != nil {
		log.Print(err)
		ajaxResponse(w, r, false, "", "Directory read failure")
		return
	}

	var fileNames []string

	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	ajaxResponse(w, r, true, fileNames, "")

}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

	fileName := r.PostFormValue("filename")

	if strings.Contains(fileName, "..") {
		log.Print("Tried to exit folder")
		ajaxResponse(w, r, false, "", "Relative paths not allowed")
		return
	}

	err := os.Remove("./files/" + fileName)

	if err != nil {
		log.Print(err)
		ajaxResponse(w, r, false, "", "Removal failed")
	}

	ajaxResponse(w, r, true, "", "")

}

func imageResize(w http.ResponseWriter, r *http.Request) {
	checkLogin(w, r)

	name := r.PostFormValue("filename")

	newName := r.PostFormValue("newname")

	ratio, err := strconv.ParseFloat(r.PostFormValue("ratio"), 64)

	if err != nil {
		log.Print(err)
		ajaxResponse(w, r, false, nil, "Ratio not a float")
	}

	url, err := resizeImage(name, newName, ratio)

	if err != nil {
		log.Print(err)
		ajaxResponse(w, r, false, "", err.Error())
	}

	ajaxResponse(w, r, true, url, "")

}

func resizeImage(name, newname string, ratio float64) (string, error) {

	if ratio <= 0 {
		return "", errors.New("Resize failed: Ratio cannot be less than 0")
	}

	if !(strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".jpg")) {
		log.Print("Not a valid ending", name)
		return "", errors.New("Resize failed: Not a PNG or JPG")
	}

	if strings.Contains(name, "..") {
		log.Print("Tried to exit folder")

		return "", errors.New("Resize failed: Relative paths not allowed")
	}

	file, err := os.Open("./files/" + name)
	defer file.Close()

	if err != nil {
		log.Print(err)

		return "", errors.New("Resize failed: File not found")
	}

	image, _, err := image.Decode(file)

	wh := image.Bounds().Size()

	width := int(float64(wh.X) * ratio)
	height := int(float64(wh.Y) * ratio)

	if width <= 0 || height <= 0 {
		return "", errors.New("Resize failed: Width or Height too small")
	}

	if err != nil {
		log.Print(err)
		return "", errors.New("Resize failed: Could not decode file")
	}

	newImage := resize.Resize(uint(width), uint(height), image, resize.Lanczos2)

	var outFile *os.File
	var newName string

	if strings.HasSuffix(name, ".png") {

		newName = "./files/" + newname + ".png"

		outFile, err = os.Create(newName)
		defer outFile.Close()

		if err != nil {
			log.Print(err)

			return "", errors.New("Resize failed: Could not create file")
		}

		err = png.Encode(outFile, newImage)

	} else if strings.HasSuffix(name, ".jpg") {

		newName = "./files/" + newname + ".jpg"

		outFile, err = os.Create(newName)
		defer outFile.Close()

		if err != nil {
			log.Print(err)

			return "", errors.New("Resize failed: Could not create file")
		}

		err = jpeg.Encode(outFile, newImage, nil)

	}

	if err != nil {
		log.Print(err)

		return "", errors.New("Resize failed: Could not encode image")
	}

	return newName, nil

}
