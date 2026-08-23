package util

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"
)

func TestSafeFetchHappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html>ok</html>"))
	}))
	defer server.Close()

	body, contentType, err := SafeFetch(server.URL, SafeFetchOptions{
		AllowPlainHTTP:    true,
		Timeout:           5 * time.Second,
		ContentTypePrefix: "text/html",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(body), "ok") {
		t.Fatalf("unexpected body: %q", body)
	}
	if !strings.HasPrefix(contentType, "text/html") {
		t.Fatalf("unexpected content type: %q", contentType)
	}
}

func TestSafeFetchAllowlist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	_, _, err := SafeFetch(server.URL, SafeFetchOptions{
		AllowedDomains: []string{"exercisedb.dev"},
		AllowPlainHTTP: true,
	})
	if !errors.Is(err, ErrFetchDomainNotAllowed) {
		t.Fatalf("expected ErrFetchDomainNotAllowed, got %v", err)
	}

	host := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")[0]
	_, _, err = SafeFetch(server.URL, SafeFetchOptions{
		AllowedDomains: []string{host},
		AllowPlainHTTP: true,
	})
	if err != nil {
		t.Fatalf("expected allowed host to pass, got %v", err)
	}
}

func TestSafeFetchBlocksPrivateTargets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	// httptest serves on 127.0.0.1: with private-IP blocking on, the dial
	// must be refused even though the URL itself parses fine
	_, _, err := SafeFetch(server.URL, SafeFetchOptions{
		AllowPlainHTTP:  true,
		BlockPrivateIPs: true,
		Timeout:         5 * time.Second,
	})
	if !errors.Is(err, ErrFetchPrivateTarget) {
		t.Fatalf("expected ErrFetchPrivateTarget, got %v", err)
	}
}

func TestSafeFetchSchemeEnforcement(t *testing.T) {
	if _, _, err := SafeFetch("http://example.com", SafeFetchOptions{}); !errors.Is(err, ErrFetchInvalidScheme) {
		t.Fatalf("expected ErrFetchInvalidScheme for plain http, got %v", err)
	}
	if _, _, err := SafeFetch("ftp://example.com/file", SafeFetchOptions{}); !errors.Is(err, ErrFetchInvalidScheme) {
		t.Fatalf("expected ErrFetchInvalidScheme for ftp, got %v", err)
	}
}

func TestSafeFetchSizeCap(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 4096)))
	}))
	defer server.Close()

	_, _, err := SafeFetch(server.URL, SafeFetchOptions{
		AllowPlainHTTP: true,
		MaxBytes:       1024,
	})
	if !errors.Is(err, ErrFetchTooLarge) {
		t.Fatalf("expected ErrFetchTooLarge, got %v", err)
	}
}

func TestSafeFetchRedirectLimit(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, server.URL+"/next", http.StatusFound)
	}))
	defer server.Close()

	_, _, err := SafeFetch(server.URL, SafeFetchOptions{
		AllowPlainHTTP: true,
		MaxRedirects:   0,
	})
	if !errors.Is(err, ErrFetchTooManyRedirects) {
		t.Fatalf("expected ErrFetchTooManyRedirects, got %v", err)
	}
}

func TestSafeFetchContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	}))
	defer server.Close()

	_, _, err := SafeFetch(server.URL, SafeFetchOptions{
		AllowPlainHTTP:    true,
		ContentTypePrefix: "text/html",
	})
	if !errors.Is(err, ErrFetchBadContentType) {
		t.Fatalf("expected ErrFetchBadContentType, got %v", err)
	}
}

func TestIsPublicIP(t *testing.T) {
	public := []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111"}
	for _, s := range public {
		ip, err := netip.ParseAddr(s)
		if err != nil || !isPublicIP(ip) {
			t.Fatalf("expected %s to be public", s)
		}
	}
	blocked := []string{
		"127.0.0.1", "10.0.0.5", "172.16.3.4", "192.168.1.10",
		"169.254.169.254", "0.0.0.0", "::1", "fd00::1", "fe80::1",
		"100.64.0.1", "198.18.0.1",
	}
	for _, s := range blocked {
		ip, err := netip.ParseAddr(s)
		if err == nil && isPublicIP(ip) {
			t.Fatalf("expected %s to be blocked", s)
		}
	}
}

