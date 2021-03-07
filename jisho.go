package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

type JishoHandler struct {
	searchUrl string
}

func NewJishoHandler(jishoSearchUrl string) *JishoHandler {
	return &JishoHandler{searchUrl: jishoSearchUrl}
}

type JishoSection struct {
	Link        string
	WordSection JishoWordSection
	Kanjis      []JishoKanji
}

type JishoWordSection struct {
	Word     string
	Reading  string
	Meanings []JishoMeaning
	//Notes *string
}

type JishoMeaning struct {
	ListItem
	Meaning string
	Tags    *string
	//MeaningSentence  *string
	//SupplementalInfo *string
}

type JishoKanji struct {
	Kanji    JishoWordWithLink
	Meaning  string
	Kunyomis []JishoWordWithLink
	Onyomis  []JishoWordWithLink
}

type JishoWordWithLink struct {
	Link string
	Word string
}

func (h *JishoHandler) GetSection(word string) (*JishoSection, error) {
	url := h.url(word)

	resp, err := http.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	sect, err := h.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("parsing: %w", err)
	}
	if sect == nil {
		return nil, nil
	}
	sect.Link = url

	return sect, nil
}

func (h *JishoHandler) url(word string) string {
	return h.searchUrl + word
}

func (h *JishoHandler) parseResponse(resp *http.Response) (*JishoSection, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	wordSection := h.parseWordSection(doc)
	if wordSection == nil {
		return nil, nil
	}

	kanjis := h.parseKanjiSection(doc)

	return &JishoSection{
		WordSection: *wordSection,
		Kanjis:      kanjis,
	}, nil
}

func (h *JishoHandler) parseWordSection(doc *goquery.Document) *JishoWordSection {
	var wordSection JishoWordSection

	wordSectionEl := doc.Find(".concept_light").First()
	if wordSectionEl.Length() == 0 {
		return nil
	}
	readingsSection := wordSectionEl.Find(".concept_light-wrapper .concept_light-readings").First()
	meaningSection := wordSectionEl.Find(".meanings-wrapper").First()

	wordSection.Word = h.parseWord(readingsSection)
	wordSection.Reading = h.parseReading(readingsSection)
	wordSection.Meanings = h.parseMeanings(meaningSection)

	return &wordSection
}

func (h *JishoHandler) parseKanjiSection(doc *goquery.Document) []JishoKanji {
	kanjiSectionEl := doc.Find("#secondary .kanji_light_block").First()
	if kanjiSectionEl.Length() == 0 {
		return nil
	}
	return h.parseKanjis(kanjiSectionEl)
}

func (h *JishoHandler) parseKanjis(kanjiSection *goquery.Selection) []JishoKanji {
	var kanjis []JishoKanji

	kanjiSection.Find(".kanji_light_content").Each(func(_ int, el *goquery.Selection) {
		kanjiLink := el.Find(".literal_block .character a")

		kunyomisEl := el.Find(".kun").First()
		kunyomisEl.Find(".type").Remove()
		onyomisEl := el.Find(".on").First()
		onyomisEl.Find(".type").Remove()

		kanjis = append(kanjis, JishoKanji{
			Kanji: JishoWordWithLink{
				Link: kanjiLink.AttrOr("href", ""),
				Word: kanjiLink.Text(),
			},
			Meaning:  el.Find(".meanings").Text(),
			Kunyomis: h.parseKanjiReadings(kunyomisEl),
			Onyomis:  h.parseKanjiReadings(onyomisEl),
		})
	})

	return kanjis
}

func (h *JishoHandler) parseKanjiReadings(readingsEl *goquery.Selection) []JishoWordWithLink {
	var readings []JishoWordWithLink
	readingsEl.Find("a").Each(func(_ int, el *goquery.Selection) {
		readings = append(readings, JishoWordWithLink{
			Link: el.AttrOr("href", ""),
			Word: el.Text(),
		})
	})

	return readings
}

func (h *JishoHandler) parseWord(readingsSection *goquery.Selection) string {
	return strings.TrimSpace(readingsSection.Find(".text").Text())
}

func (h *JishoHandler) parseReading(readingsSection *goquery.Selection) string {
	return strings.TrimSpace(readingsSection.Find(".furigana").Text())
}

func (h *JishoHandler) parseMeanings(meaningSection *goquery.Selection) []JishoMeaning {
	var meanings []JishoMeaning

	lastTag := ""
	idx := 1
	meaningSection.Children().Each(func(i int, el *goquery.Selection) {
		if el.HasClass("meaning-tags") {
			lastTag = strings.TrimSpace(el.Text())
		} else if el.HasClass("meaning-wrapper") {
			var jishoMeaning JishoMeaning
			jishoMeaning.Meaning = strings.TrimSpace(el.Find(".meaning-meaning").Text())
			if lastTag != "" {
				t := lastTag
				jishoMeaning.Tags = &t
			}
			jishoMeaning.ListIdx = idx
			idx++
			meanings = append(meanings, jishoMeaning)
			lastTag = ""
		}
	})
	return meanings
}
