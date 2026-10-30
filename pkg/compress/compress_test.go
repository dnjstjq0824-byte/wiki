package compress

import (
	"bytes"
	"testing"
)

func TestCompressDecompress(t *testing.T) {
	original := []byte("Treewiki is a structured collaborative knowledge and reference platform designed to organize human knowledge into clear, interconnected topic trees.")

	compressed, err := Compress(original)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("Decompressed bytes do not match original! Got %s", string(decompressed))
	}
}

