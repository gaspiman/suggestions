package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"suggestions/suggestion"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/gorilla/mux"
)

var expireKey = 1 * time.Hour

func queryHandle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query, ok := vars["query"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "missing query string")
		return
	}
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "query string is empty")
		return
	}
	/*
		resp := GetList(query)
		if resp != nil {
			fmt.Println("GOT RESUTLREDIS", resp)
			js, err := json.Marshal(resp)
			if err != nil {
				fmt.Println("error")
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, string(js))
			fmt.Println("resp: %s", resp)
			return
		}
	*/

	fmt.Println("Iterating DB")
	if err := server.Datastore.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		it.Seek([]byte(query))
		item := it.Item()
		k := item.Key()
		log.Printf("seek key: %s", k)
		err := item.Value(func(v []byte) error {
			resp, err := suggestion.DecodeResponse(v)
			if err != nil {
				return err
			}
			js, err := json.Marshal(resp)
			if err != nil {
				return err
			}
			//SetList(resp)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, string(js))
			fmt.Println("resp: %s", resp)
			return nil
		})
		return err
	}); err != nil {
		log.Panicf("couldn't iterate: %v", err)
	}
}

func SetList(resp *suggestion.Response) {
	log.Printf("Setting cache: %v\n", resp)
	var ctx = context.Background()
	s := make([]interface{}, 0, len(resp.Suggestions))
	for _, sugg := range resp.Suggestions {
		s = append(s, sugg.Text)
	}

	server.Cachestore.Client.LPush(ctx, resp.Query, s...)
	server.Cachestore.Client.Expire(ctx, resp.Query, expireKey)
}

func GetList(key string) *suggestion.Response {
	log.Printf("Getting from cache: %s\n", key)
	var ctx = context.Background()
	vals := server.Cachestore.Client.LRange(ctx, key, 0, 100)
	if vals == nil {
		return nil
	}
	fmt.Println("VALS>>>>>>>", vals)
	str_vals := vals.Val()
	if len(str_vals) == 0 {
		return nil
	}
	suggs := make([]*suggestion.Suggestion, 0, len(str_vals))

	for _, val := range str_vals {
		suggs = append(suggs, &suggestion.Suggestion{Text: val})
	}
	resp := &suggestion.Response{
		Query:       key,
		Suggestions: suggs,
	}
	return resp
}
