package suggestion

import (
	"bytes"
	"encoding/gob"
)

type Response struct {
	Query       string        `json:"query"`
	Suggestions []*Suggestion `json:"suggestions"`
}

type Suggestion struct {
	Text string `json:"text"`
}

func NewResponse(query string) *Response {
	return &Response{
		Query: query,
	}
}

func (s *Suggestion) String() string {
	return s.Text
}

func NewSuggestion(s string) *Suggestion {
	return &Suggestion{
		Text: s,
	}
}

func (s *Response) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}

	enc := gob.NewEncoder(buf)
	if err := enc.Encode(s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeResponse(s []byte) (*Response, error) {
	resp := &Response{}
	dec := gob.NewDecoder(bytes.NewReader(s))
	if err := dec.Decode(resp); err != nil {
		return nil, err
	}
	return resp, nil
}
