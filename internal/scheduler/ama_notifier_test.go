package scheduler

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

type sendCall struct {
	from           string
	to             []string
	subject        string
	body           string
	attachmentPath string
}

type mockEmailSender struct {
	calls []sendCall
	err   error
}

func (m *mockEmailSender) Send(from string, to []string, subject, body, attachmentPath string) error {
	m.calls = append(m.calls, sendCall{
		from:           from,
		to:             append([]string(nil), to...),
		subject:        subject,
		body:           body,
		attachmentPath: attachmentPath,
	})
	return m.err
}

func TestAMANotifier_Notify(t *testing.T) {
	const (
		from     = "ops@example.com"
		emailTo  = "operator@example.com"
		videoID  = "abc123"
		videoURL = "https://www.youtube.com/watch?v=abc123"
	)
	sentinel := errors.New("upstream boom")

	tests := []struct {
		name string

		result  AMAJobResult
		emailTo string
		sendErr error

		wantSent           bool
		wantErrContains    string
		wantSubject        string
		wantBodyContains   []string
		wantBodyExcluding  []string
	}{
		{
			name: "skipped sends no email",
			result: AMAJobResult{
				Outcome:  OutcomeSkipped,
				VideoID:  videoID,
				VideoURL: videoURL,
			},
			emailTo:  emailTo,
			wantSent: false,
		},
		{
			name: "processed sends success email",
			result: AMAJobResult{
				Outcome:  OutcomeProcessed,
				VideoID:  videoID,
				VideoURL: videoURL,
			},
			emailTo:     emailTo,
			wantSent:    true,
			wantSubject: "AMA processed: " + videoURL,
			wantBodyContains: []string{
				"AMA Processed",
				videoURL,
				`<a href="` + videoURL + `">`,
			},
		},
		{
			name: "failed sends failure email with error and url",
			result: AMAJobResult{
				Outcome:  OutcomeFailed,
				VideoID:  videoID,
				VideoURL: videoURL,
				Err:      sentinel,
			},
			emailTo:     emailTo,
			wantSent:    true,
			wantSubject: "AMA processing failed: upstream boom",
			wantBodyContains: []string{
				"AMA Processing Failed",
				videoURL,
				"upstream boom",
			},
		},
		{
			name: "failed without error details still renders",
			result: AMAJobResult{
				Outcome:  OutcomeFailed,
				VideoID:  videoID,
				VideoURL: videoURL,
				Err:      nil,
			},
			emailTo:     emailTo,
			wantSent:    true,
			wantSubject: "AMA processing failed: (no error details)",
			wantBodyContains: []string{
				"AMA Processing Failed",
				"(no error details)",
			},
		},
		{
			name: "failed without VideoURL falls back to unknown",
			result: AMAJobResult{
				Outcome: OutcomeFailed,
				Err:     sentinel,
			},
			emailTo:     emailTo,
			wantSent:    true,
			wantSubject: "AMA processing failed: upstream boom",
			wantBodyContains: []string{
				"(unknown)",
				"upstream boom",
			},
		},
		{
			name: "scheduler-error sends scheduler error email",
			result: AMAJobResult{
				Outcome: OutcomeSchedulerError,
				Err:     sentinel,
			},
			emailTo:     emailTo,
			wantSent:    true,
			wantSubject: "Scheduler error: upstream boom",
			wantBodyContains: []string{
				"AMA Scheduler Error",
				"upstream boom",
				"did not run",
			},
		},
		{
			name: "empty emailTo is a no-op for sending outcomes",
			result: AMAJobResult{
				Outcome:  OutcomeProcessed,
				VideoID:  videoID,
				VideoURL: videoURL,
			},
			emailTo:  "",
			wantSent: false,
		},
		{
			name: "empty emailTo also no-ops for skipped",
			result: AMAJobResult{
				Outcome: OutcomeSkipped,
			},
			emailTo:  "",
			wantSent: false,
		},
		{
			name: "send failure propagates wrapped",
			result: AMAJobResult{
				Outcome:  OutcomeProcessed,
				VideoID:  videoID,
				VideoURL: videoURL,
			},
			emailTo:         emailTo,
			sendErr:         errors.New("smtp down"),
			wantSent:        true,
			wantErrContains: "send AMA notification: smtp down",
		},
		{
			name: "unknown outcome returns error and does not send",
			result: AMAJobResult{
				Outcome: Outcome("bogus"),
			},
			emailTo:         emailTo,
			wantSent:        false,
			wantErrContains: `unknown outcome "bogus"`,
		},
		{
			name: "html in error message is escaped",
			result: AMAJobResult{
				Outcome: OutcomeFailed,
				Err:     errors.New("<script>alert(1)</script>"),
			},
			emailTo:     emailTo,
			wantSent:    true,
			wantSubject: "AMA processing failed: <script>alert(1)</script>",
			wantBodyContains: []string{
				"&lt;script&gt;alert(1)&lt;/script&gt;",
			},
			wantBodyExcluding: []string{
				"<script>alert(1)</script>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := &mockEmailSender{err: tt.sendErr}
			n := &AMANotifier{Sender: sender, From: from}

			err := n.Notify(tt.result, tt.emailTo)

			if tt.wantErrContains == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrContains)
				}
				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErrContains)
				}
			}

			if tt.wantSent {
				if len(sender.calls) != 1 {
					t.Fatalf("expected 1 send call, got %d", len(sender.calls))
				}
				call := sender.calls[0]
				if call.from != from {
					t.Errorf("from = %q, want %q", call.from, from)
				}
				if len(call.to) != 1 || call.to[0] != tt.emailTo {
					t.Errorf("to = %v, want [%q]", call.to, tt.emailTo)
				}
				if call.attachmentPath != "" {
					t.Errorf("attachmentPath = %q, want empty", call.attachmentPath)
				}
				if tt.wantSubject != "" && call.subject != tt.wantSubject {
					t.Errorf("subject = %q, want %q", call.subject, tt.wantSubject)
				}
				for _, want := range tt.wantBodyContains {
					if !strings.Contains(call.body, want) {
						t.Errorf("body missing %q\nbody:\n%s", want, call.body)
					}
				}
				for _, banned := range tt.wantBodyExcluding {
					if strings.Contains(call.body, banned) {
						t.Errorf("body unexpectedly contains %q\nbody:\n%s", banned, call.body)
					}
				}
			} else {
				if len(sender.calls) != 0 {
					t.Errorf("expected no send calls, got %d", len(sender.calls))
				}
			}
		})
	}
}

