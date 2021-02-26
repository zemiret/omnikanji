package main

import (
	"encoding/base64"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/zemiret/omnikanji/ptr"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	KanjidmgNoKanjiErr = fmt.Errorf("kanji not found at kanjidamage")
)

type KanjidmgHandler struct {
	links map[string]string
}

func NewKanjidmgHandler(links map[string]string) *KanjidmgHandler {
	return &KanjidmgHandler{
		links: links,
	}
}

type KanjidmgSection struct {
	WordSection KanjidmgKanji
	TopComment  *string
	Radicals    []KanjidmgKanji
	Onyomi      *string
	Mnemonic    *string
	Mutants     []KanjidmgKanji

	// Kunyomi TODO
	//Jukugo  TODO
	// UsedIn TODO
	// Synonyms TODO
	// Lookalikes TODO
}

type KanjidmgKanji struct {
	Kanji      *string
	KanjiImage *string
	Meaning    string
	Link       string
}

func (h *KanjidmgHandler) url(kanji string) string {
	return h.links[kanji]
}

func (h *KanjidmgHandler) GetSection(kanji rune) (*KanjidmgSection, error) {
	url := h.url(string(kanji))
	if url == "" {
		return nil, KanjidmgNoKanjiErr
	}

	resp, err := http.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	sect, err := h.parseResponse(resp, url)
	if err != nil {
		return nil, fmt.Errorf("parsing: %w", err)
	}

	return sect, nil
}

func (h *KanjidmgHandler) parseResponse(resp *http.Response, url string) (*KanjidmgSection, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	sect := &KanjidmgSection{}

	rows := doc.Find(".container").Last().Find(".row")
	wordSection := rows.Eq(1)
	wordSectionClone := wordSection.Clone()
	wordSectionClone.Find("h1").Remove()
	radicalsSection := wordSectionClone.Find("div.span8")
	contentSection := rows.Eq(2)

	// TODO: Top comment
	parsedWordSection, err := h.buildWordSection(wordSection, url)
	if err != nil {
		return nil, err
	}
	sect.WordSection = *parsedWordSection
	parsedRadicalsSection, err := h.parseRadicals(radicalsSection)
	if err != nil {
		return nil, err
	}
	sect.Radicals = parsedRadicalsSection
	sect.Onyomi = ptr.String(h.parseOnyomi(contentSection))
	sect.Mnemonic = ptr.String(h.parseMnemonic(contentSection))

	return sect, nil
}

func (h *KanjidmgHandler) buildWordSection(wordSection *goquery.Selection, url string) (*KanjidmgKanji, error) {
	res, err := h.parseWordSection(wordSection)
	if err != nil {
		return nil, err
	}
	res.Link = url
	return res, nil
}

func (h *KanjidmgHandler) parseWordSection(wordSection *goquery.Selection) (*KanjidmgKanji, error) {
	kanjiCharSection := wordSection.Find("h1 .kanji_character")
	kanjiStr, kanjiImg, err := h.kanjiTextOrImage(kanjiCharSection)
	if err != nil {
		return nil, err
	}

	return &KanjidmgKanji{
		Kanji:      kanjiStr,
		KanjiImage: kanjiImg,
		Meaning:    wordSection.Find("h1 .translation").Text(),
	}, nil

}

func (h *KanjidmgHandler) kanjiTextOrImage(kanjiCharSection *goquery.Selection) (*string, *string, error) {
	var kanjiStr, kanjiImg string
	var err error

	kanjiStr = kanjiCharSection.Text()
	if kanjiStr == "" {
		url := kanjiCharSection.Find("img").AttrOr("src", "")
		if url == "" {
			return nil, nil, fmt.Errorf("cannot parse word section - there does not seem to be kanji in text nor in imagr")
		}

		kanjiImg, err = h.fetchKanjiImg(url)
		if err != nil {
			return nil, nil, fmt.Errorf("parseWordSection: %w", err)
		}
	}

	return ptr.String(kanjiStr), ptr.String(kanjiImg), nil
}

func (h *KanjidmgHandler) fetchKanjiImg(url string) (string, error) {
	resp, err := http.Get(kanjidmgBaseUrl + "/" + url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", fmt.Errorf("error fetching kanji image: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading kanji image response: %w", err)
	}

	encodedImg := base64.StdEncoding.EncodeToString(body)
	return encodedImg, nil
}

func (h *KanjidmgHandler) parseRadicals(radicalsSection *goquery.Selection) (radicals []KanjidmgKanji, err error) {
	radicalsLinks := radicalsSection.Find("a")

	usedLinks := 0
	radicalsSection.Contents().Not("a").Each(func(_ int, m *goquery.Selection) {
		meaningText := m.Text() // TODO: text cleanup
		if strings.TrimSpace(meaningText) != "" {

			kanjiCharSection := radicalsLinks.Eq(usedLinks)
			kanjiStr, kanjiImg, err := h.kanjiTextOrImage(kanjiCharSection)
			if err != nil {
				radicals = nil
				err = fmt.Errorf("parseRadicals: %w", err)
				return
			}

			radicals = append(radicals, KanjidmgKanji{
				Kanji:      kanjiStr,
				KanjiImage: kanjiImg,
				Meaning:    m.Text(),
				Link:       radicalsLinks.Eq(usedLinks).AttrOr("href", ""),
			})
			usedLinks += 1
		}
	})

	return
}

func (h *KanjidmgHandler) parseOnyomi(contentSection *goquery.Selection) string {
	return h.parseContentRow(contentSection, "Onyomi")
}

func (h *KanjidmgHandler) parseMnemonic(contentSection *goquery.Selection) string {
	return h.parseContentRow(contentSection, "Mnemonic")
}

func (h *KanjidmgHandler) parseContentRow(section *goquery.Selection, sectionHeader string) string {
	return section.Find("h2:contains(" + sectionHeader + ")").Next().Text()
}
