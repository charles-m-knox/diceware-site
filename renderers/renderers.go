package renderers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	consts "gitea.cmcode.dev/cmcode/diceware-site/constants"
	"gitea.cmcode.dev/cmcode/diceware-site/utils"
)

func handleParams(r *http.Request) (int, string, int, int, bool) {
	var n string

	var s string

	var maxLen string

	var minLen string

	var extendedWordList string

	if r.Method == http.MethodPost {
		r.ParseForm()

		n = r.Form.Get("n")
		s = r.Form.Get("s")
		maxLen = r.Form.Get("u")
		minLen = r.Form.Get("l")
		extendedWordList = r.Form.Get("e")
	} else {
		n = r.URL.Query().Get("n")
		s = r.URL.Query().Get("s")
		maxLen = r.URL.Query().Get("u")
		minLen = r.URL.Query().Get("l")
		extendedWordList = r.URL.Query().Get("e")
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

	extendedWordListBool := false
	// HTML form submits a checked box as "on"
	if extendedWordList == "on" {
		extendedWordListBool = true
	} else if extendedWordList != "" {
		extendedWordListBool, _ = strconv.ParseBool(extendedWordList)
	}

	return int(nn), s, int(maxLenInt), int(minLenInt), extendedWordListBool
}

func GenPassword(w http.ResponseWriter, r *http.Request, words *utils.Words) {
	nn, s, maxLenInt, minLenInt, extendedWords := handleParams(r)

	result := make(map[string]string)
	result["p"] = utils.GeneratePassword(
		words,
		nn,
		s,
		maxLenInt,
		minLenInt,
		extendedWords,
	)

	w.Header().Add(consts.ContentTypeHeader, "application/json")

	b, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GenPassword failed to marshal result: %v", err.Error())

		return
	}

	// TODO: check if length > 512 and if so, gzip the result

	_, err = w.Write(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GenPassword failed to write result: %v", err.Error())

		return
	}
}

func Index(w http.ResponseWriter, r *http.Request, words *utils.Words, index *template.Template) {
	nn, s, maxLenInt, minLenInt, extendedWords := handleParams(r)

	result := utils.GeneratePassword(
		words,
		nn,
		s,
		maxLenInt,
		minLenInt,
		extendedWords,
	)

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
	}

	err := index.Execute(buf, data)
	if err != nil {
		log.Printf("failed to render index template: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	resultgz, err := utils.GzBytes(buf.Bytes())
	if err != nil {
		log.Printf("failed to gzip result: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Add(consts.ContentEncodingHeader, consts.ContentEncodingGzipHeaderValue)
	w.Header().Add(consts.ContentTypeHeader, "text/html")

	_, err = w.Write(resultgz)
	if err != nil {
		log.Printf("failed to write gzip result to w: %v", err.Error())
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
