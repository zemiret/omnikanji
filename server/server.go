package server

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/zemiret/omnikanji"
	"github.com/zemiret/omnikanji/jptext"
	"github.com/zemiret/omnikanji/pkg/logger"
)

var templates = template.Must(template.ParseFiles("server/index.html"))

type JishoSectionGetter interface {
	Url(word string) string
	Get(word string) (*omnikanji.JishoSection, error)
}

type KanjidmgSectionGetter interface {
	Get(kanji rune) (*omnikanji.KanjidmgSection, error)
}

type server struct {
	*logger.Logger
	kanjidmgLinks map[string]string
	jisho         JishoSectionGetter
	kanjidmg      KanjidmgSectionGetter
}

type TemplateParams struct {
	EnglishSearchedWord  string
	JishoEnglishWordLink string
	Jisho                *omnikanji.JishoSection
	Kanjidmg             []*omnikanji.KanjidmgSection
	Error                *string
}

func NewServer(jisho JishoSectionGetter, kanjidmg KanjidmgSectionGetter) *server {
	return &server{
		jisho:    jisho,
		kanjidmg: kanjidmg,
		Logger:   logger.NewLogger(),
	}
}

func (s *server) Start() {
	http.HandleFunc("/", s.handleIndex)
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	log.Println("Starting server at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/search/") {
		s.handleSearch(w, r)
		return
	}

	s.renderTemplate(w, nil)
}

func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	word := r.URL.Query().Get("word")
	if word == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if !jptext.IsJapaneseWord(word) {
		data := s.searchFromEnglish(word)
		s.renderTemplate(w, data)
		return
	}

	data := s.searchFromJapanese(word)
	s.renderTemplate(w, data)
}

func (s *server) searchFromEnglish(word string) *TemplateParams {
	var wg sync.WaitGroup
	var tParams TemplateParams
	s.doJishoSearch(&wg, &tParams, word)
	wg.Wait()

	if tParams.Jisho == nil {
		return &TemplateParams{}
	}

	var wordKanjis string
	for _, c := range tParams.Jisho.WordSection.FullWord {
		if jptext.IsKanji(c) {
			wordKanjis += string(c)
		}
	}

	if wordKanjis != "" {
		s.doKanjidmgSearch(&tParams, wordKanjis)
	}

	tParams.EnglishSearchedWord = word
	tParams.JishoEnglishWordLink = s.jisho.Url(word)
	tParams.Jisho.Link = s.jisho.Url(tParams.Jisho.WordSection.FullWord) // overwrite english word link

	return &tParams
}

func (s *server) searchFromJapanese(word string) *TemplateParams {
	data := s.getSections(word, true)
	return data
}

func (s *server) getSections(word string, parseKanjis bool) *TemplateParams {
	var tParams TemplateParams
	var wg sync.WaitGroup

	s.doJishoSearch(&wg, &tParams, word)

	if parseKanjis {
		wordKanjis := jptext.ExtractKanjis(word)
		if wordKanjis != "" {
			s.doKanjidmgSearch(&tParams, wordKanjis)
		}
	}

	wg.Wait()
	return &tParams
}

func (s *server) doJishoSearch(wg *sync.WaitGroup, tParams *TemplateParams, word string) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		jishoSection, err := s.jisho.Get(word)
		if err != nil {
			s.Errorf(err, "error getting jisho section")
			return
		}
		tParams.Jisho = jishoSection
	}()
}

func (s *server) doKanjidmgSearch(tParams *TemplateParams, word string) {
	var wg sync.WaitGroup

	results := make([]*omnikanji.KanjidmgSection, utf8.RuneCountInString(word))
	idx := 0
	for _, c := range word {
		wg.Add(1)
		go func(i int, c rune) {
			defer wg.Done()
			sect, err := s.kanjidmg.Get(c)
			if err != nil {
				s.Errorf(err, "error getting kanjidmg section")
				return
			}
			results[i] = sect
		}(idx, c)
		idx++
	}

	wg.Wait()
	for _, r := range results {
		if r != nil {
			tParams.Kanjidmg = append(tParams.Kanjidmg, r)
		}
	}
}

func (s *server) errorParams(msg string) *TemplateParams {
	return &TemplateParams{Error: &msg}
}

func (s *server) renderTemplate(w http.ResponseWriter, data *TemplateParams) {
	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
