package splithttp_test

import (
	"testing"

	. "github.com/xtls/xray-core/transport/internet/splithttp"
)

func Test_GetNormalizedPath(t *testing.T) {
	c := Config{
		Path: "/?world",
	}

	path := c.GetNormalizedPath()
	if path != "/" {
		t.Error("Unexpected: ", path)
	}
}

func Test_GetNormalizedPath_NoTrailingSlashWhenMetaOffPath(t *testing.T) {
	// Session/seq off the path (query) -> path must be sent verbatim, so it can
	// point at a real static file behind CDNs that 403 on a trailing slash.
	c := Config{
		Path:               "/static/v1/collect.js",
		SessionIDPlacement: PlacementQuery,
		SeqPlacement:       PlacementQuery,
	}

	path := c.GetNormalizedPath()
	if path != "/static/v1/collect.js" {
		t.Error("Unexpected: ", path)
	}
}

func Test_GetNormalizedPath_TrailingSlashWhenMetaOnPath(t *testing.T) {
	// Default (path) placement needs the trailing slash as a segment separator.
	c := Config{
		Path: "/static/v1/collect.js",
	}

	path := c.GetNormalizedPath()
	if path != "/static/v1/collect.js/" {
		t.Error("Unexpected: ", path)
	}
}
