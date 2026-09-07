package api

import (
	"log/slog"
	"net/http"
)

// syncBeforeAction refreshes the data repo before handlers that act on video
// state.
//
// Read handlers already pull via PullIfStale, but the publish, action and
// social routes went straight to disk. That let them operate on a clone that
// had drifted behind the data repo — publishing a video from metadata that had
// since been corrected upstream, and reporting the stale values as fact. The
// symptom that surfaced this: Hugo publishing kept failing with "Video has no
// title" for a video whose titles had already been fixed and pushed.
//
// Failures are logged, not fatal. A publish blocked on a transient network
// error would be worse than one running against a slightly old clone, and the
// handler still validates whatever it reads.
func (s *Server) syncBeforeAction(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.gitSync != nil {
			if err := s.gitSync.PullIfStale(pullOnReadThrottle); err != nil {
				slog.Warn("git: pull before action failed, continuing with local copy",
					"path", r.URL.Path, "err", err)
			}
		}
		next.ServeHTTP(w, r)
	})
}
