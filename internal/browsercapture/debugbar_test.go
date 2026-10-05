package browsercapture

import (
	"testing"

	"github.com/geodro/lerd/internal/config"
)

func TestSetDebugbar_RewritesTheSiteVhostOnChangeOnly(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	site := config.Site{Name: "shop", Domains: []string{"shop.test"}, Path: t.TempDir()}
	if err := config.AddSite(site); err != nil {
		t.Fatal(err)
	}
	var rewritten []string
	reloads := 0
	origRegen, origReload := regenerateSiteVhostFn, nginxReloadFn
	regenerateSiteVhostFn = func(s config.Site) error {
		if !s.Debugbar {
			t.Errorf("vhost rewritten from the old registry entry")
		}
		rewritten = append(rewritten, s.Name)
		return nil
	}
	nginxReloadFn = func() error { reloads++; return nil }
	t.Cleanup(func() { regenerateSiteVhostFn, nginxReloadFn = origRegen, origReload })

	if res, err := SetDebugbar(site, true); err != nil || res.NoChange {
		t.Fatalf("SetDebugbar = %+v, %v", res, err)
	}
	updated, _ := config.FindSite("shop")
	if res, _ := SetDebugbar(*updated, true); !res.NoChange {
		t.Fatalf("second SetDebugbar = %+v, want no change", res)
	}
	if len(rewritten) != 1 || reloads != 1 {
		t.Fatalf("rewritten %v with %d reloads, want [shop] once", rewritten, reloads)
	}
}
