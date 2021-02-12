package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var templates = template.Must(template.ParseFiles("index.html"))

type JishoWord struct {
	Word     string
	Reading  string
	Meanings []JishoMeaning
	//Notes *string
}

type JishoMeaning struct {
	Meaning          string
	MeaningTags      *string
	//MeaningSentence  *string
	//SupplementalInfo *string
}

func main() {
	http.HandleFunc("/", handleIndex)
	log.Println("Starting server at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/search/") {
		handleSearch(w, r)
	} else if strings.HasPrefix(r.URL.Path, "/view/") {
		handleView(w, r)
	}

	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	word := r.URL.Query().Get("word")
	if word == "" {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	log.Println("Gonna search for: " + word)
	// TODO: Validation if it's in japanese
	// TODO: Split search if it's kana (do not do kanjidamage then)
	// TODO: Kanjidamage

	resp, err := doJishoRequest(word)
	if err != nil {
		log.Println("Error doing jisho request: ", err.Error())
		return
	}
	defer resp.Body.Close()

	jishoWord, err := parseJishoResponse(resp)
	if err != nil {
		log.Println("Error parsing jisho response: ", err.Error())
		return
	}


	log.Println("JISHO CONTENT: ")
	log.Println("word:" , jishoWord.Word)
	log.Println("reading:", jishoWord.Reading)
	for _, m := range jishoWord.Meanings {
		log.Println("mezning:", m.Meaning)
		if m.MeaningTags != nil {
			log.Println("tag:", *m.MeaningTags)
		}
	}
}

func handleView(w http.ResponseWriter, r *http.Request) {

}

func doJishoRequest(word string) (*http.Response, error) {
	resp, err := http.Get("https://jisho.org/search/" + word)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("jisho status code: %d", resp.StatusCode)
	}

	return resp, nil
}

func parseJishoResponse(resp *http.Response) (*JishoWord, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var jishoWord JishoWord
	wordSection := doc.Find(".concept_light").First()

	readingSection := wordSection.Find(".concept_light-wrapper .concept_light-readings").First()
	jishoWord.Word = strings.TrimSpace(readingSection.Find(".text").Text())
	jishoWord.Reading = strings.TrimSpace(readingSection.Find(".furigana").Text())

	meaningSection := wordSection.Find(".meanings-wrapper" ).First()

	lastTag := ""
	meaningSection.Children().Each(func(_ int, el *goquery.Selection) {
		if el.HasClass("meaning-tags") {
			lastTag = strings.TrimSpace(el.Text())
		} else if el.HasClass("meaning-wrapper") {
			var jishoMeaning JishoMeaning
			jishoMeaning.Meaning = strings.TrimSpace(el.Find(".meaning-meaning").Text())
			if lastTag != "" {
				t := lastTag
				jishoMeaning.MeaningTags = &t
			}
			jishoWord.Meanings = append(jishoWord.Meanings, jishoMeaning)
			lastTag = ""
		}
	})

	return &jishoWord, nil
}
