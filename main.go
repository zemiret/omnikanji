package main

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
)

// TODO: Get back to using this
//var templates = template.Must(template.ParseFiles("index.html"))

const (
	jishoSearchUrl = "https://jisho.org/search/"

	kanjidmgBaseUrl = "http://www.kanjidamage.com"
	kanjidmgListUrl = kanjidmgBaseUrl + "/kanji"
)


// TODO: DEBUG links to radicals do not work in kanjidamage
// TODO: Periodic refresh of kanjidmg list of kanjis (once every week or even month is enough)
// TODO: Better reading handling in jisho (each kanji reading)

func main() {
	kanjidmgLinks, err := loadKanjidmgLinks()
	if err != nil {
		log.Fatal("error getting kanjidamage kanji list: " + err.Error())
	}

	jisho := NewJishoHandler(jishoSearchUrl)
	kanjidmg := NewKanjidmgHandler(kanjidmgLinks)
	srv := newServer(jisho, kanjidmg)

	http.HandleFunc("/", srv.handleIndex)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	log.Println("Starting server at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadKanjidmgLinks() (map[string]string, error) {
	resp, err := http.Get(kanjidmgListUrl)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	links := make(map[string]string)

	doc.Find(".container .row table tr").Each(func(_ int, el *goquery.Selection) {
		link := el.Find("td:nth-child(3) a").First()

		if href, ok := link.Attr("href"); ok {
			links[link.Text()] = kanjidmgBaseUrl + href
		}
	})

	return links, nil
}
