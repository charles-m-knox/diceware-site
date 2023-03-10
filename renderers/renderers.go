package renderers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strconv"

	consts "diceware-site/constants"
	"diceware-site/utils"

	"github.com/gin-gonic/gin"
)

func handleParams(c *gin.Context) (int, string, int, int, bool) {
	n := c.Query("n")
	s := c.Query("s")
	maxLen := c.Query("u")
	minLen := c.Query("l")
	extendedWordList := c.Query("e")

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
	if extendedWordList != "" {
		extendedWordListBool, _ = strconv.ParseBool(extendedWordList)
	}

	return int(nn), s, int(maxLenInt), int(minLenInt), extendedWordListBool
}

func GenPassword(c *gin.Context, w *utils.Words) {
	nn, s, maxLenInt, minLenInt, extendedWords := handleParams(c)

	result := make(map[string]string)
	result["p"] = utils.GeneratePassword(
		w,
		nn,
		s,
		maxLenInt,
		minLenInt,
		extendedWords,
	)

	// WARNING: The following code actually increases the size of the transferred
	// data from around 33 bytes to 200+ bytes, so the gzip compression isn't worth
	// it. But leaving it here for later reference.

	// resultstr, err := json.Marshal(result)
	// if err != nil {
	// 	log.Printf("failed to marshal result: %v", err.Error())
	// 	c.String(http.StatusInternalServerError, "")
	// }

	// resultgz, err := utils.GzStr(string(resultstr))
	// if err != nil {
	// 	log.Printf("failed to gzip result: %v", err.Error())
	// 	c.String(http.StatusInternalServerError, "")
	// }

	// c.Header(consts.CONTENTENCODING_HEADER, consts.CONTENTENCODING_HEADER_VALUE)
	// c.Header(consts.CONTENTTYPE_HEADER, "application/json")
	// c.String(http.StatusOK, resultgz)

	c.JSON(http.StatusOK, result) // no gzip
}

func Index(c *gin.Context, w *utils.Words, index *template.Template) {
	nn, s, maxLenInt, minLenInt, extendedWords := handleParams(c)

	result := utils.GeneratePassword(
		w,
		nn,
		s,
		maxLenInt,
		minLenInt,
		extendedWords,
	)

	buf := new(bytes.Buffer)

	data := map[string]any{
		"pw":                result,
		"simpleWordCount":   w.SimpleCount,
		"extendedWordCount": w.ComplexCount,
	}

	err := index.Execute(buf, data)
	if err != nil {
		log.Printf("failed to render index template: %v", err.Error())
		c.String(http.StatusInternalServerError, "")
	}

	resultgz, err := utils.GzStr(buf.String())
	if err != nil {
		log.Printf("failed to gzip result: %v", err.Error())
		c.String(http.StatusInternalServerError, "")
	}
	c.Header(consts.CONTENTENCODING_HEADER, consts.CONTENTENCODING_HEADER_VALUE)
	c.Header(consts.CONTENTTYPE_HEADER, "text/html")
	c.String(http.StatusOK, resultgz)

	// c.HTML(
	// 	http.StatusOK,
	// 	"templates/index.html",
	// 	gin.H{
	// 		"pw":                result,
	// 		"simpleWordCount":   w.SimpleCount,
	// 		"extendedWordCount": w.ComplexCount,
	// 	},
	// )
}
