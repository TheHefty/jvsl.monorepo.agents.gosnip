package snippet

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
)

const (
	maxCodeBytes = 1 << 20
	maxTextRunes = 100
	maxTags      = 20
)

// Draft is untrusted snippet input.
type Draft struct {
	Name     string
	Code     []byte
	Language string
	Tags     []string
}

// NewSnippet is normalized input accepted for persistence.
type NewSnippet struct {
	name     string
	nameKey  string
	code     []byte
	language string
	tags     []string
}

// Snippet is a persisted snippet.
type Snippet struct {
	ID        int64
	Name      string
	NameKey   string
	Code      []byte
	Language  string
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ValidationError identifies one invalid draft field without including snippet code.
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Field, e.Reason)
}

// Normalize validates and canonicalizes a draft.
func Normalize(draft Draft) (NewSnippet, error) {
	if len(draft.Code) == 0 {
		return NewSnippet{}, invalid("code", "must not be empty")
	}
	if len(draft.Code) > maxCodeBytes {
		return NewSnippet{}, invalid("code", "must not exceed 1 MiB")
	}
	if !utf8.Valid(draft.Code) {
		return NewSnippet{}, invalid("code", "must be valid UTF-8")
	}

	name := strings.TrimSpace(draft.Name)
	if name == "" {
		return NewSnippet{}, invalid("name", "must not be empty")
	}
	if utf8.RuneCountInString(name) > maxTextRunes {
		return NewSnippet{}, invalid("name", "must not exceed 100 Unicode characters")
	}
	decimalOnly := true
	for _, r := range name {
		if !unicode.IsDigit(r) {
			decimalOnly = false
			break
		}
	}
	if decimalOnly {
		return NewSnippet{}, invalid("name", "must not contain only decimal digits")
	}

	language := strings.ToLower(strings.TrimSpace(draft.Language))
	if language == "" {
		return NewSnippet{}, invalid("language", "must not be empty")
	}
	if utf8.RuneCountInString(language) > maxTextRunes {
		return NewSnippet{}, invalid("language", "must not exceed 100 Unicode characters")
	}
	for _, r := range language {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && !strings.ContainsRune("+#._-", r) {
			return NewSnippet{}, invalid("language", "contains a disallowed character")
		}
	}

	if len(draft.Tags) > maxTags {
		return NewSnippet{}, invalid("tags", "must not contain more than 20 values")
	}
	tagSet := make(map[string]struct{}, len(draft.Tags))
	for _, raw := range draft.Tags {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if tag == "" {
			return NewSnippet{}, invalid("tags", "must not contain an empty value")
		}
		if utf8.RuneCountInString(tag) > maxTextRunes {
			return NewSnippet{}, invalid("tags", "values must not exceed 100 Unicode characters")
		}
		for _, r := range tag {
			if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '-' {
				return NewSnippet{}, invalid("tags", "a value contains a disallowed character")
			}
		}
		tagSet[tag] = struct{}{}
	}
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	slices.Sort(tags)

	return NewSnippet{
		name:     name,
		nameKey:  cases.Fold().String(name),
		code:     bytesClone(draft.Code),
		language: language,
		tags:     tags,
	}, nil
}

func invalid(field, reason string) error {
	return &ValidationError{Field: field, Reason: reason}
}

func bytesClone(value []byte) []byte {
	return append([]byte(nil), value...)
}

func (s NewSnippet) Name() string     { return s.name }
func (s NewSnippet) NameKey() string  { return s.nameKey }
func (s NewSnippet) Code() []byte     { return bytesClone(s.code) }
func (s NewSnippet) Language() string { return s.language }
func (s NewSnippet) Tags() []string   { return slices.Clone(s.tags) }
