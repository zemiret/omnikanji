package omnikanji

import (
	"log"
	"os"
)

type Config struct {
	DebugMode bool
}

func ParseEnvConfig() *Config {
	cfg := &Config{}
	log.Println("Parsing env config...")
	if os.Getenv("DEBUG") != "" {
		log.Println("DEBUG=true")
		cfg.DebugMode = true
	}
	log.Println("Config parsed.")

	return cfg
}

const (
	JishoSearchUrl = "https://jisho.org/search/"

	KanjidmgBaseUrl = "http://www.kanjidamage.com"
	KanjidmgListUrl = KanjidmgBaseUrl + "/kanji"

	QuerySearchKey = "word"
)
