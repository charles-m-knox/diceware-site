package constants

const (
	PathAlpineJS                   = "static/alpinejs-3.12.0.min.js.gz"
	PathIndex                      = "templates/index.html"
	PathNoJs                       = "templates/nojs.html"
	PathMin                        = "templates/min.html"
	PathFonts                      = "static/themes/default/assets/fonts"
	PathFontsLocal                 = "static/fonts"
	PathFavicon                    = "static/favicon.ico"
	PathSemantic                   = "static/semantic-2.9.2.min.css.gz"
	PathStyles                     = "static/styles.min.css.gz"
	PathStylesLocal                = "static/styles.css"
	ContentTypeHeader              = "Content-Type"
	CacheControlHeader             = "Cache-Control"
	CacheControlHeaderValue        = "private, max-age=604800"
	ContentEncodingHeader          = "Content-Encoding"
	ContentEncodingGzipHeaderValue = "gzip"
)

var (
	RobotsTxt           = []byte("User-agent: *\nDisallow: /\n")
	HealthcheckResponse = []byte("1")
)
