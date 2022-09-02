package server_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zemiret/omnikanji"
	"github.com/zemiret/omnikanji/dictproxy"
	"github.com/zemiret/omnikanji/jptext"
	"github.com/zemiret/omnikanji/server"
)

type HttpClientMock struct {
	staticDir string
}

func NewHttpClientMock(staticDir string) *HttpClientMock {
	return &HttpClientMock{
		staticDir: staticDir,
	}
}

func (c *HttpClientMock) Get(searchUrl string) (*http.Response, error) {
	var filePath string

	if strings.HasPrefix(searchUrl, omnikanji.JishoSearchUrl) {
		word := strings.TrimPrefix(searchUrl, omnikanji.JishoSearchUrl) // this could go onto "jisho" aggregate not to spill out its logic
		filePath = filepath.Join(c.staticDir, "jisho", word+".html")
	} else if strings.HasPrefix(searchUrl, omnikanji.KanjidmgBaseUrl) {
		word := strings.TrimPrefix(searchUrl, omnikanji.KanjidmgBaseUrl) // this could go onto "kanjidmg" aggregate not to spill out its logic
		filePath = filepath.Join(c.staticDir, "kanjidmg", word+".html")
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w, filepath: %s", err, filePath)
	}

	return &http.Response{
		Body: f,
	}, nil
}

func AreEqualJSON(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}

