package lang

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Lang struct {
	Stopwords map[string]struct{}
	Seed      map[string]struct{}
	Dict      sync.Map
}

var stringsReplacer = strings.NewReplacer("!", "", "\"", "", "#", "", "$", "", "%", "", "&", "", "'", "", "(", "", ")", "", "*", "", "+", "", ",", "", "-", "", ".", "", "/", "", ":", "", ";", "", "<", "", "=", "", ">", "", "?", "", "@", "", "[", "", "]", "", "^", "", "`", "", "{", "", "|", "", "}", "", "~", "", "_", " ")

func NewLang(sw, freq, titles string) (*Lang, error) {
	l := &Lang{}

	l.Stopwords = map[string]struct{}{}
	l.Seed = map[string]struct{}{}
	if err := l.LoadStopwords(sw); err != nil {
		return nil, err
	}
	if err := l.LoadFrequency(freq); err != nil {
		return nil, err
	}
	if err := l.LoadWikiTitles(titles); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Lang) Load(k string, v []string) {
	l.Dict.Store(k, v)
}

func (l *Lang) IsStopword(s string) bool {
	if _, ok := l.Stopwords[s]; ok {
		return true
	}
	return false
}

func (l *Lang) LoadStopwords(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		stopword := scanner.Text()
		stopword = l.ProcessText(stopword)
		l.Stopwords[stopword] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (l *Lang) LoadFrequency(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		if len(parts) != 2 {
			continue
		}
		if count, err := strconv.Atoi(parts[1]); err == nil {
			if count < 5 {
				continue
			}
		} else {
			log.Printf("error while parsing count: %v", err)
			continue
		}
		text := l.ProcessText(parts[0])
		if l.IsStopword(text) {
			continue
		}
		l.Seed[text] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (l *Lang) LoadWikiTitles(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	header := false

	for scanner.Scan() {
		if header == false {
			header = true
			continue
		}
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) != 2 {
			continue
		}
		// Keep only the Wikipedia pages
		if parts[0] != "0" {
			break
		}
		text := parts[1]
		parts = strings.Split(text, "_(")
		text = l.ProcessText(parts[0])
		if !l.IsValidWord(text) || l.IsStopword(text) {
			continue
		}
		l.Seed[text] = struct{}{}
		if len(l.Seed)%1000000 == 0 {
			log.Printf("Processed wiki titles: %d", len(l.Seed))
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (l *Lang) ProcessText(s string) string {
	s = strings.ToLower(s)
	s = stringsReplacer.Replace(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func (l *Lang) IsValidWord(s string) bool {
	if s == "" {
		return false
	}
	// The word is a digit
	if _, err := strconv.Atoi(s); err == nil {
		return false
	}
	return true
}

func (l *Lang) LoadOrStore(s string) bool {
	_, loaded := l.Dict.LoadOrStore(s, nil)
	return loaded
}
