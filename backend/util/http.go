package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	readability "github.com/go-shiori/go-readability"
)

var (
	ErrFetchInvalidURL       = errors.New("invalid url")
	ErrFetchInvalidScheme    = errors.New("invalid scheme")
	ErrFetchDomainNotAllowed = errors.New("domain not allowed")
	ErrFetchPrivateTarget    = errors.New("target resolves to a non-public address")
	ErrFetchTooLarge         = errors.New("response too large")
	ErrFetchBadContentType   = errors.New("unexpected content type")
	ErrFetchTooManyRedirects = errors.New("too many redirects")
)

// UpstreamError reports a non-200 status from the remote server.
type UpstreamError struct {
	Status int
}

func (e *UpstreamError) Error() string {
	return "upstream status " + strconv.Itoa(e.Status)
}

// SafeFetchOptions configures a shared hardened outbound fetch.
// two validation styles: an AllowedDomains allowlist (image proxy) or private-IP
// blocking for arbitrary URLs (linked articles) — combine them freely.
type SafeFetchOptions struct {
	AllowedDomains    []string      // when non-empty, only these hosts (or subdomains) are fetched
	BlockPrivateIPs   bool          // resolve DNS per connection and refuse non-public targets
	AllowPlainHTTP    bool          // default false: https only
	MaxBytes          int64         // 0: no cap
	Timeout           time.Duration // covers the whole request
	ContentTypePrefix string        // response Content-Type must start with this when non-empty
	MaxRedirects      int           // 0 redirects allowed on zero value; negatives mean unlimited
}

var ipBlockPrefixes = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"), // CGNAT
	netip.MustParsePrefix("198.18.0.0/15"), // benchmarking
}

// isPublicIP reports whether ip is a routable public unicast address.
func isPublicIP(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsValid() || ip.IsLoopback() || ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() {
		return false
	}
	for _, prefix := range ipBlockPrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

// publicDialContext refuses dialing any host that resolves to a non-public
// address, pinning the connection to the already-validated IP.
func publicDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("dns lookup: %w", err)
	}
	if len(ips) == 0 {
		return nil, ErrFetchPrivateTarget
	}

	// refuse when any answer is non-public: a mixed answer means the
	// resolver (or attacker) can pin us to an internal target
	for _, ipAddr := range ips {
		parsed, err := netip.ParseAddr(ipAddr.IP.String())
		if err != nil || !isPublicIP(parsed) {
			return nil, ErrFetchPrivateTarget
		}
	}

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
}

// hostMatchesAllowlist reports whether host equals an allowed domain or is a
// subdomain of one.
func hostMatchesAllowlist(host string, allowed []string) bool {
	for _, domain := range allowed {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

// validateRemoteURL enforces scheme and allowlist rules on a fetch (or redirect)
// target. private-IP rules are enforced at dial time on every connection.
func validateRemoteURL(u *url.URL, opts SafeFetchOptions) error {
	if u == nil || u.Host == "" {
		return ErrFetchInvalidURL
	}
	if u.Scheme != "https" && !(opts.AllowPlainHTTP && u.Scheme == "http") {
		return ErrFetchInvalidScheme
	}
	if len(opts.AllowedDomains) > 0 && !hostMatchesAllowlist(u.Hostname(), opts.AllowedDomains) {
		return ErrFetchDomainNotAllowed
	}
	return nil
}

// SafeFetch performs a hardened outbound GET: scheme and domain validation,
// optional private-address blocking with per-hop redirect checks, a body size
// cap and a content-type assertion. Returns the body and its content type.
func SafeFetch(rawURL string, opts SafeFetchOptions) ([]byte, string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, "", ErrFetchInvalidURL
	}
	if err := validateRemoteURL(parsed, opts); err != nil {
		return nil, "", err
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{}
	if opts.BlockPrivateIPs {
		transport.DialContext = publicDialContext
	}
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if opts.MaxRedirects >= 0 && len(via) > opts.MaxRedirects {
				return ErrFetchTooManyRedirects
			}
			return validateRemoteURL(req.URL, opts)
		},
	}

	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", &UpstreamError{Status: resp.StatusCode}
	}

	contentType := resp.Header.Get("Content-Type")
	if opts.ContentTypePrefix != "" && !strings.HasPrefix(contentType, opts.ContentTypePrefix) {
		return nil, "", ErrFetchBadContentType
	}

	body := resp.Body
	if opts.MaxBytes > 0 {
		body = io.NopCloser(io.LimitReader(resp.Body, opts.MaxBytes+1))
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, "", err
	}
	if opts.MaxBytes > 0 && int64(len(data)) > opts.MaxBytes {
		return nil, "", ErrFetchTooLarge
	}

	return data, contentType, nil
}

const (
	// MaxResourceURLs caps the links extracted from a single request
	MaxResourceURLs = 3

	// MinArticleLength is the floor for usable article content: extracted
	// text shorter than this carries no program signal, so callers treat it
	// as a failed extraction rather than a thin article.
	MinArticleLength = 200

	// maxArticleLength caps the cleaned article text handed downstream: a
	// cost guard on prompt size, cut at a line boundary, never a relevance
	// filter.
	maxArticleLength = 20000

	fetchMaxBytes = 512 * 1024
	fetchTimeout  = 10 * time.Second
)

