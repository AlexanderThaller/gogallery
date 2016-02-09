package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/AlexanderThaller/httphelper"
	"github.com/juju/errgo"
	"github.com/julienschmidt/httprouter"
	"github.com/nfnt/resize"
)

func pageRoot(w http.ResponseWriter, r *http.Request, p httprouter.Params) *httphelper.HandlerError {
	http.Redirect(w, r, "/gallery", http.StatusMovedPermanently)
	return nil
}

func pageGallery(w http.ResponseWriter, r *http.Request, p httprouter.Params) *httphelper.HandlerError {
	l := httphelper.NewHandlerLogEntry(r)

	filepath := path.Join(FlagFolderGallery, p.ByName("path"))
	l.Debug("Sending ", filepath)

	stat, err := os.Stat(filepath)
	if err != nil {
		return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not stat file"))
	}

	if stat.Mode().IsDir() {
		l.Debug("Filetype: Directory")
		return pageFilesDirectory(w, r, p)
	}

	if stat.Mode().IsRegular() {
		l.Debug("Filetype: Regular")
		return pageFilesRegular(w, r, p)
	}

	if !stat.Mode().IsDir() && !stat.Mode().IsRegular() {
		return httphelper.NewHandlerErrorDef(errgo.New("filetype is not a directory and not a regular file. Something is strange."))
	}

	return httphelper.NewHandlerErrorDef(errgo.New("unreachable code reached!"))
}

func pageFilesDirectory(w http.ResponseWriter, r *http.Request, p httprouter.Params) *httphelper.HandlerError {
	filepath := path.Join(FlagFolderGallery, p.ByName("path"))
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
	l := httphelper.NewHandlerLogEntry(r)

	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not parse values from query"))
	}
	width := values.Get("width")
	height := values.Get("height")

	l.Debug("width: ", width)
	l.Debug("height: ", height)

	if width != "" || height != "" {
		err := pageFilesRegularThumbnail(w, r, p)
		if err == nil {
			return nil
		}
		l.Warning(errgo.Notef(err.Error, "can not generate thumbnail for file"))
	}

	filepath := path.Join(FlagFolderGallery, p.ByName("path"))

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

func pageFilesRegularThumbnail(w http.ResponseWriter, r *http.Request, p httprouter.Params) *httphelper.HandlerError {
	l := httphelper.NewHandlerLogEntry(r)

	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not parse values from query"))
	}

	var width uint
	if values.Get("width") != "" {
		out, err := strconv.ParseUint(values.Get("width"), 10, 64)
		if err != nil {
			return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not parse width from parameters"))
		}
		width = uint(out)
	}

	var height uint
	if values.Get("height") != "" {
		out, err := strconv.ParseUint(values.Get("height"), 10, 64)
		if err != nil {
			return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not parse height from parameters"))
		}
		height = uint(out)
	}

	cachefile := filepath.Join(FlagFolderCache, p.ByName("path"), values.Get("width"), values.Get("height")+".jpg")
	if _, err := os.Stat(cachefile); os.IsNotExist(err) {
		l.Debug("Cachefile does not exist: ", cachefile)
	} else {
		l.Debug("Cachefile exists: ", cachefile)
		cache, err := os.Open(cachefile)
		if err != nil {
			l.Warning(errgo.Notef(err, "can not open cachefile from disk"))
		} else {
			_, err := io.Copy(w, cache)
			if err != nil {
				l.Warning(errgo.Notef(err, "can not copy cache file to response writer"))
			} else {
				l.Debug("Served from cachefile")
				w.Header().Set("Content-Type", "image/jpeg")
				return nil
			}
		}
	}

	pathfile := path.Join(FlagFolderGallery, p.ByName("path"))

	file, err := os.Open(pathfile)
	if err != nil {
		return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not open file from disk"))
	}
	defer file.Close()

	ext := filepath.Ext(pathfile)
	l.Debug("Filepath Extention: ", ext)

	var img image.Image
	switch ext {
	case ".jpeg", ".JPEG", ".jpg", ".JPG":
		img, err = jpeg.Decode(file)
		if err != nil {
			return httphelper.NewHandlerErrorDef(errgo.New("can not decode file as jpeg"))
		}
	case ".png", ".PNG":
		img, err = png.Decode(file)
		if err != nil {
			return httphelper.NewHandlerErrorDef(errgo.New("can not decode file as jpeg"))
		}

	default:
		return httphelper.NewHandlerErrorDef(errgo.New("dont know how to decode image with extention " + ext))
	}

	l.Debug("Width: ", width)
	l.Debug("Height: ", height)

	err = os.MkdirAll(filepath.Dir(cachefile), 0755)
	if err != nil {
		return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not create cache folder"))
	}

	cache, err := os.Create(cachefile)
	if err != nil {
		return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not open cachefile from disk"))
	}
	defer cache.Close()

	writer := io.MultiWriter(w, cache)

	thumbnail := resize.Thumbnail(width, height, img, resize.Lanczos3)
	err = jpeg.Encode(writer, thumbnail, nil)
	if err != nil {
		return httphelper.NewHandlerErrorDef(errgo.Notef(err, "can not encode image to jpeg"))
	}

	w.Header().Set("Content-Type", "image/jpeg")

	return nil
}
