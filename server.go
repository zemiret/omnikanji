package main

import (
	"github.com/zemiret/omnikanji/jptext"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"unicode/utf8"
)


type server struct {
	*Logger
	kanjidmgLinks map[string]string
	jisho         *JishoHandler
	kanjidmg      *KanjidmgHandler
}

func newServer(jisho *JishoHandler, kanjidmg *KanjidmgHandler) *server {
	return &server{
		jisho:    jisho,
		kanjidmg: kanjidmg,
		Logger:   NewLogger(),
	}
}

type TemplateParams struct {
	Jisho    *JishoSection
	Kanjidmg []*KanjidmgSection
	Error    *string
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
		s.renderTemplate(w, s.errorParams("IT'S NOT JAPANESE YOU MORON :|"))
		return
	}

	s.Printf("Searching for: %s\n", word)

	data := s.getSections(word)
	s.renderTemplate(w, data)
}

func (s *server) getSections(word string) *TemplateParams {
	var tParams TemplateParams
	var wg sync.WaitGroup

	s.doJishoSearch(&wg, &tParams, word)

	wordKanjis := jptext.ExtractKanjis(word)
	if wordKanjis != "" {
		s.doKanjidmgSearch(&wg, &tParams, wordKanjis)
	}

	wg.Wait()
	return &tParams
}

func (s *server) doJishoSearch(wg *sync.WaitGroup, tParams *TemplateParams, word string) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		jishoSection, err := s.jisho.GetSection(word)
		if err == nil {
			tParams.Jisho = jishoSection
		} else {
			s.Errorf(err, "error getting jisho section")
		}
	}()
}

func (s *server) doKanjidmgSearch(wg *sync.WaitGroup, tParams *TemplateParams, word string) {
	tParams.Kanjidmg = make([]*KanjidmgSection, utf8.RuneCountInString(word))
	idx := 0
	for _, c := range word {
		wg.Add(1)
		go func(i int, c rune) {
			defer wg.Done()
			sect, err := s.kanjidmg.GetSection(c)
			if err == nil {
				tParams.Kanjidmg[i] = sect
			} else {
				s.Errorf(err, "error getting kanjidmg section")
			}
		}(idx, c)
		idx++
	}
}

func (s *server) errorParams(msg string) *TemplateParams {
	return &TemplateParams{Error: &msg}
}

func (s *server) renderTemplate(w http.ResponseWriter, data *TemplateParams) {
	var templates = template.Must(template.ParseFiles("index.html"))

	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
