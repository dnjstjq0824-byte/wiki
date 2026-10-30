package diff

// Conflict represents a merge conflict region
type Conflict struct {
	Base    string `json:"base"`
	Ours    string `json:"ours"`
	Theirs  string `json:"theirs"`
}

// Merge3 performs a simple line-based 3-way merge.
// Returns merged lines and a list of conflicts (if any).
func Merge3(base, ours, theirs []string) ([]string, []Conflict) {
	// Simple approach: if ours==theirs, take theirs; if only one changed, take changed;
	// if both changed differently, emit conflict markers.
	maxLen := len(base)
	if len(ours) > maxLen {
		maxLen = len(ours)
	}
	if len(theirs) > maxLen {
		maxLen = len(theirs)
	}

	merged := []string{}
	conflicts := []Conflict{}

	for i := 0; i < maxLen; i++ {
		var b, o, t string
		if i < len(base) {
			b = base[i]
		}
		if i < len(ours) {
			o = ours[i]
		}
		if i < len(theirs) {
			t = theirs[i]
		}

		if o == t {
			merged = append(merged, o)
			continue
		}
		if o == b && t != b {
			merged = append(merged, t)
			continue
		}
		if t == b && o != b {
			merged = append(merged, o)
			continue
		}
		if o == b && t == b {
			merged = append(merged, b)
			continue
		}
		// conflict
		merged = append(merged, "<<<<<<< ours")
		merged = append(merged, o)
		merged = append(merged, "=======")
		merged = append(merged, t)
		merged = append(merged, ">>>>>>> theirs")
		conflicts = append(conflicts, Conflict{Base: b, Ours: o, Theirs: t})
	}

	return merged, conflicts
}