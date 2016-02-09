package main

import (
	"net/http"

	"github.com/AlexanderThaller/httphelper"
	"github.com/juju/errgo"
	"github.com/julienschmidt/httprouter"
)

func pageRoot(w http.ResponseWriter, r *http.Request, p httprouter.Params) *httphelper.HandlerError {
	return httphelper.NewHandlerError(errgo.New("not implemented"), http.StatusInternalServerError)
}

func pageGallery(w http.ResponseWriter, r *http.Request, p httprouter.Params) *httphelper.HandlerError {
	return httphelper.NewHandlerError(errgo.New("not implemented"), http.StatusInternalServerError)
}
