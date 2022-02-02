package cell

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	okrav1alpha1 "github.com/mumoshu/okra/api/v1alpha1"
)

func TestValueFrom(t *testing.T) {
	c := &okrav1alpha1.Cell{
		Status: okrav1alpha1.CellStatus{
			DesiredVersion: "1.2.3",
		},
	}

	want := "1.2.3"
	got, err := extractValueFromCell(c, "status.desiredVersion")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("unexpected diff: %s", d)
	}
}
