package main

import (
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
)

func main() {
	log.Info("Starting")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	log.Debug("Waiting for interrupt signal")
	<-signalChan
	log.Info("Stopping")
}
