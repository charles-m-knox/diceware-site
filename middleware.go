package main

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
)

func cacheHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, max-age=604800")
		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// type zstdResponseWriter struct {
// 	io.Writer
// 	http.ResponseWriter
// }

// func (w zstdResponseWriter) Write(b []byte) (int, error) {
// 	return w.Writer.Write(b)
// }

func compressionHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// prefer zstd over gzip, but there seems to be a memory leak in zstd?

		c := r.Header.Get("Accept-Encoding")
		/* if strings.Contains(c, "zstd") {
			w.Header().Set("Content-Encoding", "zstd")
			zw, err := zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
			if err != nil {
				log.Printf("failed to instantiate zstd middleware: %v", err.Error())
				next.ServeHTTP(w, r)
			}
			defer zw.Close()
			next.ServeHTTP(zstdResponseWriter{Writer: zw, ResponseWriter: w}, r)
		} else */if strings.Contains(c, "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gz, err := gzip.NewWriterLevel(w, gzip.BestCompression)
			if err != nil {
				log.Printf("failed to instantiate gzip middleware: %v", err.Error())
				next.ServeHTTP(w, r)
			}
			defer gz.Close()
			next.ServeHTTP(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
		} else {
			next.ServeHTTP(w, r)
			return
		}
	})
}
