package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zemiret/omnikanji"
	"github.com/zemiret/omnikanji/jptext"
)

type HttpClientMock struct {
	staticDir string
}

func NewHttpClientMock(staticDir string) *HttpClientMock {
	return &HttpClientMock{
		staticDir: staticDir,
	}
}

func (c *HttpClientMock) Get(searchUrl string) (*http.Response, error) {
	var filePath string
	if strings.HasPrefix(searchUrl, omnikanji.JishoSearchUrl) {
		word := strings.TrimPrefix(searchUrl, omnikanji.JishoSearchUrl) // this could go onto "jisho" aggregate not to spill out its logic
		filePath = filepath.Join(c.staticDir, "jisho", word)
	} else if strings.HasPrefix(searchUrl, omnikanji.KanjidmgBaseUrl) {
		word := strings.TrimPrefix(searchUrl, omnikanji.KanjidmgBaseUrl) // this could go onto "kanjidmg" aggregate not to spill out its logic
		filePath = filepath.Join(c.staticDir, "kanjidmg", word)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w, filepath: %s", err, filePath)
	}

	return &http.Response{
		Body: f,
	}, nil
}

func TestServer(t *testing.T) {
	lookups := []string{
		"何",
		"兄弟",
		"路面電車停留場 ",
		"あったり前",
		"相変わらず",
		"ペラペラ",
		"driver's licence",
	}

	kanjidmgLinks := make(map[string]string)
	for _, word := range lookups {
		for _, r := range word {
			if !jptext.IsKanji(r) {
				continue
			}

			rStr := string(r)
			kanjidmgLinks[rStr] = omnikanji.KanjidmgBaseUrl + rStr
		}
	}

	type TestCase struct {
		word string
	}

	// run := func(tc *TestCase) func(t *testing.T) {
	// 	return func(t *testing.T) {
	// httpClient := NewHttpClientMock("testdata")
	// jisho := dictproxy.NewJisho(omnikanji.JishoSearchUrl, httpClient)
	// kanjidmg := dictproxy.NewKanjidmg(kanjidmgLinks, httpClient)
	// srv := NewServer(jisho, kanjidmg)

	// TODO: Alright, not 100% sure this is the best way to call this. Take a look at httptest packge!

	// req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("localhost:8080/search/?word=%s", tc.word), nil)
	// if err != nil {
	// 	t.Fatalf("http.NewRequest: %s", err)
	// }

	// TODO: We will need one step more ins server to get pure data here, and not the rendered template
	// Idea: We can add a wrapper in server over current handlers (and add it in handlers) that will call render template on the provided data

	// srv.handleIndex()
	// }
	// }
}
