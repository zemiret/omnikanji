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
	Meaning string
	Tags    *string
	//MeaningSentence  *string
	//SupplementalInfo *string
}

type JishoKanji struct {
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

	var wordSection JishoWordSection

	wordSectionEl := doc.Find(".concept_light").First()
	readingsSection := wordSectionEl.Find(".concept_light-wrapper .concept_light-readings").First()
	meaningSection := wordSectionEl.Find(".meanings-wrapper").First()

	wordSection.Word = h.parseWord(readingsSection)
	wordSection.Reading = h.parseReading(readingsSection)
	wordSection.Meanings = h.parseMeanings(meaningSection)

	return &JishoSection{
		WordSection: wordSection,
		Kanjis:      nil, // TODO: Kanjis
	}, nil
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
			meanings = append(meanings, jishoMeaning)
			lastTag = ""
		}
	})
	return meanings
}
