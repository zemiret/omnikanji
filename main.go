package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var templates = template.Must(template.ParseFiles("index.html"))
var kanjidmgLinks = make(map[string]string)

const (
	jishoSearchUrl = "https://jisho.org/search/"

	kanjidmgBaseUrl = "http://www.kanjidamage.com"
	kanjidmgListUrl = kanjidmgBaseUrl + "/kanji"
)

type TemplateParams struct {
	JishoWord        string
	JishoKanjis      []string
	KanjidmgSections []string
}

// TODO: Change back to using structures (don't be a dumbass allowing for html injection)
// TODO: Fix superflous response.WriteHeader call (main.go:87)
// TODO: Jisho kanjis section (in jukugo)
// TODO: Periodic refresh of kanjidmg list of kanjis (once every week is enough)

func main() {
	if err := loadKanjidmgLinks(); err != nil {
		log.Fatal("error getting kanjidamage kanji list: " + err.Error())
	}

	//log.Println("Kanjidmg links:")
	//for k, v := range kanjidmgLinks {
	//	log.Printf("%s: %s\n", k, v)
	//}

	http.HandleFunc("/", handleIndex)
	log.Println("Starting server at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadKanjidmgLinks() error {
	resp, err := http.Get(kanjidmgListUrl)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	doc.Find(".container .row table tr").Each(func(_ int, el *goquery.Selection) {
		link := el.Find("td:nth-child(3) a").First()

		if href, ok := link.Attr("href"); ok {
			kanjidmgLinks[link.Text()] = kanjidmgBaseUrl + href
		}
	})

	return nil
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/search/") {
		handleSearch(w, r)
		return
	}

	renderTemplate(w, nil)
}

func renderTemplate(w http.ResponseWriter, data *TemplateParams) {
	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	word := r.URL.Query().Get("word")
	if word == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	var tParams TemplateParams

	log.Println("Gonna search for: " + word)
	// TODO: Validation if it's in japanese
	// TODO: Split search if it's kana (do not do kanjidamage then)

	resp, err := doJishoRequest(word)
	if err != nil {
		log.Println("Error doing jisho request: ", err.Error())
		return
	}
	defer resp.Body.Close()

	jishoSection, err := parseJishoResponse(resp)
	if err != nil {
		log.Println("Error parsing jisho response: ", err.Error())
		return
	}

	tParams.JishoWord = jishoSection

	for _, c := range word {
		log.Printf("getting kanjidmg: %c\n", c)
		sect, err := getKanjidmgSection(c)
		if err != nil {
			log.Println("error getting kanjidamage kanji: " + err.Error())
			continue
		}

		tParams.KanjidmgSections = append(tParams.KanjidmgSections, sect)
	}

	renderTemplate(w, &tParams)
}

func getKanjidmgSection(kanji rune) (string, error) {
	url, ok := kanjidmgLinks[string(kanji)]
	if !ok {
		return "", errors.New("kanji does not exist at kanjidamage")
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("kanjidamage kanji search tatus code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	body := doc.Find("body")

	body = sanitizeSelection(body)
	body.Find(".navbar").Remove()
	body.Find("footer").Remove()

	return body.Html()
}

func doJishoRequest(word string) (*http.Response, error) {
	resp, err := http.Get(jishoSearchUrl + word)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("jisho status code: %d", resp.StatusCode)
	}

	return resp, nil
}

func parseJishoResponse(resp *http.Response) (string, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	body := sanitizeSelection(doc.Find("body"))
	return body.Find(".concept_light").First().Html()
}

func sanitizeSelection(sel *goquery.Selection) *goquery.Selection {
	sel.Find("script").Remove()
	return sel
}
