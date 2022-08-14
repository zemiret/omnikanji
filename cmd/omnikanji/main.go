package main

import (
	"log"

	"github.com/zemiret/omnikanji"
	"github.com/zemiret/omnikanji/dictproxy"
	"github.com/zemiret/omnikanji/pkg/http"
	"github.com/zemiret/omnikanji/server"
)

// TODO: Periodic refresh of kanjidmg list of kanjis (once every month is probably enough)

func main() {
	httpClient := http.NewClient()

	kanjidmgLinks, err := dictproxy.LoadKanjidmgLinks(httpClient)
	if err != nil {
		log.Fatal("error getting kanjidamage kanji list: " + err.Error())
	}

	jisho := dictproxy.NewJisho(omnikanji.JishoSearchUrl, httpClient)
	kanjidmg := dictproxy.NewKanjidmg(kanjidmgLinks, httpClient)
	srv := server.NewServer(jisho, kanjidmg)
	srv.Start()
}
