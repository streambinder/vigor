package service

import "testing"

func TestPromptRetrievalQuery(t *testing.T) {
	t.Run("summary then articles then raw prompt", func(t *testing.T) {
		got := promptRetrievalQuery("pull-up pyramid", []string{"article body"}, "raw prompt")
		want := "pull-up pyramid\n\narticle body\n\nraw prompt"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("raw prompt keeps retrieval weight without articles", func(t *testing.T) {
		got := promptRetrievalQuery("pull-up pyramid", nil, "raw prompt")
		want := "pull-up pyramid\n\nraw prompt"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("raw prompt stays in when only an article is there", func(t *testing.T) {
		got := promptRetrievalQuery("", []string{"article body"}, "raw prompt")
		if got != "article body\n\nraw prompt" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("degenerate case falls back to raw prompt", func(t *testing.T) {
		if got := promptRetrievalQuery("", nil, "raw prompt"); got != "raw prompt" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("empty input yields empty query", func(t *testing.T) {
		if got := promptRetrievalQuery("", nil, ""); got != "" {
			t.Fatalf("got %q", got)
		}
	})
}
