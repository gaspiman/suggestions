package lang

import (
	"bytes"
	"strings"
)

func (l *Lang) Chunker(s string) []string {
	/*
		chunks := []string{}
		for _, chunk := range Splitter(s, 0) {
			if !l.LoadOrStore(chunk) {
				chunks = append(chunks, chunk)
			}
		}
	*/
	chunks := Splitter(s, 0)
	return chunks
}

func Splitter(s string, count int) []string {
	parts := strings.Split(s, " ")
	if len(parts) > 4 {
		return nil
	}
	chunks := []string{}
	chunk := []byte{}
	for i := 0; i < len(s); i++ {
		chunk = append(chunk, s[i])
		if bytes.Equal([]byte{s[i]}, []byte(" ")) {
			continue
		}
		str_chunk := string(chunk)
		if str_chunk == s && count == 0 {
			continue
		}
		chunks = append(chunks, string(chunk))
	}
	if len(parts) > 1 {
		s = strings.Join(parts[1:], " ")
		chunks = append(chunks, Splitter(s, count+1)...)
	}
	return chunks
}
