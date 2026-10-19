package registry

import (
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "devices.json")
	r, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	must(t, r.Upsert(Device{Name: "smith-tv", Address: "100.64.0.5:5555", Groups: []string{"smith"}}))
	must(t, r.Upsert(Device{Name: "jones-stick", Address: "100.64.0.6:5555"}))
	must(t, r.Save())

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Devices) != 2 || loaded.Devices[0].Name != "jones-stick" {
		t.Fatalf("devices = %+v", loaded.Devices)
	}
}

func TestUpsertValidatesNames(t *testing.T) {
	r := &Registry{}
	if err := r.Upsert(Device{Name: "Bad Name", Address: "x:1"}); err == nil {
		t.Error("expected invalid name error")
	}
	if err := r.Upsert(Device{Name: "ok", Address: "x:1", Groups: []string{"../etc"}}); err == nil {
		t.Error("expected invalid group error")
	}
}

func TestUpsertReplaces(t *testing.T) {
	r := &Registry{}
	must(t, r.Upsert(Device{Name: "tv", Address: "a:1"}))
	must(t, r.Upsert(Device{Name: "tv", Address: "b:2"}))
	if len(r.Devices) != 1 || r.Devices[0].Address != "b:2" {
		t.Fatalf("devices = %+v", r.Devices)
	}
}

func TestSelectRequiresExactlyOne(t *testing.T) {
	r := &Registry{Devices: []Device{
		{Name: "a", Groups: []string{"g"}},
		{Name: "b"},
	}}
	if _, err := r.Select(Selector{}); err == nil {
		t.Error("empty selector should fail")
	}
	if _, err := r.Select(Selector{Names: []string{"a"}, All: true}); err == nil {
		t.Error("names plus --all should fail")
	}
	got, err := r.Select(Selector{Group: "g"})
	if err != nil || len(got) != 1 || got[0].Name != "a" {
		t.Errorf("group select = %+v, %v", got, err)
	}
	if _, err := r.Select(Selector{Names: []string{"missing"}}); err == nil {
		t.Error("unknown name should fail")
	}
}

func TestRemove(t *testing.T) {
	r := &Registry{Devices: []Device{{Name: "a"}, {Name: "b"}}}
	must(t, r.Remove("a"))
	if len(r.Devices) != 1 || r.Devices[0].Name != "b" {
		t.Fatalf("devices = %+v", r.Devices)
	}
	if err := r.Remove("a"); err == nil {
		t.Error("second remove should fail")
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
