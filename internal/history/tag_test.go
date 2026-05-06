package history

import (
	"testing"
)

func TestParseTags_Valid(t *testing.T) {
	tags, err := ParseTags("env=prod,region=us-east-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags.Get("env") != "prod" {
		t.Errorf("expected env=prod, got %q", tags.Get("env"))
	}
	if tags.Get("region") != "us-east-1" {
		t.Errorf("expected region=us-east-1, got %q", tags.Get("region"))
	}
}

func TestParseTags_Empty(t *testing.T) {
	tags, err := ParseTags("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tags != nil {
		t.Errorf("expected nil tags for empty input")
	}
}

func TestParseTags_Invalid(t *testing.T) {
	_, err := ParseTags("noequals")
	if err == nil {
		t.Error("expected error for invalid tag format")
	}
}

func TestParseTags_Sorted(t *testing.T) {
	tags, err := ParseTags("z=last,a=first,m=middle")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tags[0].Key != "a" || tags[1].Key != "m" || tags[2].Key != "z" {
		t.Errorf("tags not sorted: %v", tags)
	}
}

func TestTags_String(t *testing.T) {
	tags := Tags{
		{Key: "env", Value: "prod"},
		{Key: "region", Value: "us-east-1"},
	}
	got := tags.String()
	want := "env=prod,region=us-east-1"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestTags_Get_Missing(t *testing.T) {
	tags := Tags{{Key: "env", Value: "prod"}}
	if v := tags.Get("missing"); v != "" {
		t.Errorf("expected empty string, got %q", v)
	}
}
