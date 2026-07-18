package mail

import "testing"

func TestNormalizedThreadSubjectUsesPlaceholderWhenPrefixesConsumeSubject(t *testing.T) {
	for _, subject := range []string{"", "   ", "Re:", "Fwd: Re:"} {
		if got := normalizedThreadSubject(subject); got != "(no subject)" {
			t.Fatalf("normalizedThreadSubject(%q) = %q, want %q", subject, got, "(no subject)")
		}
	}
}
