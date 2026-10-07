package mcp

import (
	"strings"
	"testing"
)

const requestEventsBody = `[
	{"ts":"2026-10-07T10:00:00.000Z","kind":"browser","ctx":{"type":"browser","site":"acme","rid":"r1"},"data":{"type":"navigation"}},
	{"ts":"2026-10-07T10:00:00.010Z","kind":"query","ctx":{"type":"fpm","site":"acme","request":"GET /cart","rid":"r1"},"src":{"file":"/app/Cart.php","line":12},"data":{"sql":"select 1","time_ms":2,"trace":[{"file":"x"}]}},
	{"ts":"2026-10-07T10:00:00.020Z","kind":"query","ctx":{"type":"fpm","site":"acme","rid":"r1"},"data":{"sql":"select 2"}},
	{"ts":"2026-10-07T10:00:00.030Z","kind":"log","ctx":{"type":"fpm","site":"acme","rid":"r1"},"data":{"level":"error"}}
]`

// The lens bar an assistant reads names the request and counts each lens the
// way the dashboard does, leaving the page view out.
func TestRequestTool_LensesCountsEachLens(t *testing.T) {
	path := stubRoundTrip(t, requestEventsBody)
	got, _ := execRequestTool(map[string]any{"action": "lenses", "rid": "r1"})
	if *path != "/api/dumps?rid=r1" {
		t.Errorf("path = %q", *path)
	}
	text := toolText(got)
	for _, want := range []string{`"queries":2`, `"logs":1`, `"request":"GET /cart"`} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %s in %s", want, text)
		}
	}
	if strings.Contains(text, `"browser"`) {
		t.Errorf("a page view is not a browser event: %s", text)
	}
}

// One lens reads a page at a time, on the request's clock and without traces.
func TestRequestTool_LensPagesItsRows(t *testing.T) {
	path := stubRoundTrip(t, requestEventsBody)
	got, _ := execRequestTool(map[string]any{"action": "lens", "rid": "r1", "lens": "queries", "offset": 1, "limit": 1})
	if !strings.Contains(*path, "kind=query") || !strings.Contains(*path, "rid=r1") {
		t.Errorf("path = %q", *path)
	}
	text := toolText(got)
	if !strings.Contains(text, `"total":4`) || !strings.Contains(text, `"n":2`) || !strings.Contains(text, `"offset_ms":10`) {
		t.Errorf("page = %s", text)
	}
	if strings.Contains(text, "trace") || strings.Contains(text, `"n":3`) {
		t.Errorf("page must drop traces and stop at the limit: %s", text)
	}
}

func TestRequestTool_NeedsARidAndALens(t *testing.T) {
	stubRoundTrip(t, `[]`)
	if got, _ := execRequestTool(map[string]any{"action": "lenses"}); !strings.Contains(toolText(got), "rid is required") {
		t.Errorf("no rid = %s", toolText(got))
	}
	if got, _ := execRequestTool(map[string]any{"action": "lens", "rid": "r1", "lens": "nope"}); !strings.Contains(toolText(got), "lens is required") {
		t.Errorf("unknown lens = %s", toolText(got))
	}
}
