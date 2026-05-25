package jsonl

import (
	"bufio"
	"encoding/json"
	"os"

	"mining-pipeline/internal/model"
)

func LoadDocuments(path string) ([]model.Document, error) {
	return loadLines(path, func(line []byte) (model.Document, error) {
		var d model.Document
		err := json.Unmarshal(line, &d)
		return d, err
	})
}

func LoadChunks(path string) ([]model.Chunk, error) {
	return loadLines(path, func(line []byte) (model.Chunk, error) {
		var c model.Chunk
		err := json.Unmarshal(line, &c)
		return c, err
	})
}

func loadLines[T any](path string, parse func([]byte) (T, error)) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var out []T
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		v, err := parse(line)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, sc.Err()
}

func WriteDocuments(path string, docs []model.Document) error {
	return writeRecords(path, docs, func(d model.Document) ([]byte, error) {
		return json.Marshal(d)
	})
}

func WriteChunks(path string, chunks []model.Chunk) error {
	return writeRecords(path, chunks, func(c model.Chunk) ([]byte, error) {
		return json.Marshal(c)
	})
}

func writeRecords[T any](path string, items []T, marshal func(T) ([]byte, error)) error {
	var buf []byte
	for _, item := range items {
		b, err := marshal(item)
		if err != nil {
			return err
		}
		buf = append(buf, b...)
		buf = append(buf, '\n')
	}
	return WriteFileAtomic(path, buf)
}
