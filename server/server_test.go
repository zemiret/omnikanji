package server_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zemiret/omnikanji"
	"github.com/zemiret/omnikanji/dictproxy"
	"github.com/zemiret/omnikanji/jptext"
	"github.com/zemiret/omnikanji/server"

	"github.com/stretchr/testify/require"
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
	log.Println("Mock Get url: ", searchUrl)

	var filePath string

	if strings.HasPrefix(searchUrl, omnikanji.JishoSearchUrl) {
		word := strings.TrimPrefix(searchUrl, omnikanji.JishoSearchUrl) // this could go onto "jisho" aggregate not to spill out its logic
		filePath = filepath.Join(c.staticDir, "jisho", word+".html")
	} else if strings.HasPrefix(searchUrl, omnikanji.KanjidmgBaseUrl) {
		urlPath := strings.TrimPrefix(searchUrl, omnikanji.KanjidmgBaseUrl) // this could go onto "kanjidmg" aggregate not to spill out its logic

		// Mock image return for image radicals
		if strings.HasPrefix(urlPath, "//assets") {
			return &http.Response{
				Body: ioutil.NopCloser(strings.NewReader("imagebytes")),
			}, nil
		}

		word := urlPath
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

func TestServer(t *testing.T) {
	type TestCase struct {
		word   string
		expect func(t *testing.T, res *server.TemplateParams)

		expectJSON string
	}

	testCases := []*TestCase{
		{
			word: "noresults",
			expect: func(t *testing.T, res *server.TemplateParams) {
				require.Nil(t, res.Jisho)
				require.Nil(t, res.Kanjidmg)
			},
		},
		{
			word: "何",
			expectJSON: `{
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
					"Radicals": [
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "personleft",
						"Link": "http://www.kanjidamage.com//kanji/61-person-%E4%BA%BA"
					  },
					  {
						"Kanji": "可",
						"KanjiImage": null,
						"Meaning": "possible",
						"Link": "http://www.kanjidamage.com//kanji/55-possible-%E5%8F%AF"
					  }
					],
					"Onyomi": "KA, but you don't need to learn it",
					"Mnemonic": "When you say \" WHAAT????\", you are asking that person if what they just said is really possible"
				  }
				],
				"Error": null
			  }`,
		},
		{
			word: "兄弟",
			expectJSON: `{
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
					"Radicals": [
					  {
						"Kanji": "口",
						"KanjiImage": null,
						"Meaning": "mouth/small box radical",
						"Link": "http://www.kanjidamage.com//kanji/9-mouth-small-box-radical-%E5%8F%A3"
					  },
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "human legs",
						"Link": "http://www.kanjidamage.com//kanji/519-human-legs"
					  }
					],
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
					"Radicals": [
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "horny",
						"Link": "http://www.kanjidamage.com//kanji/859-horny"
					  },
					  {
						"Kanji": "弓",
						"KanjiImage": null,
						"Meaning": "bow",
						"Link": "http://www.kanjidamage.com//kanji/892-bow-%E5%BC%93"
					  },
					  {
						"Kanji": "  丶",
						"KanjiImage": null,
						"Meaning": "dot",
						"Link": "http://www.kanjidamage.com//kanji/1769-dot-%E4%B8%B6"
					  }
					],
					"Onyomi": "TEI, DAI",
					"Mnemonic": "When my younger brother gets horny he TAKES a bow like Cupid and shoots you and you DIE. \nHe's passionate but unfortunately not gifted with a sense of poetic metaphor"
				  }
				],
				"Error": null
			  }`,
		},
		{
			word: "路面電車停留場",
			expectJSON: `{
				"EnglishSearchedWord": "",
				"JishoEnglishWordLink": "",
				"Jisho": {
				  "Link": "https://jisho.org/search/路面電車停留場",
				  "WordSection": {
					"FullWord": "路面電車停留場",
					"Parts": null,
					"Meanings": [
					  {
						"ListIdx": 1,
						"Meaning": "Tram stop",
						"Tags": "Wikipedia definition"
					  }
					]
				  },
				  "Kanjis": [
					{
					  "Kanji": {
						"Link": "//jisho.org/search/%E9%9D%A2%20%23kanji",
						"Word": "面"
					  },
					  "Meaning": "\n            mask, \n            face, \n            features, \n            surface\n      ",
					  "Kunyomis": [
						{
						  "Link": "//jisho.org/search/%E9%9D%A2%20%E3%81%8A%E3%82%82",
						  "Word": "おも"
						},
						{
						  "Link": "//jisho.org/search/%E9%9D%A2%20%E3%81%8A%E3%82%82%E3%81%A6",
						  "Word": "おもて"
						},
						{
						  "Link": "//jisho.org/search/%E9%9D%A2%20%E3%81%A4%E3%82%89",
						  "Word": "つら"
						}
					  ],
					  "Onyomis": [
						{
						  "Link": "//jisho.org/search/%E9%9D%A2%20%E3%82%81%E3%82%93",
						  "Word": "メン"
						},
						{
						  "Link": "//jisho.org/search/%E9%9D%A2%20%E3%81%B9%E3%82%93",
						  "Word": "ベン"
						}
					  ]
					},
					{
					  "Kanji": {
						"Link": "//jisho.org/search/%E9%9B%BB%20%23kanji",
						"Word": "電"
					  },
					  "Meaning": "\n            electricity\n      ",
					  "Kunyomis": null,
					  "Onyomis": [
						{
						  "Link": "//jisho.org/search/%E9%9B%BB%20%E3%81%A7%E3%82%93",
						  "Word": "デン"
						}
					  ]
					},
					{
					  "Kanji": {
						"Link": "//jisho.org/search/%E8%BB%8A%20%23kanji",
						"Word": "車"
					  },
					  "Meaning": "\n            car\n      ",
					  "Kunyomis": [
						{
						  "Link": "//jisho.org/search/%E8%BB%8A%20%E3%81%8F%E3%82%8B%E3%81%BE",
						  "Word": "くるま"
						}
					  ],
					  "Onyomis": [
						{
						  "Link": "//jisho.org/search/%E8%BB%8A%20%E3%81%97%E3%82%83",
						  "Word": "シャ"
						}
					  ]
					},
					{
					  "Kanji": {
						"Link": "//jisho.org/search/%E5%A0%B4%20%23kanji",
						"Word": "場"
					  },
					  "Meaning": "\n            location, \n            place\n      ",
					  "Kunyomis": [
						{
						  "Link": "//jisho.org/search/%E5%A0%B4%20%E3%81%B0",
						  "Word": "ば"
						}
					  ],
					  "Onyomis": [
						{
						  "Link": "//jisho.org/search/%E5%A0%B4%20%E3%81%98%E3%82%87%E3%81%86",
						  "Word": "ジョウ"
						},
						{
						  "Link": "//jisho.org/search/%E5%A0%B4%20%E3%81%A1%E3%82%87%E3%81%86",
						  "Word": "チョウ"
						}
					  ]
					}
				  ]
				},
				"Kanjidmg": [
				  {
					"WordSection": {
					  "Kanji": "路",
					  "KanjiImage": null,
					  "Meaning": "road",
					  "Link": "http://www.kanjidamage.com路"
					},
					"Radicals": [
					  {
						"Kanji": "足",
						"KanjiImage": null,
						"Meaning": "foot/ be enough",
						"Link": "http://www.kanjidamage.com//kanji/277-foot-be-enough-%E8%B6%B3"
					  },
					  {
						"Kanji": "各",
						"KanjiImage": null,
						"Meaning": "each",
						"Link": "http://www.kanjidamage.com//kanji/609-each-%E5%90%84"
					  }
					],
					"Onyomi": "RO\n\n\nshould be easy to remember, because it sounds like ROad",
					"Mnemonic": "Each foot walks on the same road"
				  },
				  {
					"WordSection": {
					  "Kanji": "面",
					  "KanjiImage": null,
					  "Meaning": "front surface / face",
					  "Link": "http://www.kanjidamage.com面"
					},
					"Radicals": [
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "terrorist",
						"Link": "http://www.kanjidamage.com//kanji/812-terrorist"
					  },
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "big box",
						"Link": "http://www.kanjidamage.com//kanji/431-big-box"
					  }
					],
					"Onyomi": "MEN",
					"Mnemonic": "there seems to be a LADDER in the center of this kanji! So we can say . . \n\n\nThe terrorist MEN climbed up the ladder onto the surface of the big box (a Walmart) and said they'd blow it up unless the Government did the Humpty"
				  },
				  {
					"WordSection": {
					  "Kanji": "電",
					  "KanjiImage": null,
					  "Meaning": "electricity",
					  "Link": "http://www.kanjidamage.com電"
					},
					"Radicals": [
					  {
						"Kanji": "雨",
						"KanjiImage": null,
						"Meaning": "rain",
						"Link": "http://www.kanjidamage.com//kanji/1383-rain-%E9%9B%A8"
					  },
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "dragon radical",
						"Link": "http://www.kanjidamage.com//kanji/1400-dragon-radical"
					  }
					],
					"Onyomi": "DEN\n\n\nI flip the switch and DEN ('then') the electric light goes on",
					"Mnemonic": "Electricity comes from lightning in the rain and also from the lightning breath of dragons"
				  },
				  {
					"WordSection": {
					  "Kanji": "車",
					  "KanjiImage": null,
					  "Meaning": "car",
					  "Link": "http://www.kanjidamage.com車"
					},
					"Radicals": null,
					"Onyomi": "SHA\n\n\nTo get from A to B, you SHALL need a car",
					"Mnemonic": null
				  },
				  {
					"WordSection": {
					  "Kanji": "停",
					  "KanjiImage": null,
					  "Meaning": "bring to a halt",
					  "Link": "http://www.kanjidamage.com停"
					},
					"Radicals": [
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "personleft",
						"Link": "http://www.kanjidamage.com//kanji/61-person-%E4%BA%BA"
					  },
					  {
						"Kanji": "亭",
						"KanjiImage": null,
						"Meaning": "restaurant",
						"Link": "http://www.kanjidamage.com//kanji/81-restaurant-%E4%BA%AD"
					  }
					],
					"Onyomi": "TEI",
					"Mnemonic": "The police person brings a halt to the restaurant because too many people got a damn nail in their sandwitch and what the hell kind of deal is that, anyway? Imagine the pain"
				  },
				  {
					"WordSection": {
					  "Kanji": "留",
					  "KanjiImage": null,
					  "Meaning": "absent / stopped",
					  "Link": "http://www.kanjidamage.com留"
					},
					"Radicals": [
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "decapitated cow",
						"Link": "http://www.kanjidamage.com//kanji/1244-decapitated-cow"
					  },
					  {
						"Kanji": "刀",
						"KanjiImage": null,
						"Meaning": "sword",
						"Link": "http://www.kanjidamage.com//kanji/164-sword-%E5%88%80"
					  },
					  {
						"Kanji": "田",
						"KanjiImage": null,
						"Meaning": "rice field",
						"Link": "http://www.kanjidamage.com//kanji/56-rice-field-%E7%94%B0"
					  }
					],
					"Onyomi": "RYUU\n\n\nYou'll have to REUSE that phone to call me later - I'm away from home right now",
					"Mnemonic": "The cow's head is away - it was chopped off with a sword and then hidden in a rice field. Kids these days"
				  },
				  {
					"WordSection": {
					  "Kanji": "場",
					  "KanjiImage": null,
					  "Meaning": "place",
					  "Link": "http://www.kanjidamage.com場"
					},
					"Radicals": [
					  {
						"Kanji": "土",
						"KanjiImage": null,
						"Meaning": "earth",
						"Link": "http://www.kanjidamage.com//kanji/235-earth-%E5%9C%9F"
					  },
					  {
						"Kanji": "易",
						"KanjiImage": null,
						"Meaning": "easy",
						"Link": "http://www.kanjidamage.com//kanji/1203-easy-%E6%98%93"
					  }
					],
					"Onyomi": "JOU\n\n\nforget it",
					"Mnemonic": "It's easy for JOE Stalin to take over all places on the earth with his legions of fanatics"
				  }
				],
				"Error": null
			  }`,
		},
		{
			word: "あったり前",
			expectJSON: `{
				"EnglishSearchedWord": "",
				"JishoEnglishWordLink": "",
				"Jisho": {
				  "Link": "https://jisho.org/search/あったり前",
				  "WordSection": {
					"FullWord": "当ったり前",
					"Parts": [
					  {
						"MainText": "当",
						"Reading": "あ"
					  },
					  {
						"MainText": "ったり",
						"Reading": ""
					  },
					  {
						"MainText": "前",
						"Reading": "まえ"
					  }
					],
					"Meanings": [
					  {
						"ListIdx": 1,
						"Meaning": "natural; reasonable; obvious",
						"Tags": "Na-adjective (keiyodoshi), Noun which may take the genitive case particle 'no', Noun"
					  },
					  {
						"ListIdx": 2,
						"Meaning": "usual; common; ordinary",
						"Tags": "Na-adjective (keiyodoshi), Noun, Noun which may take the genitive case particle 'no'"
					  }
					]
				  },
				  "Kanjis": [
					{
					  "Kanji": {
						"Link": "//jisho.org/search/%E5%89%8D%20%23kanji",
						"Word": "前"
					  },
					  "Meaning": "\n            in front, \n            before\n      ",
					  "Kunyomis": [
						{
						  "Link": "//jisho.org/search/%E5%89%8D%20%E3%81%BE%E3%81%88",
						  "Word": "まえ"
						},
						{
						  "Link": "//jisho.org/search/%E5%89%8D%20%E3%81%BE%E3%81%88",
						  "Word": "-まえ"
						}
					  ],
					  "Onyomis": [
						{
						  "Link": "//jisho.org/search/%E5%89%8D%20%E3%81%9C%E3%82%93",
						  "Word": "ゼン"
						}
					  ]
					}
				  ]
				},
				"Kanjidmg": [
				  {
					"WordSection": {
					  "Kanji": "前",
					  "KanjiImage": null,
					  "Meaning": "before",
					  "Link": "http://www.kanjidamage.com前"
					},
					"Radicals": [
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "standing worms",
						"Link": "http://www.kanjidamage.com//kanji/147-standing-worms"
					  },
					  {
						"Kanji": "月",
						"KanjiImage": null,
						"Meaning": "moon/organ",
						"Link": "http://www.kanjidamage.com//kanji/41-moon-organ-%E6%9C%88"
					  },
					  {
						"Kanji": "刀",
						"KanjiImage": null,
						"Meaning": "sword",
						"Link": "http://www.kanjidamage.com//kanji/164-sword-%E5%88%80"
					  }
					],
					"Onyomi": "ZEN\n\n\nZEN is before NOW. That is a terrible pun but now you won't be able to forget it",
					"Mnemonic": "Cut the standing worms with your sword before the full moon. Otherwise they'll turn into WERE-worms, of which the less said, the better"
				  }
				],
				"Error": null
			  }`,
		},
		{
			word: "相変わらず",
			expectJSON: `{
				"EnglishSearchedWord": "",
				"JishoEnglishWordLink": "",
				"Jisho": {
				  "Link": "https://jisho.org/search/相変わらず",
				  "WordSection": {
					"FullWord": "相変わらず",
					"Parts": [
					  {
						"MainText": "相",
						"Reading": "あい"
					  },
					  {
						"MainText": "変",
						"Reading": "か"
					  },
					  {
						"MainText": "わらず",
						"Reading": ""
					  }
					],
					"Meanings": [
					  {
						"ListIdx": 1,
						"Meaning": "as usual; as always; as before; as ever; still",
						"Tags": "Adverb (fukushi), Noun which may take the genitive case particle 'no'"
					  },
					  {
						"ListIdx": 2,
						"Meaning": "相変らず 【あいかわらず】、あい変わらず 【あいかわらず】、あい変らず 【あいかわらず】",
						"Tags": "Other forms"
					  }
					]
				  },
				  "Kanjis": [
					{
					  "Kanji": {
						"Link": "//jisho.org/search/%E7%9B%B8%20%23kanji",
						"Word": "相"
					  },
					  "Meaning": "\n            inter-, \n            mutual, \n            together, \n            each other, \n            minister of state, \n            councillor, \n            aspect, \n            phase, \n            physiognomy\n      ",
					  "Kunyomis": [
						{
						  "Link": "//jisho.org/search/%E7%9B%B8%20%E3%81%82%E3%81%84",
						  "Word": "あい-"
						}
					  ],
					  "Onyomis": [
						{
						  "Link": "//jisho.org/search/%E7%9B%B8%20%E3%81%9D%E3%81%86",
						  "Word": "ソウ"
						},
						{
						  "Link": "//jisho.org/search/%E7%9B%B8%20%E3%81%97%E3%82%87%E3%81%86",
						  "Word": "ショウ"
						}
					  ]
					},
					{
					  "Kanji": {
						"Link": "//jisho.org/search/%E5%A4%89%20%23kanji",
						"Word": "変"
					  },
					  "Meaning": "\n            unusual, \n            change, \n            strange\n      ",
					  "Kunyomis": [
						{
						  "Link": "//jisho.org/search/%E5%A4%89%20%E3%81%8B%E3%82%8F%E3%82%8B",
						  "Word": "か.わる"
						},
						{
						  "Link": "//jisho.org/search/%E5%A4%89%20%E3%81%8B%E3%82%8F%E3%82%8A",
						  "Word": "か.わり"
						},
						{
						  "Link": "//jisho.org/search/%E5%A4%89%20%E3%81%8B%E3%81%88%E3%82%8B",
						  "Word": "か.える"
						}
					  ],
					  "Onyomis": [
						{
						  "Link": "//jisho.org/search/%E5%A4%89%20%E3%81%B8%E3%82%93",
						  "Word": "ヘン"
						}
					  ]
					}
				  ]
				},
				"Kanjidmg": [
				  {
					"WordSection": {
					  "Kanji": "相",
					  "KanjiImage": null,
					  "Meaning": "partner",
					  "Link": "http://www.kanjidamage.com相"
					},
					"Radicals": [
					  {
						"Kanji": "木",
						"KanjiImage": null,
						"Meaning": "tree",
						"Link": "http://www.kanjidamage.com//kanji/335-tree-%E6%9C%A8"
					  },
					  {
						"Kanji": "目",
						"KanjiImage": null,
						"Meaning": "eye",
						"Link": "http://www.kanjidamage.com//kanji/76-eye-%E7%9B%AE"
					  }
					],
					"Onyomi": "SOU\n\n\nHe's got SO many partners",
					"Mnemonic": "If your partner has a tree splinter in their eye, you have to take it out even if it is really gross. And vice versa"
				  },
				  {
					"WordSection": {
					  "Kanji": "変",
					  "KanjiImage": null,
					  "Meaning": "change",
					  "Link": "http://www.kanjidamage.com変"
					},
					"Radicals": [
					  {
						"Kanji": "赤",
						"KanjiImage": null,
						"Meaning": "red",
						"Link": "http://www.kanjidamage.com//kanji/900-red-%E8%B5%A4"
					  },
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "eachbottom",
						"Link": "http://www.kanjidamage.com//kanji/609-each-%E5%90%84"
					  }
					],
					"Onyomi": "HEN\n\n\nThe egg changes into a HEN after hatching",
					"Mnemonic": "Each red Communist changes into a yuppie when they turn 30"
				  }
				],
				"Error": null
			  }`,
		},
		{
			word: "ペラペラ",
			expectJSON: `{
				"EnglishSearchedWord": "",
				"JishoEnglishWordLink": "",
				"Jisho": {
				  "Link": "https://jisho.org/search/ペラペラ",
				  "WordSection": {
					"FullWord": "ペラペラ",
					"Parts": [
					  {
						"MainText": "ペラペラ",
						"Reading": ""
					  }
					],
					"Meanings": [
					  {
						"ListIdx": 1,
						"Meaning": "fluently (speaking a foreign language)",
						"Tags": "Adverb (fukushi), Adverb taking the 'to' particle, Na-adjective (keiyodoshi)"
					  },
					  {
						"ListIdx": 2,
						"Meaning": "incessantly (speaking); glibly; garrulously; volubly",
						"Tags": "Adverb (fukushi), Adverb taking the 'to' particle"
					  },
					  {
						"ListIdx": 3,
						"Meaning": "one after the other (flipping through pages)",
						"Tags": "Adverb (fukushi), Adverb taking the 'to' particle"
					  },
					  {
						"ListIdx": 4,
						"Meaning": "thin (paper, cloth, etc.); flimsy; weak",
						"Tags": "Noun which may take the genitive case particle 'no', Na-adjective (keiyodoshi), Adverb (fukushi), Suru verb"
					  },
					  {
						"ListIdx": 5,
						"Meaning": "ぺらぺら",
						"Tags": "Other forms"
					  }
					]
				  },
				  "Kanjis": null
				},
				"Kanjidmg": null,
				"Error": null
			  }`,
		},
		{
			word: "driver's licence",
			expectJSON: `{
				"EnglishSearchedWord": "driver's licence",
				"JishoEnglishWordLink": "https://jisho.org/search/driver's licence",
				"Jisho": {
				  "Link": "https://jisho.org/search/運転免許",
				  "WordSection": {
					"FullWord": "運転免許",
					"Parts": [
					  {
						"MainText": "運",
						"Reading": "うん"
					  },
					  {
						"MainText": "転",
						"Reading": "てん"
					  },
					  {
						"MainText": "免",
						"Reading": "めん"
					  },
					  {
						"MainText": "許",
						"Reading": "きょ"
					  }
					],
					"Meanings": [
					  {
						"ListIdx": 1,
						"Meaning": "driver's license; driver's licence; driving licence",
						"Tags": "Noun"
					  },
					  {
						"ListIdx": 2,
						"Meaning": "Driver's license",
						"Tags": "Wikipedia definition"
					  }
					]
				  },
				  "Kanjis": null
				},
				"Kanjidmg": [
				  {
					"WordSection": {
					  "Kanji": "運",
					  "KanjiImage": null,
					  "Meaning": "carry / luck",
					  "Link": "http://www.kanjidamage.com運"
					},
					"Radicals": [
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "motion",
						"Link": "http://www.kanjidamage.com//kanji/327-motion"
					  },
					  {
						"Kanji": "軍",
						"KanjiImage": null,
						"Meaning": "army",
						"Link": "http://www.kanjidamage.com//kanji/1068-army-%E8%BB%8D"
					  }
					],
					"Onyomi": "UN",
					"Mnemonic": "I wish the army luck when they move forward with their campaign - hopefully they only meet UN-armed civilians"
				  },
				  {
					"WordSection": {
					  "Kanji": "転",
					  "KanjiImage": null,
					  "Meaning": "roll over",
					  "Link": "http://www.kanjidamage.com転"
					},
					"Radicals": [
					  {
						"Kanji": "車",
						"KanjiImage": null,
						"Meaning": "car",
						"Link": "http://www.kanjidamage.com//kanji/1058-car-%E8%BB%8A"
					  },
					  {
						"Kanji": "云",
						"KanjiImage": null,
						"Meaning": "twin decapited cows",
						"Link": "http://www.kanjidamage.com//kanji/1265-twin-decapited-cows-%E4%BA%91"
					  }
					],
					"Onyomi": "TEN",
					"Mnemonic": "The car rolled over the twin cows TEN times, decapitating them"
				  },
				  {
					"WordSection": {
					  "Kanji": "免",
					  "KanjiImage": null,
					  "Meaning": "exemption / license",
					  "Link": "http://www.kanjidamage.com免"
					},
					"Radicals": [
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "bait",
						"Link": "http://www.kanjidamage.com//kanji/1209-bait"
					  },
					  {
						"Kanji": null,
						"KanjiImage": "aW1hZ2VieXRlcw==",
						"Meaning": "human legs",
						"Link": "http://www.kanjidamage.com//kanji/519-human-legs"
					  }
					],
					"Onyomi": "MEN\n\n\nMEN have a license to urinate standing up",
					"Mnemonic": "You need a fishing license to walk to the river on your human legs and throw your baited worm in"
				  },
				  {
					"WordSection": {
					  "Kanji": "許",
					  "KanjiImage": null,
					  "Meaning": "allow",
					  "Link": "http://www.kanjidamage.com許"
					},
					"Radicals": [
					  {
						"Kanji": "言",
						"KanjiImage": null,
						"Meaning": "say",
						"Link": "http://www.kanjidamage.com//kanji/11-say-%E8%A8%80"
					  },
					  {
						"Kanji": "午",
						"KanjiImage": null,
						"Meaning": "noon",
						"Link": "http://www.kanjidamage.com//kanji/1192-noon-%E5%8D%88"
					  }
					],
					"Onyomi": "KYO",
					"Mnemonic": "Keep Your Opinions to yourself until I allow you to say them at noon"
				  }
				],
				"Error": null
			  }`,
		},
		{
			word: "やはり",
			expectJSON: `{
				"EnglishSearchedWord": "",
				"JishoEnglishWordLink": "",
				"Jisho": {
				  "Link": "https://jisho.org/search/やはり",
				  "WordSection": {
					"FullWord": "矢張り",
					"Parts": null,
					"Meanings": [
					  {
						"ListIdx": 1,
						"Meaning": "as expected; sure enough; just as one thought",
						"Tags": "Adverb (fukushi)"
					  },
					  {
						"ListIdx": 2,
						"Meaning": "after all (is said and done); in the end; as one would expect; in any case",
						"Tags": "Adverb (fukushi)"
					  },
					  {
						"ListIdx": 3,
						"Meaning": "too; also; as well; likewise; (not) either",
						"Tags": "Adverb (fukushi)"
					  },
					  {
						"ListIdx": 4,
						"Meaning": "still; as before",
						"Tags": "Adverb (fukushi)"
					  },
					  {
						"ListIdx": 5,
						"Meaning": "all the same; even so; still; nonetheless",
						"Tags": "Adverb (fukushi)"
					  },
					  {
						"ListIdx": 6,
						"Meaning": "矢張 【やはり】",
						"Tags": "Other forms"
					  },
					  {
						"ListIdx": 7,
						"Meaning": "",
						"Tags": "Notes"
					  }
					]
				  },
				  "Kanjis": [ "notemptyarrayIExpect" ]
				},
				"Kanjidmg": null,
				"Error": null
			  }`,
		},
		{
			word: "矢張り",
			expectJSON: `{
				"EnglishSearchedWord": "",
				"JishoEnglishWordLink": "",
				"Jisho": {
				  "Link": "https://jisho.org/search/やはり",
				  "WordSection": {
					"FullWord": "矢張り",
					"Parts": null,
					"Meanings": [
					  {
						"ListIdx": 1,
						"Meaning": "as expected; sure enough; just as one thought",
						"Tags": "Adverb (fukushi)"
					  },
					  {
						"ListIdx": 2,
						"Meaning": "after all (is said and done); in the end; as one would expect; in any case",
						"Tags": "Adverb (fukushi)"
					  },
					  {
						"ListIdx": 3,
						"Meaning": "too; also; as well; likewise; (not) either",
						"Tags": "Adverb (fukushi)"
					  },
					  {
						"ListIdx": 4,
						"Meaning": "still; as before",
						"Tags": "Adverb (fukushi)"
					  },
					  {
						"ListIdx": 5,
						"Meaning": "all the same; even so; still; nonetheless",
						"Tags": "Adverb (fukushi)"
					  },
					  {
						"ListIdx": 6,
						"Meaning": "矢張 【やはり】",
						"Tags": "Other forms"
					  },
					  {
						"ListIdx": 7,
						"Meaning": "",
						"Tags": "Notes"
					  }
					]
				  },
				  "Kanjis": [ "notemptyarrayIExpect" ]
				},
				"Kanjidmg": null,
				"Error": null
			  }`,
		},
	}

	// "special case" (fill by hand) for creating kanjidmg lookup urls (when looked up words do not contain the kanjis, but jisho returns kanjis)
	kanjidmgLinkWords := []string{
		"運転免許",
		"矢張",
	}
	for _, tc := range testCases {
		kanjidmgLinkWords = append(kanjidmgLinkWords, tc.word)
	}
	kanjidmgLinks := make(map[string]string)
	for _, word := range kanjidmgLinkWords {
		for _, r := range word {
			if !jptext.IsKanji(r) {
				continue
			}

			rStr := string(r)
			kanjidmgLinks[rStr] = omnikanji.KanjidmgBaseUrl + rStr
		}
	}

	getData := func(t *testing.T, tc *TestCase) *server.TemplateParams {
		fixtureDir, err := filepath.Abs("fixture")
		require.NoError(t, err)

		httpClient := NewHttpClientMock(fixtureDir)
		jisho := dictproxy.NewJisho(omnikanji.JishoSearchUrl, httpClient)
		kanjidmg := dictproxy.NewKanjidmg(kanjidmgLinks, httpClient)
		srv := server.NewServer(nil, jisho, kanjidmg)

		queryWord := url.QueryEscape(tc.word)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:8080/search/?word=%s", queryWord), nil)

		data := srv.HandleIndex(nil, req)
		return data
	}

	// t.Run("generate JSONs", func(t *testing.T) {
	// 	for _, tc := range testCases {
	// 		t.Run(tc.word, func(t *testing.T) {
	// 			data := getData(t, tc)

	// 			b, _ := json.MarshalIndent(data, "", "  ")
	// 			log.Println(string(b))
	// 			t.Fail()
	// 		})
	// 	}
	// })

	for _, tc := range testCases {
		t.Run(tc.word, func(t *testing.T) {
			data := getData(t, tc)

			require.False(t, tc.expect == nil && tc.expectJSON == "", "No expects? You sure?")

			if tc.expectJSON != "" {
				dataB, err := json.MarshalIndent(data, "", "  ")
				require.NoError(t, err)
				require.JSONEq(t, tc.expectJSON, string(dataB))
			}

			if tc.expect != nil {
				tc.expect(t, data)
			}
		})
	}
}
