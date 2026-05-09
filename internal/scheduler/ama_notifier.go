package scheduler

import (
	"fmt"
	"html"
	"log/slog"
	"strings"
	"unicode/utf8"
)

// EmailSender abstracts the email transport used by AMANotifier so tests can
// inject a fake. The signature mirrors notification.Email.Send so that
// *notification.Email satisfies it directly when wired in production.
type EmailSender interface {
	Send(from string, to []string, subject, body, attachmentPath string) error
}

// AMANotifier turns AMAJobResult values into operator emails per the
// notification rules in PRD #386: skipped runs are silent; processed,
// failed, and scheduler-error runs each get a distinctly-worded email.
type AMANotifier struct {
	Sender EmailSender
	From   string
}

// Notify sends one email describing the outcome of an AMA job run.
//
// Behavior:
//   - OutcomeSkipped: returns nil without calling the sender.
//   - emailTo == "": logs a warning and returns nil. Treating an empty
//     recipient as "no operator configured" rather than an error keeps the
//     scheduler running when notifications are unconfigured; startup-time
//     config validation (Milestone 5) is the right place to fail loud.
//   - Other outcomes: builds a subject/body and forwards to Sender.Send.
//     Send errors propagate so the caller can log/observe them; the
//     underlying AMA work has already completed by the time Notify runs.
func (n *AMANotifier) Notify(result AMAJobResult, emailTo string) error {
	if result.Outcome == OutcomeSkipped {
		return nil
	}
	if emailTo == "" {
		slog.Warn("AMA notifier: emailTo is empty, skipping notification",
			"outcome", string(result.Outcome),
			"videoID", result.VideoID,
		)
		return nil
	}

	subject, body, ok := buildAMAEmail(result)
	if !ok {
		return fmt.Errorf("AMA notifier: unknown outcome %q", result.Outcome)
	}

	if err := n.Sender.Send(n.From, []string{emailTo}, subject, body, ""); err != nil {
		return fmt.Errorf("send AMA notification: %w", err)
	}
	return nil
}

func buildAMAEmail(result AMAJobResult) (subject, body string, ok bool) {
	switch result.Outcome {
	case OutcomeProcessed:
		url := result.VideoURL
		if url == "" {
			url = "(unknown)"
		}
		safeURL := html.EscapeString(url)
		subject = fmt.Sprintf("AMA processed: %s", url)
		body = fmt.Sprintf(`<strong>AMA Processed</strong>
<br/><br/>
<ul>
<li><strong>Video:</strong> <a href="%s">%s</a></li>
</ul>
`, safeURL, safeURL)
		return subject, body, true

	case OutcomeFailed:
		url := result.VideoURL
		if url == "" {
			url = "(unknown)"
		}
		safeURL := html.EscapeString(url)
		subject = fmt.Sprintf("AMA processing failed: %s", summarizeError(result.Err))
		body = fmt.Sprintf(`<strong>AMA Processing Failed</strong>
<br/><br/>
<ul>
<li><strong>Video:</strong> <a href="%s">%s</a></li>
<li><strong>Error:</strong></li>
</ul>
<pre>%s</pre>
`, safeURL, safeURL, html.EscapeString(errString(result.Err)))
		return subject, body, true

	case OutcomeSchedulerError:
		subject = fmt.Sprintf("Scheduler error: %s", summarizeError(result.Err))
		body = fmt.Sprintf(`<strong>AMA Scheduler Error</strong>
<br/><br/>
The AMA scheduler failed before deciding whether to process a video. The job did not run.
<br/><br/>
<strong>Error:</strong>
<pre>%s</pre>
`, html.EscapeString(errString(result.Err)))
		return subject, body, true

	default:
		return "", "", false
	}
}

func errString(err error) string {
	if err == nil {
		return "(no error details)"
	}
	return err.Error()
}

// summarizeError returns a single-line, length-capped error summary that is
// safe to embed in an email Subject header. It collapses internal whitespace
// (so multi-line errors don't break header formatting) and truncates anything
// longer than 120 runes with an ellipsis. The full pre-formatted error is
// still rendered verbatim in the email body.
func summarizeError(err error) string {
	if err == nil {
		return "(no error details)"
	}
	s := err.Error()
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return ' '
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(s), " ")

	const maxLen = 120
	if utf8.RuneCountInString(s) > maxLen {
		runes := []rune(s)
		s = string(runes[:maxLen-1]) + "…"
	}
	return s
}
