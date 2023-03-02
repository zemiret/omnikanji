package dictproxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/zemiret/omnikanji"
	"github.com/zemiret/omnikanji/jptext"
)

type Jisho struct {
	searchUrl  string
	httpClient HttpClient
}

func NewJisho(jishoSearchUrl string, httpClient HttpClient) *Jisho {
	return &Jisho{
		searchUrl:  jishoSearchUrl,
		httpClient: httpClient,
	}
}

func (h *Jisho) Get(word string) (*omnikanji.JishoSection, error) {
	url := h.Url(word)

	resp, err := h.httpClient.Get(url)
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

	// There are some jisho words that do not split niceliy into parts. Fallback
	if sect.WordSection.FullWord == "" && jptext.IsJapaneseWord(word) {
		sect.WordSection.FullWord = word
	}

	return sect, nil
}

func (h *Jisho) Url(word string) string {
	return h.searchUrl + word
}

func (h *Jisho) parseResponse(resp *http.Response) (*omnikanji.JishoSection, error) {
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	wordSection := h.parseWordSection(doc)
	if wordSection == nil {
		return nil, nil
	}

	kanjis := h.parseKanjiSection(doc)

	return &omnikanji.JishoSection{
		WordSection: *wordSection,
		Kanjis:      kanjis,
	}, nil
}

func (h *Jisho) parseWordSection(doc *goquery.Document) *omnikanji.JishoWordSection {
	var wordSection omnikanji.JishoWordSection

	wordSectionEl := doc.Find(".concept_light").First()
	if wordSectionEl.Length() == 0 {
		return nil
	}
	readingsSection := wordSectionEl.Find(".concept_light-wrapper .concept_light-readings").First()
	meaningSection := wordSectionEl.Find(".meanings-wrapper").First()

	wordSection.FullWord, wordSection.Parts = h.parseWordParts(readingsSection)
	wordSection.Meanings = h.parseMeanings(meaningSection)

	return &wordSection
}

func (h *Jisho) parseWordParts(wordSection *goquery.Selection) (string, []omnikanji.JishoWordPart) {
	fullWord := h.parseWord(wordSection)
	var furiganaInParts []string
	wordSection.Find(".furigana .kanji").Each(func(_ int, el *goquery.Selection) {
		furiganaInParts = append(furiganaInParts, el.Text())
	})

	kanjisCountInWord := jptext.KanjisCountInAWord(fullWord)

	// Some words on jisho use different html structure. Fallback for such cases
	if len(furiganaInParts) != kanjisCountInWord {
		// TODO: sth is messed up :/ Some fallback in such cases
		// TODO: Try "ruby" parsing. If that still fails, return none
		furiganaInParts = []string{}

		furiganaRuby := wordSection.Find(".furigana ruby")
		furiganaRbStr := furiganaRuby.Find("rb").Text()
		furiganaRtStr := furiganaRuby.Find("rt").Text()

		if len(furiganaRtStr) == 0 {
			return fullWord, nil
		}

		// each kanji has 1 kana to it
		if len(furiganaRbStr) == len(furiganaRtStr) {
			//var wordParts []omnikanji.JishoWordPart
			//for i, rbKanji := range furiganaRbStr {
			//	//				rtReading :=
			//}
		} else {
			// Dunno how I could figure out which kana belongs to which kanji
			return fullWord, nil
		}

		if len(furiganaInParts) != kanjisCountInWord {
			return fullWord, nil
		}
	}

	var wordParts []omnikanji.JishoWordPart
	kanaPart := ""
	kanjiIdx := 0

	for _, c := range fullWord {
		if jptext.IsKanji(c) {
			if kanaPart != "" {
				wordParts = append(wordParts, omnikanji.JishoWordPart{
					MainText: kanaPart,
				})
				kanaPart = ""
			}

			wordParts = append(wordParts, omnikanji.JishoWordPart{
				MainText: string(c),
				Reading:  furiganaInParts[kanjiIdx],
			})
			kanjiIdx++
		} else {
			kanaPart += string(c)
		}
	}

	if kanaPart != "" {
		wordParts = append(wordParts, omnikanji.JishoWordPart{
			MainText: kanaPart,
		})
	}

	return fullWord, wordParts
}

func (h *Jisho) parseWord(readingsSection *goquery.Selection) string {
	return strings.TrimSpace(readingsSection.Find(".text").Text())
}

func (h *Jisho) parseKanjiSection(doc *goquery.Document) []omnikanji.JishoKanji {
	kanjiSectionEl := doc.Find("#secondary .kanji_light_block").First()
	if kanjiSectionEl.Length() == 0 {
		return nil
	}
	return h.parseKanjis(kanjiSectionEl)
}

func (h *Jisho) parseKanjis(kanjiSection *goquery.Selection) []omnikanji.JishoKanji {
	var kanjis []omnikanji.JishoKanji

	kanjiSection.Find(".kanji_light_content").Each(func(_ int, el *goquery.Selection) {
		kanjiLink := el.Find(".literal_block .character a")

		kunyomisEl := el.Find(".kun").First()
		kunyomisEl.Find(".type").Remove()
		onyomisEl := el.Find(".on").First()
		onyomisEl.Find(".type").Remove()

		kanjis = append(kanjis, omnikanji.JishoKanji{
			Kanji: omnikanji.JishoWordWithLink{
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

func (h *Jisho) parseKanjiReadings(readingsEl *goquery.Selection) []omnikanji.JishoWordWithLink {
	var readings []omnikanji.JishoWordWithLink
	readingsEl.Find("a").Each(func(_ int, el *goquery.Selection) {
		readings = append(readings, omnikanji.JishoWordWithLink{
			Link: el.AttrOr("href", ""),
			Word: el.Text(),
		})
	})

	return readings
}

func (h *Jisho) parseMeanings(meaningSection *goquery.Selection) []omnikanji.JishoMeaning {
	var meanings []omnikanji.JishoMeaning

	lastTag := ""
	idx := 1
	meaningSection.Children().Each(func(i int, el *goquery.Selection) {
		if el.HasClass("meaning-tags") {
			lastTag = strings.TrimSpace(el.Text())
		} else if el.HasClass("meaning-wrapper") {
			var jishoMeaning omnikanji.JishoMeaning
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
