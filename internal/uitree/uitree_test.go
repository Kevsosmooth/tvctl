package uitree

import "testing"

const sample = `<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<hierarchy rotation="0">
  <node index="0" text="" resource-id="" class="android.widget.FrameLayout" package="com.android.tv.settings" content-desc="" checkable="false" checked="false" clickable="false" focusable="false" focused="false" selected="false" bounds="[0,0][1920,1080]">
    <node index="0" text="Device Preferences" resource-id="android:id/title" class="android.widget.TextView" package="com.android.tv.settings" content-desc="" checkable="false" checked="false" clickable="false" focusable="false" focused="false" selected="false" bounds="[100,200][500,260]" />
    <node index="1" text="" resource-id="com.android.tv.settings:id/row" class="android.widget.LinearLayout" package="com.android.tv.settings" content-desc="Developer options" checkable="false" checked="false" clickable="true" focusable="true" focused="true" selected="false" bounds="[100,300][900,360]" />
    <node index="2" text="USB debugging" resource-id="android:id/switch_widget" class="android.widget.Switch" package="com.android.tv.settings" content-desc="" checkable="true" checked="true" clickable="true" focusable="true" focused="false" selected="false" bounds="[100,400][900,460]" />
  </node>
</hierarchy>`

func TestParseKeepsMeaningfulNodes(t *testing.T) {
	els, err := Parse([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	if len(els) != 3 {
		t.Fatalf("got %d elements, want 3: %+v", len(els), els)
	}
	if els[0].Class != "TextView" || els[0].Bounds != (Rect{100, 200, 500, 260}) {
		t.Errorf("first = %+v", els[0])
	}
	if !els[1].Focused || els[1].Index != 1 {
		t.Errorf("second = %+v", els[1])
	}
	if els[2].Checked == nil || !*els[2].Checked {
		t.Errorf("switch checked = %v", els[2].Checked)
	}
	if x, y := els[1].Bounds.Center(); x != 500 || y != 330 {
		t.Errorf("center = %d,%d", x, y)
	}
}

func TestFind(t *testing.T) {
	els, _ := Parse([]byte(sample))
	cases := []struct {
		q    Query
		want int
	}{
		{Query{Text: "developer options"}, 1},
		{Query{Text: "debugging"}, 2},
		{Query{ID: "switch_widget"}, 2},
		{Query{Text: "device preferences", ID: "android:id/title"}, 0},
	}
	for _, c := range cases {
		got, err := Find(els, c.q)
		if err != nil || got.Index != c.want {
			t.Errorf("Find(%+v) = %d, %v; want %d", c.q, got.Index, err, c.want)
		}
	}
	if _, err := Find(els, Query{Text: "debugging", Exact: true}); err == nil {
		t.Error("exact query should not match a substring")
	}
	if _, err := Find(els, Query{Text: "Netflix"}); err == nil {
		t.Error("missing text should fail")
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	if _, err := Parse([]byte("ERROR: could not get idle state")); err == nil {
		t.Error("expected parse error")
	}
}

const rows = `<hierarchy rotation="0">
  <node text="" class="android.widget.LinearLayout" clickable="true" focusable="true" focused="true" bounds="[0,0][800,100]">
    <node text="Network &amp; Internet" class="android.widget.TextView" bounds="[10,10][400,50]" />
    <node text="La_Yema" class="android.widget.TextView" bounds="[10,50][400,90]" />
  </node>
  <node text="Reboot" class="android.widget.TextView" bounds="[0,100][800,200]" />
</hierarchy>`

func TestRowsTakeChildLabels(t *testing.T) {
	els, err := Parse([]byte(rows))
	if err != nil {
		t.Fatal(err)
	}
	if len(els) != 2 {
		t.Fatalf("got %d elements: %+v", len(els), els)
	}
	if els[0].Text != "Network & Internet · La_Yema" || !els[0].Focused {
		t.Errorf("row = %+v", els[0])
	}
	if els[1].Text != "Reboot" || els[1].Index != 1 {
		t.Errorf("plain text = %+v", els[1])
	}
	got, err := Find(els, Query{Text: "network"})
	if err != nil || got.Index != 0 {
		t.Errorf("Find network = %+v, %v", got, err)
	}
}
