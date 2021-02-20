package main

import (
	"log"
	"net/http"
	"strings"
)

type server struct {
	Logger
	kanjidmgLinks map[string]string
	jisho         *JishoHandler
	kanjidmg      *KanjidmgHandler
}

func newServer(jisho *JishoHandler, kanjidmg *KanjidmgHandler) *server {
	return &server{
		jisho:    jisho,
		kanjidmg: kanjidmg,
	}
}

type TemplateParams struct {
	Jisho    *JishoSection
	Kanjidmg []*KanjidmgSection
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

	log.Println("Gonna search for: " + word)
	// TODO: Validation if it's in japanese
	// TODO: Split search if it's kana (do not do kanjidamage then)

	s.renderTemplate(w, s.getSections(word))
}

func (s *server) getSections(word string) *TemplateParams {
	var tParams TemplateParams

	// TODO: This could go in parallel
	jishoSection, err := s.jisho.GetSection(word)
	if err == nil {
		tParams.Jisho = jishoSection
	} else {
		s.Errorf(err, "error getting jisho section")
	}

	for _, c := range word {
		sect, err := s.kanjidmg.GetSection(c)
		if err == nil {
			tParams.Kanjidmg = append(tParams.Kanjidmg, sect)
		} else {
			s.Errorf(err, "error getting kanjidmg section")
		}
	}

	return &tParams
}

func (s *server) renderTemplate(w http.ResponseWriter, data *TemplateParams) {
	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
