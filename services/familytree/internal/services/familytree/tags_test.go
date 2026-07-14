package familytree

import (
	"reflect"
	"testing"
)

func TestNormalizeTagCodes(t *testing.T) {
	got := normalizeTagCodes([]string{" Historical ", "ROYALTY", "historical", "", " royalty "})
	want := []string{"historical", "royalty"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeTagCodes() = %#v, want %#v", got, want)
	}
}
