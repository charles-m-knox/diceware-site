package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"

	dice "git.cmcode.dev/cmcode/go-dicewarelib"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
)

func getMinifier() *minify.M {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)

	return m
}

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

func loadResources() {
	m := getMinifier()
	{
		var err error
		var indexr []byte // The unminified index.html template
		indexr, err = getFromEmbed(content, pathIndex)
		if err != nil {
			log.Fatalf("failed to load from %v: %v", pathIndex, err.Error())
		}

		ib := bytes.NewBuffer(indexr)
		ob := bytes.NewBuffer([]byte{})
		err = m.Minify("text/html", ob, ib)
		if err != nil {
			log.Fatalf("failed to minify index: %v", err.Error())
		}

		i := ob.String()
		Index, err = template.New(pathIndex).Parse(i)
		if err != nil {
			log.Fatalf("failed to parse index html template: %v", err.Error())
		}

		log.Printf("minified index.html (%v/%v bytes)", len(i), len(indexr))
	}
	{
		var err error
		var stylesr []byte // The unminified styles.css
		stylesr, err = getFromEmbed(static, pathStyles)
		if err != nil {
			log.Fatalf("failed to load from %v: %v", pathStyles, err.Error())
		}

		ib := bytes.NewBuffer(stylesr)
		ob := bytes.NewBuffer([]byte{})
		err = m.Minify("text/css", ob, ib)
		if err != nil {
			log.Fatalf("failed to minify styles: %v", err.Error())
		}

		styles = ob.Bytes()

		log.Printf("minified styles (%v/%v bytes)", len(styles), len(stylesr))
	}

	{
		simples, scount := dice.GetWords(content, simple)
		Words = &dice.Words{}
		Words.Simple = &simples
		Words.SimpleCount = scount

		if flagExtra {
			complexs, ccount := dice.GetWords(content, complex)
			Words.Complex = &complexs
			Words.ComplexCount = ccount
		}
	}
}
