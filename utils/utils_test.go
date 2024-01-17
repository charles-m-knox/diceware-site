package utils

import (
	"log"
	"slices"
	"strings"
	"testing"
)

func TestGetRandomInt(t *testing.T) {
	// a very naive testing method that just generates a bunch
	// of GetRandomInt values and asserts that none of them are outside
	// the specified range
	maxInt := 10

	for i := 0; i < 2500; i++ {
		actual := GetRandomInt(maxInt)
		if actual > maxInt {
			t.Fail()
		}

		if actual < 0 {
			t.Fail()
		}
	}
}

func TestGetRandomSymbol(t *testing.T) {
	for i := 0; i < 2500; i++ {
		actual := getRandomSymbol()
		if !slices.Contains(symbols, actual) {
			t.Fail()
		}
	}
}

func setupRandomWordTest() ([]string, *Words) {
	// generate a min-length string
	sb := new(strings.Builder)
	for i := 0; i < MinWordLength; i++ {
		sb.WriteString("a")
	}

	shortWord := sb.String()

	sb.Reset()

	// generate a max-length string
	for i := 0; i < MaxWordLength+1; i++ {
		sb.WriteString("a")
	}

	longWord := sb.String()

	// there's probably a more efficient way to do this but it's not critical
	// right now
	validWords := []string{
		shortWord,
		"onlinux",
		longWord,
		"fedora",
		"debian",
		"archlinux",
		"qubes",
	}

	wordList := map[int]string{
		0: shortWord,
		1: "onlinux",
		2: longWord,
		3: "fedora",
		4: "debian",
		5: "archlinux",
		6: "qubes",
	}

	words := Words{
		Simple:       &wordList,
		SimpleCount:  7,
		Complex:      &wordList,
		ComplexCount: 7,
	}

	return validWords, &words
}

func TestGetRandomWord(t *testing.T) {
	validWords, words := setupRandomWordTest()

	for i := 0; i < 5000; i++ {
		actual := getRandomWord(*words.Complex)

		if len(actual) > MaxWordLength {
			t.Fail()
		}

		if len(actual) < MinWordLength {
			t.Fail()
		}

		if !slices.Contains(validWords, actual) {
			t.Fail()
		}

		if actual == "" {
			t.Fail()
		}
	}
}

func TestGeneratePassword(t *testing.T) {
	_, words := setupRandomWordTest()

	tests := []struct {
		n             int
		s             string
		maxLen        int
		minLen        int
		shouldSucceed bool
		name          string
	}{
		{4, " ", 64, 4, true, "1"},
		{4, " ", 32, 4, true, "2"},
		{5, " ", 32, 4, true, "3"},
		{3, " ", 64, 4, true, "4"},
		{4, "2", 64, 4, true, "5"},
		{90, "3", 64, 4, false, "6"},
		{90, "-", 64, 4, false, "7"},
		{90, " ", 10, 4, false, "8"},
	}

	for _, test := range tests {
		log.Printf("test: %v", test.name)
		actual := GeneratePassword(words, test.n, test.s, test.maxLen, test.minLen, false)

		if test.shouldSucceed {
			if len(actual) > test.maxLen {
				t.Fail()
			}

			if len(actual) < test.minLen {
				t.Fail()
			}

			if actual == "" {
				t.Fail()
			}

			continue
		}

		if actual != "" {
			t.Fail()
		}
	}
}
