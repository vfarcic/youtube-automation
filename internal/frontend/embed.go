package frontend

import "embed"

// DistFS holds the built frontend assets.
// During development, this contains a placeholder index.html.
// During production builds, `make frontend-build` copies web/dist here first.
//
//go:embed all:dist
var DistFS embed.FS
