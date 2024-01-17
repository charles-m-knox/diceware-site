package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	consts "gitea.cmcode.dev/cmcode/diceware-site/constants"
	"gitea.cmcode.dev/cmcode/diceware-site/renderers"
	"gitea.cmcode.dev/cmcode/diceware-site/utils"
)

func getAlpine(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add(consts.ContentEncodingHeader, consts.ContentEncodingGzipHeaderValue)
	w.Header().Add(consts.ContentTypeHeader, "application/javascript")
	w.Header().Add(consts.CacheControlHeader, consts.CacheControlHeaderValue)

	_, err := w.Write(Alpinejs)
	if err != nil {
		log.Printf("getAlpine write error: %v", err.Error())

		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getSemantic(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add(consts.ContentEncodingHeader, consts.ContentEncodingGzipHeaderValue)
	w.Header().Add(consts.ContentTypeHeader, "text/css")
	w.Header().Add(consts.CacheControlHeader, consts.CacheControlHeaderValue)

	_, err := w.Write(Semantic)
	if err != nil {
		log.Printf("getSemantic write error: %v", err.Error())

		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getStyles(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add(consts.ContentEncodingHeader, consts.ContentEncodingGzipHeaderValue)
	w.Header().Add(consts.ContentTypeHeader, "text/css")
	w.Header().Add(consts.CacheControlHeader, consts.CacheControlHeaderValue)

	_, err := w.Write(Stylesgz)
	if err != nil {
		log.Printf("getStyles write error: %v", err.Error())

		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getFavicon(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add(consts.ContentEncodingHeader, consts.ContentEncodingGzipHeaderValue)
	w.Header().Add(consts.ContentTypeHeader, "image/svg+xml")
	w.Header().Add(consts.CacheControlHeader, consts.CacheControlHeaderValue)

	_, err := w.Write(Favicongz)
	if err != nil {
		log.Printf("getFavicon write error: %v", err.Error())

		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getRobotsTxt(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write(consts.RobotsTxt)
	if err != nil {
		log.Printf("getRobotsTxt err: %v", err.Error())
	}
}

func getHealthCheck(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write(consts.HealthcheckResponse)
	if err != nil {
		log.Printf("getHealthCheck err: %v", err.Error())
	}
}

func getGenPassword(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	renderers.GenPassword(w, r, Words)

	finishedTime := time.Now()

	elapsed := finishedTime.Sub(startTime)

	// enforce a minimum response time of ~30ms
	if elapsed.Milliseconds() < 30 {
		randSleep := time.Duration(30+utils.GetRandomInt(50)) * time.Millisecond
		time.Sleep(randSleep - elapsed)
	}
}

func getIndex(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	renderers.Index(w, r, Words, Index)

	finishedTime := time.Now()

	elapsed := finishedTime.Sub(startTime)

	// enforce a minimum response time of ~30ms
	if elapsed.Milliseconds() < 30 {
		randSleep := time.Duration(30+utils.GetRandomInt(50)) * time.Millisecond
		time.Sleep(randSleep - elapsed)
	}
}

func getFonts(w http.ResponseWriter, r *http.Request) {
	// note: gzipping the fonts makes no difference on the resulting size
	fontTarget := r.URL.Query().Get("font")

	w.Header().Add(consts.ContentTypeHeader, "application/font-woff")

	if strings.LastIndex(fontTarget, "woff2") == len(fontTarget)-1 {
		w.Header().Add(consts.ContentTypeHeader, "application/font-woff")
	}

	w.Header().Add(consts.CacheControlHeader, consts.CacheControlHeaderValue)

	cachedFont, ok := FontCache[fontTarget]
	if !ok {
		log.Printf("no cached font for %v %v", cachedFont, fontTarget)

		var err error

		FontCache[fontTarget], err = getFromEmbed(
			content,
			fmt.Sprintf("%v/%v", consts.PathFontsLocal, fontTarget),
		)

		if err != nil {
			log.Fatalf("failed to load from %v: %v", consts.PathStyles, err.Error())
		}

		cachedFont = FontCache[fontTarget]
	}

	_, err := w.Write(cachedFont)
	if err != nil {
		log.Printf("getFonts write error: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
