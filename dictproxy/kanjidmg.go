package dictproxy

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/zemiret/omnikanji"
	"github.com/zemiret/omnikanji/pkg/ptr"
)

var (
	KanjidmgNoKanjiErr = fmt.Errorf("kanji not found at kanjidamage")
)

func LoadKanjidmgLinks(httpClient HttpClient) (map[string]string, error) {
	resp, err := httpClient.Get(omnikanji.KanjidmgListUrl)
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
			links[link.Text()] = omnikanji.KanjidmgBaseUrl + href
		}
	})

	return links, nil
}

type Kanjidmg struct {
	links      map[string]string
	httpClient HttpClient
}

func NewKanjidmg(links map[string]string, httpClient HttpClient) *Kanjidmg {
	return &Kanjidmg{
		links:      links,
		httpClient: httpClient,
	}
}

func (h *Kanjidmg) Url(kanji string) string {
	return h.links[kanji]
}

func (h *Kanjidmg) Get(kanji rune) (*omnikanji.KanjidmgSection, error) {
	url := h.Url(string(kanji))
	if url == "" {
		return nil, KanjidmgNoKanjiErr
	}

	resp, err := h.httpClient.Get(url)
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

func (h *Kanjidmg) parseResponse(resp *http.Response, url string) (*omnikanji.KanjidmgSection, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	sect := &omnikanji.KanjidmgSection{}

	rows := doc.Find(".container").Last().Find(".row")
	wordSection := rows.Eq(1)
	wordSectionClone := wordSection.Clone()
	wordSectionClone.Find("h1").Remove()
	radicalsSection := wordSectionClone.Find(".col-md-8")
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

func (h *Kanjidmg) buildWordSection(wordSection *goquery.Selection, url string) (*omnikanji.KanjidmgKanji, error) {
	res, err := h.parseWordSection(wordSection)
	if err != nil {
		return nil, err
	}
	res.Link = url
	return res, nil
}

func (h *Kanjidmg) parseWordSection(wordSection *goquery.Selection) (*omnikanji.KanjidmgKanji, error) {
	kanjiCharSection := wordSection.Find("h1 .kanji_character")
	kanjiStr, kanjiImg, err := h.kanjiTextOrImage(kanjiCharSection)
	if err != nil {
		return nil, err
	}

	return &omnikanji.KanjidmgKanji{
		Kanji:      kanjiStr,
		KanjiImage: kanjiImg,
		Meaning:    wordSection.Find("h1 .translation").Text(),
	}, nil

}

func (h *Kanjidmg) kanjiTextOrImage(kanjiCharSection *goquery.Selection) (*string, *string, error) {
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

func (h *Kanjidmg) fetchKanjiImg(url string) (string, error) {
	resp, err := h.httpClient.Get(omnikanji.KanjidmgBaseUrl + "/" + url)
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

func (h *Kanjidmg) parseRadicals(radicalsSection *goquery.Selection) (radicals []omnikanji.KanjidmgKanji, err error) {
	radicalsSection.Find("h1").Remove()

	radicalsLinks := radicalsSection.Find("a")

	usedLinks := 0
	radicalsSection.Contents().Not("a").Each(func(_ int, m *goquery.Selection) {
		meaningText := h.trimNotCharacters(m.Text())
		if strings.TrimSpace(meaningText) != "" {

			kanjiCharSection := radicalsLinks.Eq(usedLinks)
			kanjiStr, kanjiImg, err := h.kanjiTextOrImage(kanjiCharSection)
			if err != nil {
				radicals = nil
				err = fmt.Errorf("parseRadicals: %w", err)
				return
			}

			radicals = append(radicals, omnikanji.KanjidmgKanji{
				Kanji:      kanjiStr,
				KanjiImage: kanjiImg,
				Meaning:    meaningText,
				Link:       omnikanji.KanjidmgBaseUrl + "/" + radicalsLinks.Eq(usedLinks).AttrOr("href", ""),
			})
			usedLinks += 1
		}
	})

	return
}

func (h *Kanjidmg) trimNotCharacters(text string) string {
	return strings.Trim(text, "+_-=,.:;'\"/|\\][() \t\n!?@$#%*")
}

func (h *Kanjidmg) parseOnyomi(contentSection *goquery.Selection) string {
	return h.parseContentRow(contentSection, "Onyomi")
}

func (h *Kanjidmg) parseMnemonic(contentSection *goquery.Selection) string {
	return h.parseContentRow(contentSection, "Mnemonic")
}

func (h *Kanjidmg) parseContentRow(section *goquery.Selection, sectionHeader string) string {
	return h.trimNotCharacters(section.Find("h2:contains(" + sectionHeader + ")").Next().Text())
}