func TestServer(t *testing.T) {
	// TODO: Replace lookups with testCases
	// lookups := []string{
	// 	"no results",
	// 	"何",
	// 	"兄弟",
	// 	"路面電車停留場",
	// 	"あったり前",
	// 	"相変わらず",
	// 	"ペラペラ",
	// 	"driver's licence",
	// }

	type TestCase struct {
		word   string
		expect func(t *testing.T, res *server.TemplateParams)

		expectJson string
	}

	testCases := []*TestCase{
		{
			word: "noresults",
			expect: func(t *testing.T, res *server.TemplateParams) {
				if res.Jisho != nil || res.Kanjidmg != nil {
					t.Error("strange, expected Jisho = nil and Kanjidmg = nil")
				}
			},
		},
		{
			word: "何",
			expectJson: `{
				"EnglishSearchedWord": "",
				"JishoEnglishWordLink": "",
				"Jisho": {
				  "Link": "https://jisho.org/search/何",
				  "WordSection": {
					"FullWord": "何",
					"Parts": [
					  {
						"MainText": "何",
						"Reading": "なに"
					  }
					],
					"Meanings": [
					  {
						"ListIdx": 1,
						"Meaning": "what",
						"Tags": "Pronoun"
					  },
					  {
						"ListIdx": 2,
						"Meaning": "you-know-what; that thing",
						"Tags": "Pronoun"
					  },
					  {
						"ListIdx": 3,
						"Meaning": "whatsit; whachamacallit; what's-his-name; what's-her-name",
						"Tags": "Pronoun"
					  },
					  {
						"ListIdx": 4,
						"Meaning": "penis; (one's) thing; dick",
						"Tags": "Noun"
					  },
					  {
						"ListIdx": 5,
						"Meaning": "(not) at all; (not) in the slightest",
						"Tags": "Adverb (fukushi)"
					  },
					  {
						"ListIdx": 6,
						"Meaning": "what?; huh?",
						"Tags": null
					  },
					  {
						"ListIdx": 7,
						"Meaning": "hey!; come on!",
						"Tags": null
					  },
					  {
						"ListIdx": 8,
						"Meaning": "oh, no (it's fine); why (it's nothing); oh (certainly not)",
						"Tags": null
					  },
					  {
						"ListIdx": 9,
						"Meaning": "ナニ",
						"Tags": "Other forms"
					  }
					]
				  },
				  "Kanjis": [
					{
					  "Kanji": {
						"Link": "//jisho.org/search/%E4%BD%95%20%23kanji",
						"Word": "何"
					  },
					  "Meaning": "\n            what\n      ",
					  "Kunyomis": [
						{
						  "Link": "//jisho.org/search/%E4%BD%95%20%E3%81%AA%E3%81%AB",
						  "Word": "なに"
						},
						{
						  "Link": "//jisho.org/search/%E4%BD%95%20%E3%81%AA%E3%82%93",
						  "Word": "なん"
						},
						{
						  "Link": "//jisho.org/search/%E4%BD%95%20%E3%81%AA%E3%81%AB",
						  "Word": "なに-"
						},
						{
						  "Link": "//jisho.org/search/%E4%BD%95%20%E3%81%AA%E3%82%93",
						  "Word": "なん-"
						}
					  ],
					  "Onyomis": [
						{
						  "Link": "//jisho.org/search/%E4%BD%95%20%E3%81%8B",
						  "Word": "カ"
						}
					  ]
					}
				  ]
				},
				"Kanjidmg": [
				  {
					"WordSection": {
					  "Kanji": "何",
					  "KanjiImage": null,
					  "Meaning": "what?!?",
					  "Link": "http://www.kanjidamage.com何"
					},
					"Radicals": null,
					"Onyomi": "KA, but you don't need to learn it",
					"Mnemonic": "When you say \" WHAAT????\", you are asking that person if what they just said is really possible"
				  }
				],
				"Error": null
			  }`,
		},
		{
			word: "兄弟",
			expectJson: `{
				"EnglishSearchedWord": "",
				"JishoEnglishWordLink": "",
				"Jisho": {
				  "Link": "https://jisho.org/search/兄弟",
				  "WordSection": {
					"FullWord": "兄弟",
					"Parts": [
					  {
						"MainText": "兄",
						"Reading": "きょう"
					  },
					  {
						"MainText": "弟",
						"Reading": "だい"
					  }
					],
					"Meanings": [
					  {
						"ListIdx": 1,
						"Meaning": "siblings; brothers and sisters",
						"Tags": "Noun"
					  },
					  {
						"ListIdx": 2,
						"Meaning": "brothers",
						"Tags": "Noun"
					  },
					  {
						"ListIdx": 3,
						"Meaning": "siblings-in-law; brothers-in-law; sisters-in-law",
						"Tags": "Noun"
					  },
					  {
						"ListIdx": 4,
						"Meaning": "mate; friend",
						"Tags": "Noun"
					  },
					  {
						"ListIdx": 5,
						"Meaning": "兄弟 【けいてい】",
						"Tags": "Other forms"
					  }
					]
				  },
				  "Kanjis": [
					{
					  "Kanji": {
						"Link": "//jisho.org/search/%E5%85%84%20%23kanji",
						"Word": "兄"
					  },
					  "Meaning": "\n            elder brother, \n            big brother\n      ",
					  "Kunyomis": [
						{
						  "Link": "//jisho.org/search/%E5%85%84%20%E3%81%82%E3%81%AB",
						  "Word": "あに"
						}
					  ],
					  "Onyomis": [
						{
						  "Link": "//jisho.org/search/%E5%85%84%20%E3%81%91%E3%81%84",
						  "Word": "ケイ"
						},
						{
						  "Link": "//jisho.org/search/%E5%85%84%20%E3%81%8D%E3%82%87%E3%81%86",
						  "Word": "キョウ"
						}
					  ]
					},
					{
					  "Kanji": {
						"Link": "//jisho.org/search/%E5%BC%9F%20%23kanji",
						"Word": "弟"
					  },
					  "Meaning": "\n            younger brother, \n            faithful service to elders\n      ",
					  "Kunyomis": [
						{
						  "Link": "//jisho.org/search/%E5%BC%9F%20%E3%81%8A%E3%81%A8%E3%81%86%E3%81%A8",
						  "Word": "おとうと"
						}
					  ],
					  "Onyomis": [
						{
						  "Link": "//jisho.org/search/%E5%BC%9F%20%E3%81%A6%E3%81%84",
						  "Word": "テイ"
						},
						{
						  "Link": "//jisho.org/search/%E5%BC%9F%20%E3%81%A0%E3%81%84",
						  "Word": "ダイ"
						},
						{
						  "Link": "//jisho.org/search/%E5%BC%9F%20%E3%81%A7",
						  "Word": "デ"
						}
					  ]
					}
				  ]
				},
				"Kanjidmg": [
				  {
					"WordSection": {
					  "Kanji": "兄",
					  "KanjiImage": null,
					  "Meaning": "older brother",
					  "Link": "http://www.kanjidamage.com兄"
					},
					"Radicals": null,
					"Onyomi": "KYOU / KEI\n\n\nas in \"KYOU (今日)  my older brother acted OK. Tomorrow he'll be a bully again",
					"Mnemonic": "My older brother is eating so much, he is basically a mouth on legs"
				  },
				  {
					"WordSection": {
					  "Kanji": "弟",
					  "KanjiImage": null,
					  "Meaning": "younger brother",
					  "Link": "http://www.kanjidamage.com弟"
					},
					"Radicals": null,
					"Onyomi": "TEI, DAI",
					"Mnemonic": "When my younger brother gets horny he TAKES a bow like Cupid and shoots you and you DIE. \nHe's passionate but unfortunately not gifted with a sense of poetic metaphor"
				  }
				],
				"Error": null
			  }`,
		},
		{
			word: "路面電車停留場", // TODO: Huh. We got nil jisho for this!!!
		},
	}

	log.Println("elo")
	// expect: func(t *testing.T, res *server.TemplateParams) {
	// 	b, _ := json.MarshalIndent(res, "", "  ")
	// 	log.Println(string(b))
	// 	t.Fail()
	// },

	kanjidmgLinks := make(map[string]string)
	for _, tc := range testCases {
		for _, r := range tc.word {
			if !jptext.IsKanji(r) {
				continue
			}

			rStr := string(r)
			kanjidmgLinks[rStr] = omnikanji.KanjidmgBaseUrl + rStr
		}
	}

	for _, tc := range testCases {
		t.Run(tc.word, func(t *testing.T) {
			fixtureDir, err := filepath.Abs("fixture")
			if err != nil {
				t.FailNow()
			}

			httpClient := NewHttpClientMock(fixtureDir)
			jisho := dictproxy.NewJisho(omnikanji.JishoSearchUrl, httpClient)
			kanjidmg := dictproxy.NewKanjidmg(kanjidmgLinks, httpClient)
			srv := server.NewServer(jisho, kanjidmg)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:8080/search/?word=%s", tc.word), nil)

			data := srv.HandleIndex(nil, req)

			if tc.expect == nil && tc.expectJson == "" {
				t.Error("No expects? You sure?")
			}

			if tc.expectJson != "" {
				dataB, err := json.MarshalIndent(data, "", "  ")
				areEqual, err := AreEqualJSON(string(dataB), tc.expectJson)
				if err != nil {
					t.Errorf("AreEqualJSON: %s", err)
				}

				if !areEqual {
					t.Error("received response != expectedJson")
				}
			}

			if tc.expect != nil {
				tc.expect(t, data)
			}
		})
	}
}
