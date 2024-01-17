package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	consts "gitea.cmcode.dev/cmcode/diceware-site/constants"
	"gitea.cmcode.dev/cmcode/diceware-site/utils"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/svg"
)

//go:embed resources/words_simple.txt
//go:embed resources/words_alpha.txt
//go:embed static/alpinejs-3.12.0.min.js.gz
//go:embed static/semantic-2.9.2.min.css.gz
//go:embed static/styles.css
//go:embed templates/index.html
//go:embed static/fonts
//go:embed static/favicon.ico
var content embed.FS

var (
	Words     *utils.Words
	Index     *template.Template
	FontCache map[string][]byte

	Alpinejs  []byte
	Semantic  []byte
	Stylesgz  []byte
	Favicongz []byte

	flagAddr string
	flagCert string
	flagKey  string
)

func getFromEmbed(fs embed.FS, path string) ([]byte, error) {
	var b []byte

	f, err := fs.Open(path)
	if err != nil {
		return b, fmt.Errorf("failed to get from embed: %w", err)
	}

	defer f.Close()

	b, err = io.ReadAll(f)
	if err != nil {
		return b, fmt.Errorf("failed to read all from embed: %w", err)
	}

	return b, nil
}

func getMinifier() *minify.M {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)

	return m
}

func loadResources() {
	var err error

	// load the static files into memory
	Alpinejs, err = getFromEmbed(content, consts.PathAlpineJS)
	if err != nil {
		log.Fatalf("failed to load from %v: %v", consts.PathAlpineJS, err.Error())
	}

	Semantic, err = getFromEmbed(content, consts.PathSemantic)
	if err != nil {
		log.Fatalf("failed to load from %v: %v", consts.PathSemantic, err.Error())
	}

	stylesRaw, err := getFromEmbed(content, consts.PathStylesLocal)
	if err != nil {
		log.Fatalf("failed to load from %v: %v", consts.PathStylesLocal, err.Error())
	}

	indexRaw, err := getFromEmbed(content, consts.PathIndex)
	if err != nil {
		log.Fatalf("failed to load from %v: %v", consts.PathIndex, err.Error())
	}

	iconRaw, err := getFromEmbed(content, consts.PathFavicon)
	if err != nil {
		log.Fatalf("failed to load from %v: %v", consts.PathFavicon, err.Error())
	}

	// set up the minifier
	m := getMinifier()

	indexMini, err := m.Bytes("text/html", indexRaw)
	if err != nil {
		log.Fatalf("failed to minify index: %v", err.Error())
	}

	styles, err := m.Bytes("text/css", stylesRaw)
	if err != nil {
		log.Fatalf("failed to minify styles: %v", err.Error())
	}

	favicon, err := m.Bytes("image/svg+xml", iconRaw)
	if err != nil {
		log.Fatalf("failed to minify favicon: %v", err.Error())
	}

	Stylesgz, err = utils.GzBytes(styles)
	if err != nil {
		log.Fatalf("failed to gzip minified styles: %v", err.Error())
	}

	Favicongz, err = utils.GzBytes(favicon)
	if err != nil {
		log.Fatalf("failed to gzip minified favicon: %v", err.Error())
	}

	simpleWords, simpleWordCount := utils.GetWords(
		content,
		utils.WordsSimplePath,
	)

	complexWords, complexWordCount := utils.GetWords(
		content,
		utils.WordsComplexPath,
	)

	words := utils.Words{
		Simple:       &simpleWords,
		SimpleCount:  simpleWordCount,
		Complex:      &complexWords,
		ComplexCount: complexWordCount,
	}

	Words = &words

	Index, err = template.New(consts.PathIndex).Parse(string(indexMini))
	if err != nil {
		log.Fatalf("failed to parse index html template: %v", err.Error())
	}
}

// parseFlags parses the command line flags, using t as the translation map.
func parseFlags() {
	flag.StringVar(&flagAddr, "addr", "0.0.0.0:29102", "the address (host and port) to listen on")
	flag.StringVar(&flagCert, "cert", "", "the cert.pem file to use for TLS - leave blank for no TLS")
	flag.StringVar(&flagKey, "key", "", "the key.pem file to use for TLS - leave blank for no TLS")

	flag.Parse()
}

func main() {
	var err error

	parseFlags()

	loadResources()

	routeAlpine := fmt.Sprintf("/%v", consts.PathAlpineJS)
	routeSemantic := fmt.Sprintf("/%v", consts.PathSemantic)
	routeStyles := fmt.Sprintf("/%v", consts.PathStyles)
	routeFonts := fmt.Sprintf("/%v/:font", consts.PathFonts)
	routeFavicon := "/favicon.ico"

	FontCache = make(map[string][]byte)

	http.HandleFunc("/", getIndex)
	http.HandleFunc("/gen", getGenPassword)
	http.HandleFunc("/robots.txt", getRobotsTxt)
	http.HandleFunc("/healthcheck", getHealthCheck)
	http.HandleFunc(routeFonts, getFonts)
	http.HandleFunc(routeAlpine, getAlpine)
	http.HandleFunc(routeSemantic, getSemantic)
	http.HandleFunc(routeStyles, getStyles)
	http.HandleFunc(routeFavicon, getFavicon)

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
