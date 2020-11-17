package load

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"suggestions/datastore"
	"suggestions/lang"
	"suggestions/suggestion"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/dgraph-io/badger/v2"
)

var suggEndpoint = "http://suggestqueries.google.com/complete/search?output=toolbar&hl=en-US&q="

type Toplevel struct {
	CompleteSuggestion []struct {
		Suggestion struct {
			Data string `xml:"data,attr"`
		} `xml:"suggestion"`
	} `xml:"CompleteSuggestion"`
}

type Loader struct {
	Lang      *lang.Lang
	Datastore *datastore.Datastore
}

func NewLoader(sw, freq, titles string) *Loader {
	language, err := lang.NewLang(sw, freq, titles)
	if err != nil {
		log.Fatalf("couldn't load language: %f", err)
	}
	datastore, err := datastore.NewDatastore()
	if err != nil {
		log.Fatalf("couldn't create datastore: %v", err)
	}
	loader := &Loader{
		Lang:      language,
		Datastore: datastore,
	}
	return loader
}

func Load(sw, freq, titles string) error {
	loader := NewLoader(sw, freq, titles)

	// load the dict from the db
	loader.dictFromDB()
	//log.Fatal("END")
	wg := new(sync.WaitGroup)
	wg2 := new(sync.WaitGroup)
	inCH := make(chan *suggestion.Response)
	outCH := make(chan *suggestion.Response)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		wg2.Add(1)
		go loader.remote(inCH, outCH, wg)
		go loader.ProcessKeywords(inCH, outCH, wg2)
	}

	loader.PushSeed(inCH)

	for loader.IterateDB(inCH) {

	}

	close(inCH)
	wg.Wait()
	close(outCH)
	wg2.Wait()

	return nil
}

func (loader *Loader) dictFromDB() {
	if err := loader.Datastore.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()
		count := 0
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			loader.Lang.Dict.Store(string(k), nil)
			count++
			if count%100 == 0 {
				log.Printf("loaded keys to dict: %d\n", count)
			}
		}
		return nil
	}); err != nil {
		log.Panicf("couldn't iterate: %v", err)
	}
}

func (loader *Loader) remote(inCH chan *suggestion.Response, outCH chan *suggestion.Response, wg *sync.WaitGroup) {
	defer wg.Done()
	for sugg := range inCH {
		data, err := request(suggEndpoint + url.QueryEscape(sugg.Query))
		if err != nil {
			log.Fatalf("Request error: %v", err)
		}
		if !bytes.HasPrefix(data, []byte("<?xml version=\"1.0\"?>")) {
			log.Printf("Non xml response with query: %s", sugg.Query)
			log.Printf("url: %s\n%s", suggEndpoint+url.QueryEscape(sugg.Query), data)
			continue
		}
		suggestions, err := requestXML(data)
		if err != nil {
			log.Fatalf("error while unmarshal: %v", err)
			continue
		}
		if len(suggestions) > 0 {
			sugg.Suggestions = suggestions
		} else {
			sugg.Suggestions = []*suggestion.Suggestion{
				&suggestion.Suggestion{
					Text: sugg.Query,
				},
			}
		}
		outCH <- sugg
	}
}

// ValidUTF8Reader implements a Reader which reads only bytes that constitute valid UTF-8
type ValidUTF8Reader struct {
	buffer *bufio.Reader
}

// Function Read reads bytes in the byte array b. n is the number of bytes read.
func (rd ValidUTF8Reader) Read(b []byte) (n int, err error) {
	for {
		var r rune
		var size int
		r, size, err = rd.buffer.ReadRune()
		if err != nil {
			return
		}
		if r == unicode.ReplacementChar && size == 1 {
			continue
		} else if n+size < len(b) {
			utf8.EncodeRune(b[n:], r)
			n += size
		} else {
			rd.buffer.UnreadRune()
			break
		}
	}
	return
}

// NewValidUTF8Reader constructs a new ValidUTF8Reader that wraps an existing io.Reader
func NewValidUTF8Reader(rd io.Reader) ValidUTF8Reader {
	return ValidUTF8Reader{bufio.NewReader(rd)}
}

func request(u string) ([]byte, error) {
	fmt.Println("Request URL:", u)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "*/*")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	vutf := NewValidUTF8Reader(resp.Body)

	defer resp.Body.Close()
	return ioutil.ReadAll(vutf)
}

func requestXML(data []byte) ([]*suggestion.Suggestion, error) {
	rs := Toplevel{}
	if err := xml.Unmarshal(data, &rs); err != nil {
		return nil, err
	}
	suggestions := make([]*suggestion.Suggestion, 0, len(rs.CompleteSuggestion))
	for _, sugg := range rs.CompleteSuggestion {
		suggestions = append(suggestions, suggestion.NewSuggestion(sugg.Suggestion.Data))
	}
	return suggestions, nil
}
