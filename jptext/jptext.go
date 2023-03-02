package jptext

const (
	kanjiStartRange = 0x4e00
	kanjiEndRange   = 0x9faf

	jpStartRange = 0x3000
	jpEndRange   = 0x9faf
)

func IsJapanese(c rune) bool {
	return c >= jpStartRange && c <= jpEndRange
}

func IsKanji(c rune) bool {
	return c >= kanjiStartRange && c <= kanjiEndRange
}

func IsJapaneseWord(word string) bool {
	for _, r := range word {
		if !IsJapanese(r) {
			return false
		}
	}
	return true
}

func IsKanjiWord(word string) bool {
	for _, r := range word {
		if !IsKanji(r) {
			return false
		}
	}
	return true
}

func ExtractKanjis(word string) string {
	wordKanjis := ""
	for _, c := range word {
		if IsKanji(c) {
			wordKanjis += string(c)
		}
	}
	return wordKanjis
}

func KanjisCountInAWord(word string) int {
	kanjiCount := 0
	for _, c := range word {
		if IsKanji(c) {
			kanjiCount++
		}
	}
	return kanjiCount
}
