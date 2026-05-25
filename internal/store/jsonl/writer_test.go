package jsonl_test

import (
	"os"
	"path/filepath"
	"testing"

	"mining-pipeline/internal/store/jsonl"
)

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.jsonl")
	if err := jsonl.WriteFileAtomic(path, []byte("{\"a\":1}\n")); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "{\"a\":1}\n" {
		t.Fatalf("got %q", b)
	}
}