// ErrNoProgramContent reports a fetched page with no usable text.
var ErrNoProgramContent = errors.New("no usable content in resource")

var (
	// matches explicit scheme URLs inside free text, stopping at whitespace
	// or markdown delimiters
	urlPattern = regexp.MustCompile(`https?://[^\s"'<>\)\]]+`)

	whitespace = regexp.MustCompile(`\s+`)
)

// ExtractURLs pulls http(s) URLs out of free text: deduped, capped, with
// plain http always upgraded to https.
func ExtractURLs(text string) []string {
	var urls []string
	seen := make(map[string]bool)
	for _, match := range urlPattern.FindAllString(text, -1) {
		match = strings.TrimRight(match, `.,;:!?'"`)
		if match == "" {
			continue
		}
		match = strings.TrimPrefix(match, "http://")
		match = strings.TrimPrefix(match, "https://")
		match = "https://" + match
		if seen[match] {
			continue
		}
		seen[match] = true
		urls = append(urls, match)
		if len(urls) == MaxResourceURLs {
			break
		}
	}
	return urls
}

// FetchResource downloads an article and returns its cleaned main text,
// one block per line. any failure is fatal to the calling request.
func FetchResource(rawURL string) (string, error) {
	body, _, err := SafeFetch(rawURL, SafeFetchOptions{
		BlockPrivateIPs:   true,
		MaxBytes:          fetchMaxBytes,
		Timeout:           fetchTimeout,
		ContentTypePrefix: "text/html",
		MaxRedirects:      5,
	})
	if err != nil {
		return "", err
	}

	return extractMainText(string(body), rawURL)
}

// extractMainText returns the article's main text as one block per line.
// A publisher-declared JSON-LD articleBody is authoritative when present;
// otherwise Mozilla's Readability algorithm (via go-readability) isolates
// the main content. Text under MinArticleLength is a failed extraction.
func extractMainText(page, rawURL string) (string, error) {
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		return "", err
	}

	if body := ldArticleBody(doc); len(body) >= MinArticleLength {
		return capArticle(cleanLines(body)), nil
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", ErrNoProgramContent
	}
	article, err := readability.FromReader(strings.NewReader(page), parsed)
	if err != nil {
		return "", ErrNoProgramContent
	}
	lines := cleanLines(article.TextContent)
	if len(strings.Join(lines, "\n")) < MinArticleLength {
		return "", ErrNoProgramContent
	}
	return capArticle(lines), nil
}

// hasAttr reports whether an element carries the attribute value (compared
// case-insensitively).
func hasAttr(n *html.Node, key, value string) bool {
	for _, attr := range n.Attr {
		if attr.Key == key && strings.EqualFold(strings.TrimSpace(attr.Val), value) {
			return true
		}
	}
	return false
}

// articleTypes names the JSON-LD @type values expected to carry an articleBody.
var articleTypes = map[string]bool{
	"Article": true, "NewsArticle": true, "BlogPosting": true,
	"Report": true, "ScholarlyArticle": true, "SocialMediaPosting": true,
}

// ldArticleBody returns the publisher-declared article body from the page's
// JSON-LD metadata blocks, or the empty string when none is present.
func ldArticleBody(doc *html.Node) string {
	var body string
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" && hasAttr(n, "type", "application/ld+json") {
			var blob strings.Builder
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					blob.WriteString(c.Data)
				}
			}
			if found := longestArticleBody(blob.String()); len(found) > len(body) {
				body = found
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return body
}

// longestArticleBody parses a JSON-LD blob (object, array, or @graph) and
// returns the longest articleBody carried by an article-family node.
func longestArticleBody(blob string) string {
	var value any
	if err := json.Unmarshal([]byte(strings.TrimSpace(blob)), &value); err != nil {
		return ""
	}
	return deepestArticleBody(value)
}

func deepestArticleBody(value any) string {
	best := ""
	switch node := value.(type) {
	case []any:
		for _, item := range node {
			if body := deepestArticleBody(item); len(body) > len(best) {
				best = body
			}
		}
	case map[string]any:
		if articleFamily(node["@type"]) {
			if body, ok := node["articleBody"].(string); ok && len(body) > len(best) {
				best = body
			}
		}
		for _, item := range node {
			if body := deepestArticleBody(item); len(body) > len(best) {
				best = body
			}
		}
	}
	return best
}

// articleFamily reports whether a JSON-LD @type (single or list) names an
// article-family schema.
func articleFamily(value any) bool {
	switch typed := value.(type) {
	case string:
		return articleTypes[typed]
	case []any:
		for _, item := range typed {
			if name, ok := item.(string); ok && articleTypes[name] {
				return true
			}
		}
	}
	return false
}

func normalizeLine(s string) string {
	return strings.TrimSpace(whitespace.ReplaceAllString(s, " "))
}

// cleanLines splits text into normalized non-empty lines.
func cleanLines(text string) []string {
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		if line = normalizeLine(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// capArticle joins lines up to maxArticleLength, cutting at a line
// boundary so downstream text never ends mid-sentence.
func capArticle(lines []string) string {
	var joined strings.Builder
	for _, line := range lines {
		if joined.Len() > 0 && joined.Len()+1+len(line) > maxArticleLength {
			break
		}
		if joined.Len() > 0 {
			joined.WriteString("\n")
		}
		joined.WriteString(line)
	}
	return joined.String()
}
