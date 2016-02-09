package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/AlexanderThaller/httphelper"
	log "github.com/Sirupsen/logrus"
	"github.com/juju/errgo"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	Stopping bool
)

var (
	FlagBindingGallery string
	FlagBindingMetrics string

	FlagFolderGallery string
	FlagFolderCache   string

	FlagLogLevel string
	FlagLogFile  string
)

func init() {
	// Binding
	flag.StringVar(&FlagBindingGallery, "binding.gallery", ":6112",
		"the network binding of the gallery")
	flag.StringVar(&FlagBindingMetrics, "binding.metrics", ":6113",
		"the network binding of the metrics")

	// Folder
	flag.StringVar(&FlagFolderGallery, "folder.gallery", "gallery",
		"the folder from which to serve the gallery")
	// Folder
	flag.StringVar(&FlagFolderCache, "folder.cache", ".cache",
		"the folder from which to serve the gallery")

	// Log
	flag.StringVar(&FlagLogLevel, "log.level", "info",
		"the loglevel of the application (debug, info, warning, error")
	flag.StringVar(&FlagLogFile, "log.file", "",
		"the path to the logfile. if empty logs will go to stdout and not a logfile")

	flag.Parse()

	level, err := log.ParseLevel(FlagLogLevel)
	if err != nil {
		log.Fatal(errgo.Notef(err, "can not parse loglevel from flag"))
	}
	log.SetLevel(level)

	if FlagLogFile != "" {
		logfile, err := os.OpenFile(FlagLogFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			log.Fatal(errgo.Notef(err, "can not open logfile for writing"))
		}
		defer logfile.Close()

		log.SetOutput(logfile)
	}
}

func main() {
	log.Info("Starting")

	go func() {
		router := httprouter.New()

		// Router handler
		router.MethodNotAllowed = httphelper.HandlerLoggerHTTP(httphelper.PageRouterMethodNotAllowed)
		router.NotFound = httphelper.HandlerLoggerHTTP(httphelper.PageRouterNotFound)

		// Root and Favicon
		router.GET("/", httphelper.HandlerLoggerRouter(pageRoot))
		router.GET("/favicon.ico", httphelper.HandlerLoggerRouter(httphelper.PageMinimalFavicon))

		// Gallery
		router.GET("/gallery/*path", httphelper.HandlerLoggerRouter(pageGallery))

		log.Info("Start serving files on ", FlagBindingGallery)
		log.Fatal(http.ListenAndServe(FlagBindingGallery, router))
	}()

	go func() {
		if FlagBindingMetrics != "" {
			log.Info("Starting Metrics", FlagBindingMetrics)
			http.Handle("/metrics", prometheus.Handler())
			go http.ListenAndServe(FlagBindingMetrics, nil)
		}
	}()

	log.Debug("Waiting for interrupt signal")
	httphelper.WaitForStopSignal()
	log.Info("Stopping")
}
