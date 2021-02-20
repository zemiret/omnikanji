package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/zemiret/omnikanji/ptr"
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

	// TODO: HANDLING KANJIS THAT ARE IMAGES NOT TEXT
	// TODO: Top comment
	sect.WordSection = h.buildWordSection(wordSection, url)
	sect.Radicals = h.parseRadicals(radicalsSection)
	sect.Onyomi = ptr.String(h.parseOnyomi(contentSection))
	sect.Mnemonic = ptr.String(h.parseMnemonic(contentSection))

	return sect, nil
}

func (h *KanjidmgHandler) buildWordSection(wordSection *goquery.Selection, url string) KanjidmgKanji {
	res := h.parseWordSection(wordSection)
	res.Link = url
	return res
}

func (h *KanjidmgHandler) parseWordSection(wordSection *goquery.Selection) KanjidmgKanji {
	return KanjidmgKanji{
		Kanji:   wordSection.Find("h1 .kanji_character").Text(),
		Meaning: wordSection.Find("h1 .translation").Text(),
	}
}

func (h *KanjidmgHandler) parseRadicals(radicalsSection *goquery.Selection) []KanjidmgKanji {
	var radicals []KanjidmgKanji
	radicalsLinks := radicalsSection.Find("a")

	usedLinks := 0
	radicalsSection.Contents().Not("a").Each(func(_ int, m *goquery.Selection) {
		meaningText := m.Text() // TODO: text cleanup
		if strings.TrimSpace(meaningText) != "" {

			radicals = append(radicals, KanjidmgKanji{
				Kanji:   radicalsLinks.Eq(usedLinks).Text(),
				Meaning: m.Text(),
				Link:    radicalsLinks.Eq(usedLinks).AttrOr("href", ""),
			})
			usedLinks += 1
		}
	})

	return radicals
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
