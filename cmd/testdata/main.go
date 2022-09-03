package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/zemiret/omnikanji"
	"github.com/zemiret/omnikanji/dictproxy"
	"github.com/zemiret/omnikanji/jptext"
	"github.com/zemiret/omnikanji/pkg/http"
)

func main() {
	log.Println("Loading fixture")

	httpClient := http.NewClient()
	kanjidmgLinks, err := dictproxy.LoadKanjidmgLinks(httpClient)
	if err != nil {
		log.Fatalf("LoadKanjidmgLinks: %s", err)
	}
	kanjidmg := dictproxy.NewKanjidmg(kanjidmgLinks, httpClient)
	jisho := dictproxy.NewJisho(omnikanji.JishoSearchUrl, httpClient)

	serverFixtureDir := "./server/fixture"
	if err := os.MkdirAll(fmt.Sprintf("%s/jisho", serverFixtureDir), os.ModePerm); err != nil {
		log.Fatalf("os.Mkdir: %s", err)
	}
	if err := os.MkdirAll(fmt.Sprintf("%s/kanjidmg", serverFixtureDir), os.ModePerm); err != nil {
		log.Fatalf("os.Mkdir: %s", err)
	}

	lookups := []string{
		"何",
		"兄弟",
		"路面電車停留場",
		"あったり前",
		"相変わらず",
		"ペラペラ",
		"driver's licence",
	}

	for _, word := range lookups {
		log.Printf("Looking up: %s", word)

		resp, err := httpClient.Get(jisho.Url(word))
		if err != nil {
			log.Fatalf("Error getting %s: %s", word, err)
		}
		defer resp.Body.Close()
		fn := fmt.Sprintf("%s/jisho/%s.html", serverFixtureDir, word)
		f, err := os.Create(fn)
		if err != nil {
			log.Fatalf("os.Create: %s: %s", fn, err)
		}
		defer f.Close()
		f.ReadFrom(resp.Body)

		for _, r := range word {
			if !jptext.IsKanji(r) {
				continue
			}
			resp, err := httpClient.Get(kanjidmg.Url(string(r)))
			if err != nil {
				log.Fatalf("Error getting %s: %s", word, err)
			}
			defer resp.Body.Close()
			fn := fmt.Sprintf("%s/kanjidmg/%s.html", serverFixtureDir, string(r))
			f, err := os.Create(fn)
			if err != nil {
				log.Fatalf("os.Create: %s: %s", fn, err)
			}
			defer f.Close()
			f.ReadFrom(resp.Body)
		}

		sleeptime := time.Duration(float32(time.Second) * (rand.Float32() + 0.5))
		log.Printf("Sleeping for: %dms", sleeptime.Milliseconds())
		time.Sleep(sleeptime) // be nice to DOS attack detectors
	}
}
