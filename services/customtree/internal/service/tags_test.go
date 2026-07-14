package service

import (
	"reflect"
	"testing"
)

func TestNormalizeTags(t *testing.T) {
	got := normalizeTags([]string{" Academic ", "MENTORSHIP", "academic", ""})
	want := []string{"academic", "mentorship"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeTags() = %#v, want %#v", got, want)
	}
}
