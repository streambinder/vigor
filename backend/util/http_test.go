package util

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
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
<p>Each training week follows a fixed wave of percentages with an
assistance block after the main lift, and every fourth week is a deload
that lets accumulated fatigue clear before the next cycle starts over.</p>
</article>
<footer><p>Copyright 2026</p></footer>
</body></html>`

	joined, err := extractMainText(page, "https://example.com/wendler")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
	if _, err := extractMainText(`<html><body><script>only()</script></body></html>`, "https://example.com/x"); err != ErrNoProgramContent {
		t.Fatalf("expected ErrNoProgramContent, got %v", err)
	}
}

func TestExtractMainTextJSONLDBody(t *testing.T) {
	// regression: promo widgets wrapped in bare <article> tags used to poison
	// the DOM walk (only nav bylines survived) — the real body must come
	// from the JSON-LD articleBody fallback instead
	page, err := os.ReadFile("testdata/extractor_jsonld_fallback.html")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	joined, err := extractMainText(string(page), "https://example.com/article")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"1 trazione alla sbarra", "50 squat"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected %q in extracted lines: %q", want, joined)
		}
	}
	for _, unwanted := range []string{"l'editoriale", "Il Borghese", "Abbonamenti"} {
		if strings.Contains(joined, unwanted) {
			t.Fatalf("promo box text must not leak into extraction: %q", joined)
		}
	}
}

func TestExtractMainTextPicksRichestContainer(t *testing.T) {
	var realBody strings.Builder
	realBody.WriteString("Back squat 5x5 at 80%, rest 3 minutes between heavy sets.")
	for range 20 {
		realBody.WriteString(" Volume and intensity notes carry the rest of the article")
	}

	page := `<html><body>
<article><h2>Newsletter</h2><p>Iscriviti</p></article>
<article><h1>Il programma</h1><p>` + realBody.String() + `</p></article>
<article><h2>Shop</h2><p>Abbonamenti digitali</p></article>
</body></html>`

	joined, err := extractMainText(page, "https://example.com/program")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(joined, "Back squat 5x5 at 80%") {
		t.Fatalf("expected the richest container to win: %q", joined)
	}
	for _, unwanted := range []string{"Newsletter", "Abbonamenti"} {
		if strings.Contains(joined, unwanted) {
			t.Fatalf("promo container text must not leak into extraction: %q", joined)
		}
	}
}

func TestExtractMainTextThinExtraction(t *testing.T) {
	// no JSON-LD and no meaningful content: a thin extraction is a failed
	// extraction, so the caller can treat the resource as unusable
	page := `<html><body>
<article><h2>l'editoriale</h2></article>
<article><h2>Il Borghese</h2></article>
</body></html>`

	if _, err := extractMainText(page, "https://example.com/thin"); err != ErrNoProgramContent {
		t.Fatalf("expected ErrNoProgramContent, got %v", err)
	}
}

func TestExtractMainTextKeepsConsecutiveSchemeLines(t *testing.T) {
	// regression: enumerated scheme lines (digit-led, one movement each)
	// must all survive extraction: a relevance filter used to collapse
	// consecutive short lines into the last one, dropping the scheme head
	page := `<html><head><title>t</title></head><body>
<nav><ul><li>Home</li><li>Shop</li></ul></nav>
<article>
<h1>Il circuito a corpo libero</h1>
<p>Imposta un timer di venti minuti ed esegui la sequenza a oltranza,
ripartendo dal primo esercizio non appena concluso l'ultimo, cercando di
chiudere il maggior numero di giri completi nel tempo a disposizione.</p>
<p>5 trazioni</p>
<p>10 piegamenti</p>
<p>15 squat a corpo libero</p>
<p>La qualita del movimento resta il vincolo principale: meglio un giro
in meno che ripetizioni sporche, soprattutto quando la fatica accumulata
nella seconda meta del lavoro inizia a farsi sentire sulle spalle.</p>
</article>
<footer><p>Copyright 2026</p></footer>
</body></html>`

	got, err := extractMainText(page, "https://example.com/circuit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"5 trazioni", "10 piegamenti", "15 squat a corpo libero"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in extracted text: %q", want, got)
		}
	}
	for _, unwanted := range []string{"Home", "Copyright"} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("did not expect %q in extracted text: %q", unwanted, got)
		}
	}
}

func TestExtractMainTextCap(t *testing.T) {
	var body strings.Builder
	body.WriteString(`<html><body><article><h1>Long read</h1>`)
	for range 200 {
		body.WriteString("<p>Squat 5x5 with 3 minutes rest between heavy sets and steady tempo.</p>")
	}
	body.WriteString(`</article></body></html>`)

	got, err := extractMainText(body.String(), "https://example.com/long")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) > maxArticleLength {
		t.Fatalf("expected output capped at %d, got %d", maxArticleLength, len(got))
	}
	if !strings.HasSuffix(got, "tempo.") {
		t.Fatalf("expected the cap to cut at a line boundary, got tail %q", got[max(0, len(got)-40):])
	}
}
