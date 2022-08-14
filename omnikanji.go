package omnikanji


const (
	JishoSearchUrl = "https://jisho.org/search/"

	KanjidmgBaseUrl = "http://www.kanjidamage.com"
	KanjidmgListUrl = KanjidmgBaseUrl + "/kanji"
)

type JishoSection struct {
	Link        string
	WordSection JishoWordSection
	Kanjis      []JishoKanji
}

type JishoWordSection struct {
	FullWord string
	Parts    []JishoWordPart
	Meanings []JishoMeaning
	//Notes *string
}

type JishoWordPart struct {
	MainText string
	Reading  string // Reading can be empty in case it's not a kanji
}

type JishoMeaning struct {
	TemplateListItem
	Meaning string
	Tags    *string
	//MeaningSentence  *string
	//SupplementalInfo *string
}

type TemplateListItem struct {
	ListIdx int
}

type JishoKanji struct {
	Kanji    JishoWordWithLink
	Meaning  string
	Kunyomis []JishoWordWithLink
	Onyomis  []JishoWordWithLink
}

type JishoWordWithLink struct {
	Link string
	Word string
}


type KanjidmgSection struct {
	WordSection KanjidmgKanji
	TopComment  *string
	Radicals    []KanjidmgKanji
	Onyomi      *string
	Mnemonic    *string
	Mutants     []KanjidmgKanji

	// Kunyomi TODO
	//Jukugo  TODO
	// UsedIn TODO
	// Synonyms TODO
	// Lookalikes TODO
}

type KanjidmgKanji struct {
	Kanji      *string
	KanjiImage *string
	Meaning    string
	Link       string
}