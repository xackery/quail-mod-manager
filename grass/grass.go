package grass

import "embed"

var (
	//go:embed assets/*.dds
	Assets embed.FS
)
