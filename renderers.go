package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	dice "git.cmcode.dev/cmcode/go-dicewarelib"
)

func handleParams(r *http.Request) (int, string, int, int, bool) {
	var n string
	var s string
	var maxLen string
	var minLen string
	var extra string

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Printf("handleParams failed to parse: %v", err.Error())
		}

		n = r.Form.Get("n")
		s = r.Form.Get("s")
		maxLen = r.Form.Get("u")
		minLen = r.Form.Get("l")
		extra = r.Form.Get("e")
	} else {
		n = r.URL.Query().Get("n")
		s = r.URL.Query().Get("s")
		maxLen = r.URL.Query().Get("u")
		minLen = r.URL.Query().Get("l")
		extra = r.URL.Query().Get("e")
	}

	if strings.ToLower(s) == "space" {
		s = " "
	}

	// if no values are specified, set the default separator to a space
	// character - this is because the default load page doesn't specify any
	// params
	if n == "" && s == "" && maxLen == "" && minLen == "" {
		s = " "
	}

	// assign a default value of 3 words if the user didn't
	// specify a desired number of words
	nn := int64(3)

	if n != "" {
		var err error

		nn, err = strconv.ParseInt(n, 10, 64)
		if err != nil {
			nn = 3
		}
	}

	// assign a default value of 32 characters if the user didn't
	// specify a desired max length
	maxLenInt := int64(32)
	if maxLen != "" {
		maxLenInt, _ = strconv.ParseInt(maxLen, 10, 64)
	}

	// assign a default value of 20 characters if the user didn't
	// specify a desired min length
	minLenInt := int64(20)
	if minLen != "" {
		minLenInt, _ = strconv.ParseInt(minLen, 10, 64)
	}

	useExtra := false
	// HTML form submits a checked box as "on"
	if extra == "on" {
		useExtra = true
	} else if extra != "" {
		useExtra, _ = strconv.ParseBool(extra)
	}

	return int(nn), s, int(maxLenInt), int(minLenInt), useExtra
}

// Generates a password via an API call.
func renderGenPassword(w http.ResponseWriter, r *http.Request, words *dice.Words) {
	nn, s, maxLenInt, minLenInt, extendedWords := handleParams(r)

	if !flagExtra {
		extendedWords = false
	}

	result := make(map[string]string)
	result["p"] = dice.GeneratePassword(words, nn, s, maxLenInt, minLenInt, extendedWords)

	w.Header().Add(contentTypeHeader, "application/json")

	b, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GenPassword failed to marshal result: %v", err.Error())

		return
	}

	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GenPassword failed to write result: %v", err.Error())

		return
	}
}

// Generates a password and renders the index template.
func renderIndex(w http.ResponseWriter, r *http.Request, words *dice.Words, index *template.Template) {
	nn, s, maxLenInt, minLenInt, extendedWords := handleParams(r)
	result := dice.GeneratePassword(words, nn, s, maxLenInt, minLenInt, extendedWords)
	buf := new(bytes.Buffer)
	failed := false
	if result == "" {
		failed = true
	}

	data := map[string]any{
		"pw":                result,
		"simpleWordCount":   words.SimpleCount,
		"extendedWordCount": words.ComplexCount,
		"nn":                nn,
		"s":                 s,
		"maxLenInt":         maxLenInt,
		"minLenInt":         minLenInt,
		"e":                 extendedWords,
		"pwLength":          len(result),
		"failed":            failed,
		"flagExtra":         flagExtra,
	}

	err := index.Execute(buf, data)
	if err != nil {
		log.Printf("failed to render index template: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Add(contentTypeHeader, "text/html")

	_, err = w.Write(buf.Bytes())
	if err != nil {
		log.Printf("failed to write gzip result to w: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
