package search

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestSearchAndAutocomplete(t *testing.T) {
	idx := NewSearchIndex()

	idx.Index(&SearchDocument{
		ID:        "doc-1",
		Namespace: 0,
		Title:     "Quantum Computing",
		Content:   "Quantum computing utilizes qubits, superposition, and entanglement to solve hard computational problems.",
	})

	idx.Index(&SearchDocument{
		ID:        "doc-2",
		Namespace: 0,
		Title:     "Classical Computing",
		Content:   "Classical computing utilizes binary bits to perform deterministic logic operations.",
	})

	// Search
	results := idx.Search("qubits", 0)
	if len(results) != 1 || results[0].Document.Title != "Quantum Computing" {
		t.Errorf("expected Quantum Computing result, got: %+v", results)
	}

	// Autocomplete
	ac := idx.Autocomplete("Quant", 5)
	if len(ac) != 1 || ac[0] != "Quantum Computing" {
		t.Errorf("expected autocomplete Quantum Computing, got: %+v", ac)
	}

	// Media Store
	media := NewMediaStore()
	media.Add(&MediaAsset{
		Filename:     "diagram.png",
		OriginalName: "quantum_circuit.png",
		MimeType:     "image/png",
		FileSize:     1024,
		LicenseType:  "CC-BY-SA-4.0",
	})

	list := media.List()
	if len(list) != 1 {
		t.Errorf("expected 1 media asset, got %d", len(list))
	}
}

func TestImageScrubAndThumbnail(t *testing.T) {
	// Create sample 100x100 PNG image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{R: 15, G: 118, B: 110, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode failed: %v", err)
	}
	rawPNG := buf.Bytes()

	// 1. Test ScrubEXIF
	scrubbed, err := ScrubEXIF(rawPNG, "png")
	if err != nil {
		t.Fatalf("ScrubEXIF failed: %v", err)
	}
	if len(scrubbed) == 0 {
		t.Errorf("scrubbed PNG is empty")
	}

	// 2. Test GenerateThumbnail
	thumb, err := GenerateThumbnail(rawPNG, "png", 50, 50)
	if err != nil {
		t.Fatalf("GenerateThumbnail failed: %v", err)
	}
	if len(thumb) == 0 {
		t.Errorf("thumbnail PNG is empty")
	}
}