func TestSummarizeError(t *testing.T) {
	long := strings.Repeat("a", 200)

	tests := []struct {
		name        string
		err         error
		want        string
		wantNoNL    bool
		wantNoTab   bool
		maxRuneLen  int
		endsInEllip bool
	}{
		{
			name: "nil error returns placeholder",
			err:  nil,
			want: "(no error details)",
		},
		{
			name: "short error is unchanged",
			err:  errors.New("upstream boom"),
			want: "upstream boom",
		},
		{
			name:      "multi-line error is collapsed to single line",
			err:       errors.New("first line\nsecond line\r\nthird line"),
			want:      "first line second line third line",
			wantNoNL:  true,
			wantNoTab: true,
		},
		{
			name:      "tabs and CRs are normalised to spaces",
			err:       errors.New("col1\tcol2\rcol3\t\tcol4"),
			want:      "col1 col2 col3 col4",
			wantNoNL:  true,
			wantNoTab: true,
		},
		{
			name:        "very long error is truncated with ellipsis at 120 runes",
			err:         errors.New(long),
			maxRuneLen:  120,
			endsInEllip: true,
		},
		{
			name: "runs of internal whitespace are collapsed",
			err:  errors.New("a    b\t\t c"),
			want: "a b c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizeError(tt.err)

			if tt.want != "" && got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
			if tt.wantNoNL && (strings.ContainsAny(got, "\n\r")) {
				t.Errorf("got %q contains newline/CR", got)
			}
			if tt.wantNoTab && strings.Contains(got, "\t") {
				t.Errorf("got %q contains tab", got)
			}
			if tt.maxRuneLen > 0 {
				if n := utf8.RuneCountInString(got); n > tt.maxRuneLen {
					t.Errorf("got %d runes, want <= %d (value=%q)", n, tt.maxRuneLen, got)
				}
			}
			if tt.endsInEllip && !strings.HasSuffix(got, "…") {
				t.Errorf("expected truncated value to end with ellipsis, got %q", got)
			}
		})
	}
}
