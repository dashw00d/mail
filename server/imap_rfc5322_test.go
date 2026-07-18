package mail

import "testing"

func TestParseRFC5322LegacyCharsets(t *testing.T) {
	tests := []struct {
		name    string
		charset string
		body    string
		want    string
	}{
		{
			name:    "windows-1252",
			charset: "windows-1252",
			body:    "Price: \x8099\r\n",
			want:    "Price: \u20ac99\r\n",
		},
		{
			name:    "iso-8859-1",
			charset: "iso-8859-1",
			body:    "Caf\xe9\r\n",
			want:    "Caf\u00e9\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := "From: sender@example.com\r\n" +
				"To: recipient@example.com\r\n" +
				"Message-ID: <legacy-" + tt.name + "@example.com>\r\n" +
				"Content-Type: text/plain; charset=" + tt.charset + "\r\n" +
				"Content-Transfer-Encoding: 8bit\r\n\r\n" + tt.body

			message, err := parseRFC5322([]byte(raw))
			if err != nil {
				t.Fatalf("parseRFC5322 returned an error: %v", err)
			}
			if message.TextBody != tt.want {
				t.Fatalf("TextBody = %q, want %q", message.TextBody, tt.want)
			}
		})
	}
}

func TestParseRFC5322UsesPlaceholderForBlankSubject(t *testing.T) {
	raw := "From: sender@example.com\r\n" +
		"To: recipient@example.com\r\n" +
		"Message-ID: <no-subject@example.com>\r\n\r\n" +
		"Message without a subject.\r\n"

	message, err := parseRFC5322([]byte(raw))
	if err != nil {
		t.Fatalf("parseRFC5322 returned an error: %v", err)
	}
	if message.Subject != "(no subject)" {
		t.Fatalf("Subject = %q, want %q", message.Subject, "(no subject)")
	}
}
