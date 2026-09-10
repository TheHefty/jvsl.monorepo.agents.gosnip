package snippet

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestNormalizeCanonicalizesAValidDraft(t *testing.T) {
	t.Parallel()

	code := []byte(" fmt.Println(\"Olá\")\r\n")
	got, err := Normalize(Draft{
		Name:     "  Auth Token  ",
		Code:     code,
		Language: " Go ",
		Tags:     []string{" CLI ", "examples", "cli"},
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	if got.Name() != "Auth Token" {
		t.Errorf("Name() = %q", got.Name())
	}
	if got.NameKey() != "auth token" {
		t.Errorf("NameKey() = %q", got.NameKey())
	}
	if got.Language() != "go" {
		t.Errorf("Language() = %q", got.Language())
	}
	if want := []string{"cli", "examples"}; !slicesEqual(got.Tags(), want) {
		t.Errorf("Tags() = %q, want %q", got.Tags(), want)
	}
	if !bytes.Equal(got.Code(), code) {
		t.Errorf("Code() changed bytes: %q", got.Code())
	}

	code[0] = 'X'
	if got.Code()[0] == 'X' {
		t.Error("NewSnippet retained the caller's code slice")
	}
}

func TestNormalizeUsesUnicodeCaseFolding(t *testing.T) {
	t.Parallel()

	upper, err := Normalize(validDraft("Straße"))
	if err != nil {
		t.Fatal(err)
	}
	lower, err := Normalize(validDraft("STRASSE"))
	if err != nil {
		t.Fatal(err)
	}
	if upper.NameKey() != lower.NameKey() {
		t.Fatalf("case-folded keys differ: %q != %q", upper.NameKey(), lower.NameKey())
	}
}

func TestNormalizeRejectsInvalidDrafts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*Draft)
		field  string
	}{
		{"empty code", func(d *Draft) { d.Code = nil }, "code"},
		{"invalid UTF-8", func(d *Draft) { d.Code = []byte{0xff} }, "code"},
		{"code over one MiB", func(d *Draft) { d.Code = bytes.Repeat([]byte{'x'}, maxCodeBytes+1) }, "code"},
		{"empty name", func(d *Draft) { d.Name = " \t" }, "name"},
		{"decimal name", func(d *Draft) { d.Name = "１２3" }, "name"},
		{"long name", func(d *Draft) { d.Name = strings.Repeat("界", maxTextRunes+1) }, "name"},
		{"missing language", func(d *Draft) { d.Language = " " }, "language"},
		{"long language", func(d *Draft) { d.Language = strings.Repeat("界", maxTextRunes+1) }, "language"},
		{"invalid language", func(d *Draft) { d.Language = "c lang" }, "language"},
		{"too many tags", func(d *Draft) { d.Tags = make([]string, maxTags+1) }, "tags"},
		{"empty tag", func(d *Draft) { d.Tags = []string{" "} }, "tags"},
		{"long tag", func(d *Draft) { d.Tags = []string{strings.Repeat("界", maxTextRunes+1)} }, "tags"},
		{"comma tag", func(d *Draft) { d.Tags = []string{"one,two"} }, "tags"},
		{"invalid tag", func(d *Draft) { d.Tags = []string{"c++"} }, "tags"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			draft := validDraft("Snippet")
			test.mutate(&draft)
			_, err := Normalize(draft)
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("Normalize() error = %v, want ValidationError", err)
			}
			if validationErr.Field != test.field {
				t.Errorf("ValidationError.Field = %q, want %q", validationErr.Field, test.field)
			}
		})
	}
}

func validDraft(name string) Draft {
	return Draft{Name: name, Code: []byte("code\n"), Language: "Go", Tags: []string{"test"}}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
