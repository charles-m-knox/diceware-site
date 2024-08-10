package main

import (
	"log"
	"net/http"
	"time"

	dice "git.cmcode.dev/cmcode/go-dicewarelib"
)

// Receives and routes requests.
//
// Wrap router like this so that it can be used with other middleware:
//
//	http.HandlerFunc(router)
func router(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/gen":
		getGenPassword(w, r)
	case "/":
		getIndex(w, r)
	case "/css.css":
		getStyles(w, r)
	default:
		// redirect to the home page
		w.WriteHeader(http.StatusTemporaryRedirect)
		_, _ = w.Write([]byte("/"))
	}
}

func getStyles(w http.ResponseWriter, r *http.Request) {
	// getStyles is technically a static asset but it isn't served via the
	// static filesystem, so it isn't passed through the cache middleware.
	w.Header().Set("Cache-Control", "private, max-age=604800")
	_, err := w.Write(styles)
	if err != nil {
		log.Printf("failed to write styles http response: %v", err.Error())
	}
}

func getGenPassword(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	renderGenPassword(w, r, Words)
	finishedTime := time.Now()
	elapsed := finishedTime.Sub(startTime)
	// enforce a minimum response time of ~30ms
	if elapsed.Milliseconds() < 30 {
		randSleep := time.Duration(30+dice.GetRandomInt(50)) * time.Millisecond
		time.Sleep(randSleep - elapsed)
	}
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	renderIndex(w, r, Words, Index)
	finishedTime := time.Now()
	elapsed := finishedTime.Sub(startTime)
	// enforce a minimum response time of ~30ms
	if elapsed.Milliseconds() < 30 {
		randSleep := time.Duration(30+dice.GetRandomInt(50)) * time.Millisecond
		time.Sleep(randSleep - elapsed)
	}
}
