package diff

import (
	"testing"
)

func TestDiffLines(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected int
	}{
		{
			name:     "identical texts",
			a:        []string{"line1", "line2"},
			b:        []string{"line1", "line2"},
			expected: 2,
		},
		{
			name:     "insertion",
			a:        []string{"line1", "line3"},
			b:        []string{"line1", "line2", "line3"},
			expected: 3,
		},
		{
			name:     "deletion",
			a:        []string{"line1", "to_delete", "line2"},
			b:        []string{"line1", "line2"},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := DiffLines(tt.a, tt.b)
			if len(res) != tt.expected {
				t.Errorf("DiffLines() returned %d chunks, want %d", len(res), tt.expected)
			}
		})
	}
}
