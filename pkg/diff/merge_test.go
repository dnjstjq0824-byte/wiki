package diff

import (
	"reflect"
	"testing"
)

func TestMerge3(t *testing.T) {
	// Case 1: No conflict, both match
	base := []string{"line1", "line2", "line3"}
	ours := []string{"line1", "line2", "line3"}
	theirs := []string{"line1", "line2", "line3"}

	merged, conflicts := Merge3(base, ours, theirs)
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(conflicts))
	}
	if !reflect.DeepEqual(merged, ours) {
		t.Errorf("merged mismatch: %+v", merged)
	}

	// Case 2: Only ours changed line2
	oursChanged := []string{"line1", "line2_ours", "line3"}
	merged2, conflicts2 := Merge3(base, oursChanged, theirs)
	if len(conflicts2) != 0 {
		t.Errorf("expected 0 conflicts, got %d", len(conflicts2))
	}
	if !reflect.DeepEqual(merged2, oursChanged) {
		t.Errorf("expected ours changes applied, got: %+v", merged2)
	}

	// Case 3: Both changed line2 differently -> conflict
	theirsChanged := []string{"line1", "line2_theirs", "line3"}
	merged3, conflicts3 := Merge3(base, oursChanged, theirsChanged)
	if len(conflicts3) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(conflicts3))
	}
	if conflicts3[0].Ours != "line2_ours" || conflicts3[0].Theirs != "line2_theirs" {
		t.Errorf("unexpected conflict content: %+v", conflicts3[0])
	}
	if len(merged3) < 5 {
		t.Errorf("expected conflict markers in merged result, got: %+v", merged3)
	}
}

