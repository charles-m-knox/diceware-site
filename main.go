package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	consts "gitea.cmcode.dev/cmcode/diceware-site/constants"
	"gitea.cmcode.dev/cmcode/diceware-site/renderers"
	"gitea.cmcode.dev/cmcode/diceware-site/utils"

	"github.com/gin-gonic/gin"
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

// https://golangdocs.com/reading-files-in-golang
func getFromEmbed(fs embed.FS, path string) (string, error) {
	f, err := fs.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to get from embed: %v", err.Error())
	}

	var s strings.Builder

	b := make([]byte, 4) // max size of chunks
	for {
		readTotal, err := f.Read(b)
		if err != nil {
			if err != io.EOF {
				return "", fmt.Errorf(
					"failed to read from embedded file %v: %v",
					path,
					err.Error(),
				)
			}
			break
		}
		s.WriteString(string(b[:readTotal]))
	}

	return s.String(), nil
}

func main() {
	// load the static files into memory
	alpinejs, err := getFromEmbed(content, consts.PATH_ALPINEJS)
	if err != nil {
		log.Fatalf("failed to load from %v: %v", consts.PATH_ALPINEJS, err.Error())
	}
	semantic, err := getFromEmbed(content, consts.PATH_SEMANTIC)
	if err != nil {
		log.Fatalf("failed to load from %v: %v", consts.PATH_SEMANTIC, err.Error())
	}
	stylesRaw, err := getFromEmbed(content, consts.PATH_STYLES_LOCAL)
	if err != nil {
		log.Fatalf("failed to load from %v: %v", consts.PATH_STYLES_LOCAL, err.Error())
	}
	indexRaw, err := getFromEmbed(content, consts.PATH_INDEX)
	if err != nil {
		log.Fatalf("failed to load from %v: %v", consts.PATH_INDEX, err.Error())
	}
	iconRaw, err := getFromEmbed(content, consts.PATH_FAVICON)
	if err != nil {
		log.Fatalf("failed to load from %v: %v", consts.PATH_FAVICON, err.Error())
	}

	// set up the minifier
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)

	indexMini, err := m.String("text/html", indexRaw)
	if err != nil {
		log.Fatalf("failed to minify index: %v", err.Error())
	}

	styles, err := m.String("text/css", stylesRaw)
	if err != nil {
		log.Fatalf("failed to minify styles: %v", err.Error())
	}

	favicon, err := m.String("image/svg+xml", iconRaw)
	if err != nil {
		log.Fatalf("failed to minify favicon: %v", err.Error())
	}

	stylesgz, err := utils.GzStr(styles)
	if err != nil {
		log.Fatalf("failed to gzip minified styles: %v", err.Error())
	}

	favicongz, err := utils.GzStr(favicon)
	if err != nil {
		log.Fatalf("failed to gzip minified favicon: %v", err.Error())
	}

	simpleWords, simpleWordCount := utils.GetWords(
		content,
		utils.WORDS_SIMPLE_PATH,
	)

	complexWords, complexWordCount := utils.GetWords(
		content,
		utils.WORDS_COMPLEX_PATH,
	)

	words := utils.Words{
		Simple:       &simpleWords,
		SimpleCount:  simpleWordCount,
		Complex:      &complexWords,
		ComplexCount: complexWordCount,
	}

	// leaving here for reference: the gin.Default() actually logs IP addresses
	// upon receiving requests; we don't want this
	// r := gin.Default()

	r := gin.New()
	r.Use(
		// gin.LoggerWithWriter(gin.DefaultWriter, "/"),
		gin.Recovery(),
	)

	trusted := []string{
		// "192.168.0.1",
	}

	// https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies
	r.SetTrustedProxies(trusted)

	r.SetFuncMap(template.FuncMap{})

	// the below LoadHTMLGlob doesn't seem compatible with embed.FS
	// r.LoadHTMLGlob("templates/*")
	// this is the fix
	// https://github.com/gin-gonic/gin/issues/2795
	// LoadHTMLFromEmbedFS(r, content, "templates/*")

	index, err := template.New(consts.PATH_INDEX).Parse(indexMini)
	if err != nil {
		log.Fatalf("failed to parse index html template: %v", err.Error())
	}

	// in order to support gzipping, we elect not to use gin's html
	// templating
	// r.SetHTMLTemplate(index)

	// normally this would be called, but because of the above LoadHTMLGlob
	// issues, this should not be called
	// r.Static("/static", "./static")

	routeAlpine := fmt.Sprintf("/%v", consts.PATH_ALPINEJS)
	routeSemantic := fmt.Sprintf("/%v", consts.PATH_SEMANTIC)
	routeStyles := fmt.Sprintf("/%v", consts.PATH_STYLES)
	routeFonts := fmt.Sprintf("/%v/:font", consts.PATH_FONTS)
	routeFavicon := "/favicon.ico"
	fontCache := make(map[string]string)
	r.GET(routeAlpine, func(c *gin.Context) {
		c.Header(consts.CONTENTENCODING_HEADER, consts.CONTENTENCODING_HEADER_VALUE)
		c.Header(consts.CONTENTTYPE_HEADER, "application/javascript")
		c.Header(consts.CACHECONTROL_HEADER, consts.CACHECONTROL_HEADER_VALUE)
		c.String(http.StatusOK, alpinejs)
	})
	r.GET(routeSemantic, func(c *gin.Context) {
		c.Header(consts.CONTENTENCODING_HEADER, consts.CONTENTENCODING_HEADER_VALUE)
		c.Header(consts.CONTENTTYPE_HEADER, "text/css")
		c.Header(consts.CACHECONTROL_HEADER, consts.CACHECONTROL_HEADER_VALUE)
		c.String(http.StatusOK, semantic)
	})
	r.GET(routeStyles, func(c *gin.Context) {
		c.Header(consts.CONTENTENCODING_HEADER, consts.CONTENTENCODING_HEADER_VALUE)
		c.Header(consts.CONTENTTYPE_HEADER, "text/css")
		c.Header(consts.CACHECONTROL_HEADER, consts.CACHECONTROL_HEADER_VALUE)
		c.String(http.StatusOK, stylesgz)
	})
	r.GET(routeFavicon, func(c *gin.Context) {
		c.Header(consts.CONTENTENCODING_HEADER, consts.CONTENTENCODING_HEADER_VALUE)
		c.Header(consts.CONTENTTYPE_HEADER, "image/svg+xml")
		c.Header(consts.CACHECONTROL_HEADER, consts.CACHECONTROL_HEADER_VALUE)
		c.String(http.StatusOK, favicongz)
	})
	r.GET(routeFonts, func(c *gin.Context) {
		// note: gzipping the fonts makes no difference on the resulting size
		fontTarget := c.Param("font")
		c.Header(consts.CONTENTTYPE_HEADER, "application/font-woff")
		if strings.LastIndex(fontTarget, "woff2") == len(fontTarget)-1 {
			c.Header(consts.CONTENTTYPE_HEADER, "application/font-woff")
		}
		c.Header(consts.CACHECONTROL_HEADER, consts.CACHECONTROL_HEADER_VALUE)
		cachedFont, ok := fontCache[fontTarget]
		if !ok {
			log.Printf("no cached font for %v %v", cachedFont, fontTarget)
			fontCache[fontTarget], err = getFromEmbed(
				content,
				fmt.Sprintf("%v/%v", consts.PATH_FONTS_LOCAL, fontTarget),
			)
			if err != nil {
				log.Fatalf("failed to load from %v: %v", consts.PATH_STYLES, err.Error())
			}

			cachedFont = fontCache[fontTarget]
		}
		c.String(http.StatusOK, cachedFont)
	})

	r.GET("/", func(c *gin.Context) {
		startTime := time.Now()
		renderers.Index(c, &words, index)

		finishedTime := time.Now()

		elapsed := finishedTime.Sub(startTime)

		// enforce a minimum response time of ~30ms
		if elapsed.Milliseconds() < 30 {
			randSleep := time.Duration(30+utils.GetRandomInt(50)) * time.Millisecond
			time.Sleep(randSleep - elapsed)
		}
	})
	r.GET("/gen", func(c *gin.Context) {
		startTime := time.Now()
		renderers.GenPassword(c, &words)

		finishedTime := time.Now()

		elapsed := finishedTime.Sub(startTime)

		// enforce a minimum response time of ~30ms
		if elapsed.Milliseconds() < 30 {
			randSleep := time.Duration(30+utils.GetRandomInt(50)) * time.Millisecond
			time.Sleep(randSleep - elapsed)
		}
	})
	r.GET("/healthcheck", func(c *gin.Context) {
		c.String(http.StatusOK, consts.HEALTHCHECK_RESPONSE)
	})
	r.GET("/robots.txt", func(c *gin.Context) {
		c.String(http.StatusOK, consts.ROBOTSTXT)
	})

	err = r.RunTLS("0.0.0.0:29102", "cert.pem", "key.pem")
	if err != nil {
		log.Println("failed to run gin", err.Error())
	}
}
