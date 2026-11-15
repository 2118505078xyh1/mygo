package ui

import (
	"fmt"
	"runtime"
	"slices"
	"testing"
	"weak"

	"github.com/egoist/mygo/internal/text"
	"github.com/egoist/mygo/internal/wait"
	"golang.org/x/image/font/gofont/gomono"
)

// TestTwoTestersInParallel runs Testers on different goroutines, as
// t.Parallel does, and fails under go test -race while they share one
// glyph atlas (issue #1722).
func TestTwoTestersInParallel(t *testing.T) {
	for i := range 4 {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			frame := 0
			tt := NewTester(func(c *Context) {
				Column(c).Children(func() {
					for j := range 30 {
						r := rune(0x4e00 + i*3000 + frame*30 + j) // new glyphs every frame
						Row(c).Children(func() {
							Text(c, fmt.Sprintf("tester %d frame %d %c", i, frame, r)).FontSize(float32(11 + (frame+j)%9))
						})
					}
				})
			}, 400, 900)
			for range 40 {
				frame++
				tt.Frame()
			}
		})
	}
}

// A Tester's window, which shapes text in a system of its own, draws the
// fonts the app registered, as the app's windows do. Go Mono is
// monospaced, so "iiiiiiii" is as wide as "mmmmmmmm" only in it.
func TestATesterDrawsRegisteredFonts(t *testing.T) {
	const family = "MyGo Test Tester Mono"
	if err := RegisterFont(gomono.TTF, family); err != nil {
		t.Fatal(err)
	}
	tt := NewTester(func(c *Context) {
		Column(c).Children(func() {
			for _, s := range []string{"iiiiiiii", "mmmmmmmm"} {
				Row(c).Children(func() { Text(c, s).Font(family).FontSize(20) })
			}
		})
	}, 400, 300)
	i, _ := tt.Find("iiiiiiii")
	m, _ := tt.Find("mmmmmmmm")
	if i.W < m.W-0.5 || i.W > m.W+0.5 {
		t.Fatalf("the registered font isn't drawn: %v wide, %v in Go Mono", i.W, m.W)
	}
}

// The text system of a Tester's window, its own, becomes unreachable with
// the Tester, which closes its engine: a test binary of a thousand Testers
// otherwise kept a thousand, each with a Pango font map and a thread of
// fontconfig's on Linux (#1902).
func TestTesterTextSystemUnreachable(t *testing.T) {
	var systems []weak.Pointer[text.System]
	for i := range 8 {
		tt := NewTester(func(c *Context) {
			Column(c).Children(func() {
				Text(c, fmt.Sprintf("tester %d", i))
				TextInput(c, new(string)).Placeholder("Name")
			})
		}, 300, 200)
		tt.Frame()
		if tt.h.text == nil {
			t.Fatal("the Tester's window has no text system of its own")
		}
		systems = append(systems, weak.Make(tt.h.text))
	}
	wait.Until(t, "the text systems of the Testers dropped unreachable", func() bool {
		runtime.GC()
		return !slices.ContainsFunc(systems, func(w weak.Pointer[text.System]) bool { return w.Value() != nil })
	})
}
