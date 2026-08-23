package service

import "testing"

func TestFreeTextRetrievalQuery(t *testing.T) {
	t.Run("summary then articles then raw request as fallback", func(t *testing.T) {
		got := freeTextRetrievalQuery("pull-up pyramid", []string{"article body"}, "raw request")
		want := "pull-up pyramid\n\narticle body"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("raw request carries weight only without articles", func(t *testing.T) {
		got := freeTextRetrievalQuery("pull-up pyramid", nil, "raw request")
		want := "pull-up pyramid\n\nraw request"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("raw request stays out when an article is there", func(t *testing.T) {
		got := freeTextRetrievalQuery("", []string{"article body"}, "raw request")
		if got != "article body" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("degenerate case falls back to raw request", func(t *testing.T) {
		if got := freeTextRetrievalQuery("", nil, "raw request"); got != "raw request" {
			t.Fatalf("got %q", got)
		}
	})
}
