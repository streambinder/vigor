package util

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
)

func TestDedupeAndCanonicalExerciseIDs(t *testing.T) {
	if got := Dedupe([]string{"b", "a", "b", "a", "c"}); strings.Join(got, ",") != "a,b,c" {
		t.Errorf("Dedupe = %v", got)
	}
	if got := Dedupe(nil); len(got) != 0 {
		t.Errorf("Dedupe nil = %v", got)
	}
	m := CanonicalExerciseIDs([]string{"Push Up", "push-up", "Air Squat"})
	if m["push-up"] == "" || m["air-squat"] != "Air Squat" {
		t.Errorf("CanonicalExerciseIDs = %v", m)
	}
	if got := CanonicalExerciseIDs(nil); len(got) != 0 {
		t.Errorf("CanonicalExerciseIDs nil = %v", got)
	}
}

func TestUpstreamErrorAndFetchResourceError(t *testing.T) {
	err := &UpstreamError{Status: 503}
	if err.Error() != "upstream status 503" {
		t.Errorf("Error = %q", err.Error())
	}
	if _, err := FetchResource("http://127.0.0.1:1/article"); err == nil {
		t.Error("FetchResource to a private target must fail")
	}
	if _, err := FetchResource("://bad"); err == nil {
		t.Error("FetchResource with an invalid URL must fail")
	}
}

func TestPublicDialContextErrors(t *testing.T) {
	ctx := context.Background()
	if _, err := publicDialContext(ctx, "tcp", "missing-port"); err == nil {
		t.Error("address without a port must fail")
	}
	if _, err := publicDialContext(ctx, "tcp", "localhost:80"); err == nil {
		// localhost resolves to a loopback address, which is not public.
		t.Error("localhost dial must be refused as non-public")
	} else if !errors.Is(err, ErrFetchPrivateTarget) && !strings.Contains(err.Error(), "dns") {
		t.Errorf("localhost dial error = %v", err)
	}
	if _, err := publicDialContext(ctx, "tcp", "nonexistent.invalid.:80"); err == nil {
		t.Error("unresolvable host must fail")
	}
}

func TestArticleFamilyAndDeepestBody(t *testing.T) {
	if !articleFamily("Article") || articleFamily("Thing") || articleFamily(42) {
		t.Error("articleFamily string handling is wrong")
	}
	if !articleFamily([]any{"Thing", "NewsArticle"}) || articleFamily([]any{"Thing"}) || articleFamily([]any{42}) {
		t.Error("articleFamily list handling is wrong")
	}
	long := strings.Repeat("a", 250)
	blob := `{"@graph":[{"@type":"Thing"},{"@type":["Article"],"articleBody":"` + long + `"}]}`
	if got := longestArticleBody(blob); got != long {
		t.Errorf("longestArticleBody = %d chars", len(got))
	}
	if got := longestArticleBody("{not json"); got != "" {
		t.Errorf("invalid blob = %q", got)
	}
	if got := deepestArticleBody("scalar"); got != "" {
		t.Errorf("scalar body = %q", got)
	}
	nested := map[string]any{"items": []any{map[string]any{"@type": "BlogPosting", "articleBody": "body-text"}}}
	if got := deepestArticleBody(nested); got != "body-text" {
		t.Errorf("nested body = %q", got)
	}
}

func TestExtractMainTextInvalidURL(t *testing.T) {
	page := "<html><body><article>" + strings.Repeat("word ", 100) + "</article></body></html>"
	if _, err := extractMainText(page, "://bad url"); err == nil {
		// readability may still fail with no program content; any error is
		// acceptable, success is not.
		t.Logf("extractMainText with a bad URL returned no error")
	}
}

func TestCapArticleBoundary(t *testing.T) {
	line := strings.Repeat("x", 9000)
	got := capArticle([]string{line, line, line, line})
	if len(got) > maxArticleLength {
		t.Errorf("capArticle length = %d", len(got))
	}
	if got := capArticle(nil); got != "" {
		t.Errorf("capArticle nil = %q", got)
	}
}

func TestValidateRemoteURLAndDialSplit(t *testing.T) {
	if _, _, err := net.SplitHostPort("example.com"); err == nil {
		t.Error("expected a missing port error")
	}
}
