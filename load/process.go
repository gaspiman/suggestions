package load

import (
	"fmt"
	"log"
	"suggestions/suggestion"
	"sync"

	"github.com/dgraph-io/badger/v2"
)

func (l *Loader) PushSeed(inCH chan *suggestion.Response) {
	i := 0
	l.Lang.Seed = map[string]struct{}{
		"michael jackson": struct{}{},
	}
	for k := range l.Lang.Seed {
		if i%100 == 0 {
			log.Printf("processed: %v", i)
		}
		i++
		inCH <- suggestion.NewResponse(k)
	}
}

func (l *Loader) ProcessKeywords(inCH chan *suggestion.Response, outCH chan *suggestion.Response, wg *sync.WaitGroup) {
	batch := l.Datastore.DB.NewWriteBatch()
	batch.SetMaxPendingTxns(16)
	defer wg.Done()
	defer batch.Cancel()
	for sugg := range outCH {
		gobbed, err := sugg.Encode()
		fmt.Println(">>> HERE >>>>>", sugg)
		if err != nil {
			log.Fatalf("couldn't gob: %v", err)
		}
		if err := batch.Set([]byte(sugg.Query), gobbed); err != nil {
			log.Fatalf("couldn't SetEntry: %v", err)
		}
	}
	if err := batch.Flush(); err != nil {
		log.Fatalf("couldn't flush: %v", err)
	}
}

func (l *Loader) IterateDB(inCH chan *suggestion.Response) bool {
	fmt.Println("Iterating DB")
	hasPushed := false
	if err := l.Datastore.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				resp, err := suggestion.DecodeResponse(v)
				if err != nil {
					return err
				}
				if l.PushBackSuggestions(resp, inCH) {
					hasPushed = true
				}
				fmt.Printf("key=%s, value=%s\n", k, resp)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Panicf("couldn't iterate: %v", err)
	}
	return hasPushed
}

func (l *Loader) PushBackSuggestions(resp *suggestion.Response, inCH chan *suggestion.Response) bool {
	hasPushed := false
	chunks := l.Lang.Chunker(resp.Query)
	fmt.Println("CHUNKS RECEIVED", chunks)
	for _, chunk := range chunks {
		log.Printf("pushing query: %s\n", chunk)
		hasPushed = true
		inCH <- &suggestion.Response{
			Query: chunk,
		}
	}
	for _, sugg := range resp.Suggestions {
		chunks := l.Lang.Chunker(sugg.Text)
		for _, chunk := range chunks {
			log.Printf("pushing query 2: %s\n", chunk)
			hasPushed = true
			inCH <- &suggestion.Response{
				Query: chunk,
			}
		}
	}
	return hasPushed
}
