//go:build windows

package e2e

import (
	"sync/atomic"
	"testing"

	"github.com/egoist/mygo"
	win "github.com/egoist/mygo/internal/windows"
	"github.com/egoist/mygo/ui"
)

// TestZoomControl checks that Ctrl+wheel zooms a page only when
// PageOptions.ZoomControl asks, as WebView2's settings read back, in the
// pages of windows and of web views in native UI.
func TestZoomControl(t *testing.T) {
	zoom := mygo.PageOptions{ZoomControl: true}
	plain := newWindow(t, mygo.WindowOptions{Title: "No zoom control", Width: 300, Height: 200})
	on := newWindow(t, mygo.WindowOptions{Title: "Zoom control", Width: 300, Height: 200, Page: zoom})

	var views atomic.Pointer[[2]*mygo.WebView]
	var frames atomic.Int32
	host := newWindow(t, mygo.WindowOptions{Title: "Web views' zoom control", Width: 400, Height: 300, Content: ui.View(func(c *ui.Context) {
		frames.Add(1)
		ui.Column(c).Fill().AlignItems(ui.Stretch).Children(func() {
			if v := views.Load(); v != nil {
				ui.WebView(c, v[0]).Grow(1)
				ui.WebView(c, v[1]).Grow(1)
			}
		})
	})})
	eventually(t, "a frame", func() bool { return frames.Load() > 0 })
	plainView, err := host.NewWebView(mygo.WebViewOptions{})
	if err != nil {
		t.Fatal(err)
	}
	onView, err := host.NewWebView(mygo.WebViewOptions{Page: zoom})
	if err != nil {
		t.Fatal(err)
	}
	host.Update(func() { views.Store(&[2]*mygo.WebView{plainView, onView}) })

	// The pages load, so that the web views are made: WebView2 makes its
	// own after the window.
	for _, p := range []*mygo.Page{plain.Page(), on.Page(), plainView.Page(), onView.Page()} {
		p.LoadHTML(gesturesPage, "")
		waitForPage(t, p, "window.ready")
	}

	for _, tc := range []struct {
		name    string
		w       *mygo.Window
		webView int
		want    bool
	}{
		{"a window's page by default", plain, -1, false},
		{"a window's page with ZoomControl", on, -1, true},
		{"a web view by default", host, 0, false},
		{"a web view with ZoomControl", host, 1, true},
	} {
		var got, ok bool
		mygo.RunOnMain(func() { got, ok = win.TestZoomControl(tc.w.NativeHandle(), tc.webView) })
		if !ok {
			t.Fatalf("%s has no web view", tc.name)
		}
		if got != tc.want {
			t.Errorf("%s: Ctrl+wheel zooms %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestUserZoomReported zooms a page as the user's Ctrl+wheel does, through
// WebView2's controller behind SetZoomFactor's back: Page.ZoomFactor
// follows, and SetZoomFactor still sets the zoom after it, which the
// change WebView2 reports of it leaves as it is.
func TestUserZoomReported(t *testing.T) {
	w := newWindow(t, mygo.WindowOptions{Title: "User zoom", Width: 300, Height: 200, Page: mygo.PageOptions{ZoomControl: true}})
	w.Page().LoadHTML(gesturesPage, "")
	waitFor(t, w, "window.ready")

	var ok bool
	mygo.RunOnMain(func() { ok = win.TestUserZoom(w.NativeHandle(), 1.5) })
	if !ok {
		t.Fatal("the page has no web view")
	}
	eventually(t, "ZoomFactor to follow the user's zoom", func() bool { return w.Page().ZoomFactor() == 1.5 })

	w.Page().SetZoomFactor(1.25)
	if z := w.Page().ZoomFactor(); z != 1.25 {
		t.Errorf("ZoomFactor after SetZoomFactor(1.25) = %v", z)
	}
	// A script's round trip, after which WebView2 has told of the zoom.
	waitFor(t, w, "window.ready")
	if z := w.Page().ZoomFactor(); z != 1.25 {
		t.Errorf("ZoomFactor once WebView2 told of SetZoomFactor(1.25) = %v", z)
	}
}
