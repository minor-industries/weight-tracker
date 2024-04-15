package assets

import "embed"

//go:embed *.html purecss/*.css
var FS embed.FS
