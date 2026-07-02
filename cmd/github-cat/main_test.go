package main

import (
	"bytes"
	"reflect"
	"regexp"
	"testing"

	"github.com/google/go-github/v71/github"
)

func TestSplitLines(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"a\n", []string{"a"}},
		{"a\nb\n", []string{"a", "b"}},
		{"a\nb", []string{"a", "b"}},
		{"a\n\nb\n", []string{"a", "", "b"}},
		{"2.6.3\n", []string{"2.6.3"}},
	}
	for _, c := range cases {
		got := splitLines(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitLines(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

func TestFilterRegex(t *testing.T) {
	if got := filterRegex("a\nb\n", nil); got != "a\nb\n" {
		t.Errorf("nil regex should return content unchanged, got %q", got)
	}

	re := regexp.MustCompile("^b")
	if got := filterRegex("apple\nbanana\nberry\ncherry", re); got != "banana\nberry" {
		t.Errorf("filterRegex = %q, want %q", got, "banana\nberry")
	}
}

func TestPrintText(t *testing.T) {
	results := []result{
		{Repo: "someorg/app-foo", Content: "2.6.3\n"},
		{Repo: "someorg/app-bar", Content: "2.4.6\n"},
	}

	var withPrefix bytes.Buffer
	if err := printText(&withPrefix, results, true); err != nil {
		t.Fatal(err)
	}
	want := "someorg/app-foo:2.6.3\nsomeorg/app-bar:2.4.6\n"
	if withPrefix.String() != want {
		t.Errorf("printText with prefix = %q, want %q", withPrefix.String(), want)
	}

	var noPrefix bytes.Buffer
	if err := printText(&noPrefix, results, false); err != nil {
		t.Fatal(err)
	}
	want = "2.6.3\n2.4.6\n"
	if noPrefix.String() != want {
		t.Errorf("printText no prefix = %q, want %q", noPrefix.String(), want)
	}
}

func TestPrintTextMultiline(t *testing.T) {
	results := []result{
		{Repo: "someorg/app-foo", Content: "FROM ruby:2.6.3\nCOPY . /app/\nCMD [\"foo\"]\n"},
	}
	var buf bytes.Buffer
	if err := printText(&buf, results, true); err != nil {
		t.Fatal(err)
	}
	want := "someorg/app-foo:FROM ruby:2.6.3\nsomeorg/app-foo:COPY . /app/\nsomeorg/app-foo:CMD [\"foo\"]\n"
	if buf.String() != want {
		t.Errorf("printText multiline = %q, want %q", buf.String(), want)
	}
}

func TestPrintJSON(t *testing.T) {
	results := []result{
		{Repo: "someorg/app-foo", Content: "2.6.3\n"},
	}
	var buf bytes.Buffer
	if err := printJSON(&buf, results); err != nil {
		t.Fatal(err)
	}
	want := `[{"repo":"someorg/app-foo","content":"2.6.3\n"}]` + "\n"
	if buf.String() != want {
		t.Errorf("printJSON = %q, want %q", buf.String(), want)
	}
}

func TestPrintJSONEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := printJSON(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "[]\n" {
		t.Errorf("printJSON(nil) = %q, want %q", buf.String(), "[]\n")
	}
}

func TestAppendFilteredRepos(t *testing.T) {
	repos := []*github.Repository{
		{Name: github.Ptr("plain")},
		{Name: github.Ptr("archived"), Archived: github.Ptr(true)},
		{Name: github.Ptr("forked"), Fork: github.Ptr(true)},
	}

	cases := []struct {
		name string
		opts options
		want []string
	}{
		{"defaults exclude archived and forks", options{}, []string{"plain"}},
		{"include archived", options{includeArchived: true}, []string{"plain", "archived"}},
		{"include forks", options{includeForks: true}, []string{"plain", "forked"}},
		{"include both", options{includeArchived: true, includeForks: true}, []string{"plain", "archived", "forked"}},
	}

	for _, c := range cases {
		got := appendFilteredRepos(nil, repos, c.opts)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %#v, want %#v", c.name, got, c.want)
		}
	}
}
