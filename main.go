package main

import (
	"embed"
	"flag"
	"html/template"
	"log"
	"net/http"
	"time"

	dice "git.cmcode.dev/cmcode/go-dicewarelib"
)

//go:embed words-complex.txt
//go:embed words-simple.txt
//go:embed index.html
var content embed.FS

//go:embed static/styles.css
//go:embed static/i.svg
//go:embed static/robots.txt
var static embed.FS

var (
	Words  *dice.Words
	Index  *template.Template
	styles []byte

	flagAddr  string
	flagCert  string
	flagKey   string
	flagExtra bool
)

func parseFlags() {
	flag.StringVar(&flagAddr, "addr", "0.0.0.0:29102", "the address (host and port) to listen on")
	flag.StringVar(&flagCert, "cert", "", "the cert.pem file to use for TLS - leave blank for no TLS")
	flag.StringVar(&flagKey, "key", "", "the key.pem file to use for TLS - leave blank for no TLS")
	flag.BoolVar(&flagExtra, "x", false, "whether to load the extra word dictionary or not (uses about 30-40MB RAM)")
	flag.Parse()
}

func main() {
	var err error

	parseFlags()
	loadResources()

	fs := http.FileServerFS(static)
	http.Handle("/", compressionHandler(http.HandlerFunc(router)))
	http.Handle("/static/", compressionHandler(cacheHandler(fs)))

	srv := &http.Server{
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 3 * time.Second,
		Addr:         flagAddr,
	}

	if flagCert == "" || flagKey == "" {
		log.Printf("listening on %v", srv.Addr)
		err = srv.ListenAndServe()
	} else {
		log.Printf("listening on %v with TLS", srv.Addr)
		err = srv.ListenAndServeTLS(flagCert, flagKey)
	}

	if err != nil {
		log.Fatalf("failed to run web server: %v", err.Error())
	}
}