func TestExtractURLs(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "https passes through",
			text: "follow https://example.com/program please",
			want: []string{"https://example.com/program"},
		},
		{
			name: "http is upgraded",
			text: "see http://example.com/5x5",
			want: []string{"https://example.com/5x5"},
		},
		{
			name: "markdown parens and trailing punctuation stripped",
			text: "read [this](https://example.com/a). also https://example.com/b, thanks",
			want: []string{"https://example.com/a", "https://example.com/b"},
		},
		{
			name: "dedupes",
			text: "https://example.com/x and again https://example.com/x",
			want: []string{"https://example.com/x"},
		},
		{
			name: "capped at max",
			text: "https://a.example.com/1 https://b.example.com/2 https://c.example.com/3 https://d.example.com/4",
			want: []string{"https://a.example.com/1", "https://b.example.com/2", "https://c.example.com/3"},
		},
		{
			name: "none",
			text: "just train legs hard",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractURLs(tt.text)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d urls, got %d (%v)", len(tt.want), len(got), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("url %d: expected %q, got %q", i, tt.want[i], got[i])
				}
			}
		})
	}
}

func TestExtractMainText(t *testing.T) {
	page := `<html><head><title>t</title><script>var x = 1;</script></head>
<body>
<nav><ul><li>Home</li></ul></nav>
<article>
<h1> Wendler 5/3/1 </h1>
<p>The classic strength template.</p>
<ul>
<li>Squat 3x5 at 85%</li>
</ul>
</article>
<footer><p>Copyright 2026</p></footer>
</body></html>`

	lines, err := extractMainText(page)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"Wendler 5/3/1", "classic strength template", "Squat 3x5 at 85%"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected %q in extracted lines: %q", want, joined)
		}
	}
	for _, unwanted := range []string{"var x", "Home", "Copyright"} {
		if strings.Contains(joined, unwanted) {
			t.Fatalf("did not expect %q in extracted lines: %q", unwanted, joined)
		}
	}
}

func TestExtractMainTextEmpty(t *testing.T) {
	if _, err := extractMainText(`<html><body><script>only()</script></body></html>`); err != ErrNoProgramContent {
		t.Fatalf("expected ErrNoProgramContent, got %v", err)
	}
}

func TestFilterProgram(t *testing.T) {
	lines := []string{
		"The Program",
		"Welcome to my blog about my journey through fitness and wellness over the years.",
		"Back squat 5x5 at 80% 1RM, rest 3 minutes between sets.",
		"Keep the bar path vertical throughout the lift to stay balanced.",
		"Bench press 3x8, rest 90s.",
	}

	got := FilterProgram(lines)
	for _, want := range []string{"The Program", "5x5 at 80%", "Bench press 3x8"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in filtered output: %q", want, got)
		}
	}
	if strings.Contains(got, "journey through fitness") {
		t.Fatalf("prose should have been filtered out: %q", got)
	}
	// no bar-path vocabulary: filtered too, despite sitting between kept lines
	if strings.Contains(got, "bar path") {
		t.Fatalf("neutral line should have been filtered out: %q", got)
	}
}

func TestFilterProgramFallback(t *testing.T) {
	lines := []string{"Just a story.", "No program vocabulary anywhere here at all."}
	got := FilterProgram(lines)
	if !strings.Contains(got, "Just a story.") {
		t.Fatalf("expected fallback to keep the article lead, got %q", got)
	}
}

func TestFilterProgramCap(t *testing.T) {
	var lines []string
	for range 200 {
		lines = append(lines, "Squat 5x5 with 3 minutes rest between heavy sets.")
	}
	if got := FilterProgram(lines); len(got) > maxCleanLength {
		t.Fatalf("expected output capped at %d, got %d", maxCleanLength, len(got))
	}
}
