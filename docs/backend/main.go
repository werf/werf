package main

import (
	"context"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

func newRouter() *mux.Router {
	r := mux.NewRouter()

	staticFileDirectoryMain := http.Dir("./root/main")
	staticFileDirectoryRu := http.Dir("./root/ru")

	r.PathPrefix("/status").HandlerFunc(statusHandler)
	r.PathPrefix("/backend/").HandlerFunc(ssiHandler)
	r.PathPrefix("/v{group:[0-9]+.[0-9]+}-{channel:alpha|beta|ea|stable|rock-solid}").HandlerFunc(groupChannelHandler)
	r.PathPrefix("/documentation").HandlerFunc(rootDocumentationHandler)
	r.PathPrefix("/health").HandlerFunc(healthCheckHandler)
	r.Path("/includes/topnav.html").HandlerFunc(topnavHandler)
	r.Path("/includes/version-menu.html").HandlerFunc(topnavHandler)
	r.Path("/404.html").HandlerFunc(notFoundHandler)
	// En static
	r.PathPrefix("/").Host("werf.io").Handler(serveFilesHandler(staticFileDirectoryMain))
	r.PathPrefix("/").Host("www.werf.io").Handler(serveFilesHandler(staticFileDirectoryMain))
	r.PathPrefix("/").Host("ng.werf.io").Handler(serveFilesHandler(staticFileDirectoryMain))
	r.PathPrefix("/").Host("werf.test.flant.com").Handler(serveFilesHandler(staticFileDirectoryMain))
	r.PathPrefix("/").Host("werfng.test.flant.com").Handler(serveFilesHandler(staticFileDirectoryMain))
	// Ru static
	r.PathPrefix("/").Host("ru.werf.io").Handler(serveFilesHandler(staticFileDirectoryRu))
	r.PathPrefix("/").Host("ru.ng.werf.io").Handler(serveFilesHandler(staticFileDirectoryRu))
	r.PathPrefix("/").Host("ru.werf.test.flant.com").Handler(serveFilesHandler(staticFileDirectoryRu))
	r.PathPrefix("/").Host("ru.werfng.test.flant.com").Handler(serveFilesHandler(staticFileDirectoryRu))

	r.Use(LoggingMiddleware)

	r.NotFoundHandler = r.NewRoute().HandlerFunc(notFoundHandler).GetHandler()

	return r
}

func main() {
	var wait time.Duration
	logLevel := os.Getenv("LOG_LEVEL")
	if strings.ToLower(logLevel) == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if strings.ToLower(logLevel) == "trace" {
		log.SetLevel(log.TraceLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	//log.SetFormatter(&log.JSONFormatter{})
	//log.SetFlags(log.Ldate | log.Ltime)
	log.Infoln("Started")
	r := newRouter()

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		err := srv.ListenAndServe()
		if err == http.ErrServerClosed {
			err = nil
		}
		if err != nil {
			log.Errorln(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	log.Infoln("Shutting down")
	os.Exit(0)
}
