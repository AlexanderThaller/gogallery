package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/AlexanderThaller/httphelper"
	"github.com/Unknwon/log"
	"github.com/juju/errgo"
	"github.com/julienschmidt/httprouter"
)

func pageRoot(w http.ResponseWriter, r *http.Request, p httprouter.Params) *httphelper.HandlerError {
	http.Redirect(w, r, "/gallery", http.StatusMovedPermanently)
	return nil
}

func pageGallery(w http.ResponseWriter, r *http.Request, p httprouter.Params) *httphelper.HandlerError {
	filepath := path.Join(FlagFolder, p.ByName("path"))
	log.Debug("Sending ", filepath)

	stat, err := os.Stat(filepath)
	if err != nil {
		return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not stat file"))
	}

	if stat.Mode().IsDir() {
		log.Debug("Filetype: Directory")
		return pageFilesDirectory(w, r, p)
	}

	if stat.Mode().IsRegular() {
		log.Debug("Filetype: Regular")
		return pageFilesRegular(w, r, p)
	}

	if !stat.Mode().IsDir() && !stat.Mode().IsRegular() {
		return httphelper.NewHandlerErrorDef(errgo.New("filetype is not a directory and not a regular file. Something is strange."))
	}

	return httphelper.NewHandlerErrorDef(errgo.New("unreachable code reached!"))
}

func pageFilesDirectory(w http.ResponseWriter, r *http.Request, p httprouter.Params) *httphelper.HandlerError {
	filepath := path.Join(FlagFolder, p.ByName("path"))
	files, err := ioutil.ReadDir(filepath)
	if err != nil {
		return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not read from directory"))
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
  <html lang="en">
  <head>
  <meta charset="utf-8">
  <title>Filehasher - `+filepath+`</title>
  </head>
  <body>`)
	for _, file := range files {
		filepath := path.Join(r.URL.Path, file.Name())
		if file.IsDir() {
			fmt.Fprintf(w, "Link: <a href="+filepath+">"+file.Name()+"</a> <b>[d]</b>")
			fmt.Fprintf(w, "<br>\n")
			continue
		}

		fmt.Fprintf(w, "Link: <a href="+filepath+">"+file.Name()+"</a>")
		fmt.Fprintf(w, "<br>\n")
	}
	fmt.Fprintf(w, `</body>
  </html>`)

	return nil
}

func pageFilesRegular(w http.ResponseWriter, r *http.Request, p httprouter.Params) *httphelper.HandlerError {
	filepath := path.Join(FlagFolder, p.ByName("path"))

	file, err := os.Open(filepath)
	if err != nil {
		return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not open file for reading"))
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not get file information"))
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%v", info.Size()))

	_, err = io.Copy(w, file)
	if err != nil {
		return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not copy file to response writer"))
	}

	return nil
}
