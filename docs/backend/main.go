package main

import (
	"context"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func newRouter() *mux.Router {
	r := mux.NewRouter()

	staticFileDirectoryMain := http.Dir("./root/main")
	staticFileDirectoryRu := http.Dir("./root/ru")

	r.PathPrefix("/status").HandlerFunc(statusHandler).Methods("GET")
	r.PathPrefix("/backend/").HandlerFunc(ssiHandler).Methods("GET")
	r.PathPrefix("/v{group:[0-9]+.[0-9]+}-{channel:alpha|beta|ea|stable|rock-solid}").HandlerFunc(groupChannelHandler).Methods("GET")
	r.PathPrefix("/documentation").HandlerFunc(rootDocumentationHandler).Methods("GET")
	r.PathPrefix("/health").HandlerFunc(healthCheckHandler).Methods("GET")
	r.Path("/includes/topnav.html").HandlerFunc(topnavHandler).Methods("GET")
	r.Path("/includes/version-menu.html").HandlerFunc(topnavHandler).Methods("GET")
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
	log.SetFlags(log.Ldate | log.Ltime)
	log.Println("Started")
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
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Shutting down")
	os.Exit(0)
}
