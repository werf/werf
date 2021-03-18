package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"
)

func newRouter() *mux.Router {
	r := mux.NewRouter()

	staticFileDirectoryMain := http.Dir("./root/main")
	staticFileDirectoryRu := http.Dir("./root/ru")

	var ruHostMatch mux.MatcherFunc = func(r *http.Request, rm *mux.RouteMatch) bool {
		result := false
		result, _ = regexp.MatchString("^ru\\..*(.+\\.flant\\.com|werf\\.io)$", r.Host)
		return result
	}

	r.PathPrefix("/status").HandlerFunc(statusHandler)
	r.PathPrefix("/backend/").HandlerFunc(ssiHandler)
	r.PathPrefix("/documentation/v{group:[0-9]+.[0-9]+}-{channel:alpha|beta|ea|stable|rock-solid}").HandlerFunc(groupChannelHandler)
	r.PathPrefix("/documentation/v{group:[0-9]+.[0-9]+}").HandlerFunc(groupHandler)
	r.PathPrefix("/documentation").HandlerFunc(rootDocHandler)
	r.PathPrefix("/health").HandlerFunc(healthCheckHandler)
	r.Path("/includes/topnav.html").HandlerFunc(topnavHandler)
	r.Path("/includes/version-menu.html").HandlerFunc(topnavHandler)
	r.Path("/includes/group-menu.html").HandlerFunc(groupMenuHandler)
	r.Path("/includes/channel-menu.html").HandlerFunc(channelMenuHandler)
	r.Path("/404.html").HandlerFunc(notFoundHandler)
	// Ru static
	r.MatcherFunc(ruHostMatch).Handler(serveFilesHandler(staticFileDirectoryRu))
	// Other (En) static
	r.PathPrefix("/").Handler(serveFilesHandler(staticFileDirectoryMain))

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

	log.Infoln(fmt.Sprintf("Started with LOG_LEVEL %s", logLevel))
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
