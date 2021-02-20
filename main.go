package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/zemiret/omnikanji/ptr"
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
	Jisho    *JishoSection
	Kanjidmg []*KanjidmgSection
}

type KanjidmgSection struct {
	Kanji      KanjidmgKanji
	TopComment *string
	Radicals   []KanjidmgKanji
	Onyomi     *string
	Mnemonic   *string
	Mutants    []KanjidmgKanji

	// TODO: Sections find by text (e.g. Onyomi), some pages have different sections

	// Kunyomi TODO
	//Jukugo  TODO
	// UsedIn TODO
	// Synonyms TODO
	// Lookalikes TODO
}

type KanjidmgKanji struct {
	Kanji   string
	Meaning string
	Link    string
}

type JishoSection struct {
	Link   string
	Word   JishoWord
	Kanjis []JishoKanji
}

type JishoWord struct {
	Word     string
	Reading  string
	Meanings []JishoMeaning
	//Notes *string
}

type JishoMeaning struct {
	Meaning string
	Tags    *string
	//MeaningSentence  *string
	//SupplementalInfo *string
}

type JishoKanji struct {
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
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		log.Println("Error doing jisho request: ", err.Error())
		return
	}

	jishoSection, err := parseJishoResponse(resp)
	if err != nil {
		log.Println("Error parsing jisho response: ", err.Error())
		return
	}
	jishoSection.Link = jishoUrl(word)
	tParams.Jisho = jishoSection

	for _, c := range word {
		log.Printf("getting kanjidmg: %c\n", c)
		sect, err := getKanjidmgSection(c)
		if err != nil {
			log.Println("error getting kanjidamage kanji: " + err.Error())
			continue
		}

		tParams.Kanjidmg = append(tParams.Kanjidmg, sect)
	}

	log.Println("TEMPLATE PARAMS JISHO: ")
	log.Printf("%+v\n", tParams.Jisho)

	log.Println("TEMPLATE PARAMS KANJIDMG: ")
	log.Printf("%+v\n", tParams.Kanjidmg[0])

	renderTemplate(w, &tParams)
}

func doKanjidmgRequest(kanji rune) (*http.Response, error) {
	url := kanjidmgUrl(string(kanji))
	if url == "" {
		return nil, errors.New("kanji does not exist at kanjidamage")
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("kanjidamage kanji search tatus code: %d", resp.StatusCode)
	}

	return resp, nil
}

func getKanjidmgSection(kanji rune) (*KanjidmgSection, error) {
	resp, err := doKanjidmgRequest(kanji)
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

	sect := &KanjidmgSection{}
	rows := doc.Find(".container").Last().Find(".row")

	kanjiRow := rows.Eq(1)

	// TODO: HANDLING KANJIS THAT ARE IMAGES NOT TEXT0
	if kanjiRow != nil {
		sect.Kanji.Kanji = kanjiRow.Find("h1 .kanji_character").Text()
		sect.Kanji.Meaning = kanjiRow.Find("h1 .translation").Text()
		sect.Kanji.Link = kanjidmgUrl(string(kanji))

		kanjiRow.Find("h1").Remove()

		radicalSection := kanjiRow.Find("div.span8")
		radicalLinks := radicalSection.Find("a")

		usedLinks := 0
		radicalSection.Contents().Not("a").Each(func(_ int, m *goquery.Selection) {
			meaningText := m.Text() // TODO: text cleanup
			if strings.TrimSpace(meaningText) != "" {
				log.Println("Meaning:  ", meaningText)

				sect.Radicals = append(sect.Radicals, KanjidmgKanji{
					Kanji:   radicalLinks.Eq(usedLinks).Text(),
					Meaning: m.Text(),
					Link:    radicalLinks.Eq(usedLinks).AttrOr("href", ""),
				})
				usedLinks += 1
			}
		})
	} else {
		log.Println("Kanjidmg: Kanji row is nil :/")
	}

	sectionsRow := rows.Eq(2)
	if sectionsRow != nil {
		sect.Onyomi = ptr.String(sectionsRow.Find("h2:contains(\"Onyomi\")").Next().Text())
		sect.Mnemonic = ptr.String(sectionsRow.Find("h2:contains(\"Mnemonic\")").Next().Text())
	} else {
		log.Println("Kanjidmg: sections row is nil :/")
	}

	// TODO: Top comment

	return sect, nil
}

func doJishoRequest(word string) (*http.Response, error) {
	resp, err := http.Get(jishoUrl(word))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("jisho status code: %d", resp.StatusCode)
	}

	return resp, nil
}

func parseJishoResponse(resp *http.Response) (*JishoSection, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	var jishoWord JishoWord
	wordSection := doc.Find(".concept_light").First()

	readingSection := wordSection.Find(".concept_light-wrapper .concept_light-readings").First()
	jishoWord.Word = strings.TrimSpace(readingSection.Find(".text").Text())
	jishoWord.Reading = strings.TrimSpace(readingSection.Find(".furigana").Text())

	meaningSection := wordSection.Find(".meanings-wrapper").First()

	lastTag := ""
	meaningSection.Children().Each(func(_ int, el *goquery.Selection) {
		if el.HasClass("meaning-tags") {
			lastTag = strings.TrimSpace(el.Text())
		} else if el.HasClass("meaning-wrapper") {
			var jishoMeaning JishoMeaning
			jishoMeaning.Meaning = strings.TrimSpace(el.Find(".meaning-meaning").Text())
			if lastTag != "" {
				t := lastTag
				jishoMeaning.Tags = &t
			}
			jishoWord.Meanings = append(jishoWord.Meanings, jishoMeaning)
			lastTag = ""
		}
	})

	return &JishoSection{
		Word:   jishoWord,
		Kanjis: nil, // TODO: Kanjis
	}, nil
}

func jishoUrl(word string) string {
	return jishoSearchUrl + word
}

func kanjidmgUrl(kanji string) string {
	return kanjidmgLinks[kanji]
}
