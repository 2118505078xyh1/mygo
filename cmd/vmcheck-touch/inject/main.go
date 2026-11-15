//go:build windows

// Inject drives the "touch check" window with synthetic touches through
// InjectTouchInput: a four-finger swipe, a five-finger tap, and a
// one-finger tap (#1884 check).
package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

type pointerInfo struct {
	Type, ID, Frame, Flags                 uint32
	Source, Hwnd                           uintptr
	Pixel, Himetric, PixelRaw, HimetricRaw [2]int32
	Time, History                          uint32
	InputData                              int32
	KeyStates                              uint32
	PerfCount                              uint64
	ButtonChange                           int32
	_                                      int32
}

type touchInfo struct {
	P                     pointerInfo
	TouchFlags, TouchMask uint32
	Contact, ContactRaw   [4]int32
	Orientation, Pressure uint32
}

const (
	ptTouch       = 2
	flagInRange   = 0x2
	flagInContact = 0x4
	flagDown      = 0x10000
	flagUpdate    = 0x20000
	flagUp        = 0x40000
)

var (
	user32     = syscall.NewLazyDLL("user32.dll")
	initInject = user32.NewProc("InitializeTouchInjection")
	inject     = user32.NewProc("InjectTouchInput")
	findWindow = user32.NewProc("FindWindowW")
	getRect    = user32.NewProc("GetWindowRect")
	setFg      = user32.NewProc("SetForegroundWindow")
	keybdEvent = user32.NewProc("keybd_event")
)

func send(pts [][2]int32, flags uint32) {
	infos := make([]touchInfo, len(pts))
	for i, p := range pts {
		t := &infos[i]
		t.P.Type, t.P.ID, t.P.Flags = ptTouch, uint32(i), flags
		t.P.Pixel = p
		t.TouchMask = 1 | 2 | 4
		t.Contact = [4]int32{p[0] - 2, p[1] - 2, p[0] + 2, p[1] + 2}
		t.Orientation, t.Pressure = 90, 32000
	}
	r, _, err := inject.Call(uintptr(len(infos)), uintptr(unsafe.Pointer(&infos[0])))
	if r == 0 {
		fmt.Println("InjectTouchInput failed:", err)
	}
	time.Sleep(20 * time.Millisecond)
}

func gesture(name string, pts [][2]int32, dx int32, steps int) {
	fmt.Println(time.Now().Format("15:04:05.000"), "INJECT", name, len(pts), "fingers")
	send(pts, flagDown|flagInRange|flagInContact)
	for s := 0; s < steps; s++ {
		for i := range pts {
			pts[i][0] += dx
		}
		send(pts, flagUpdate|flagInRange|flagInContact)
	}
	send(pts, flagUp)
	time.Sleep(700 * time.Millisecond)
}

func main() {
	title, _ := syscall.UTF16PtrFromString("touch check")
	hwnd, _, _ := findWindow.Call(0, uintptr(unsafe.Pointer(title)))
	if hwnd == 0 {
		fmt.Println("no window titled touch check")
		os.Exit(1)
	}
	keybdEvent.Call(0x87, 0, 0, 0)
	keybdEvent.Call(0x87, 0, 2, 0)
	setFg.Call(hwnd)
	time.Sleep(500 * time.Millisecond)
	var rc [4]int32
	getRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	cx, cy := (rc[0]+rc[2])/2, (rc[1]+rc[3])/2
	fmt.Println("window rect", rc, "center", cx, cy)
	if r, _, err := initInject.Call(10, 1); r == 0 {
		fmt.Println("InitializeTouchInjection failed:", err)
		os.Exit(1)
	}
	four := [][2]int32{}
	for i := int32(0); i < 4; i++ {
		four = append(four, [2]int32{cx - 200 + i*40, cy})
	}
	gesture("four-finger swipe", four, 20, 10)
	five := [][2]int32{}
	for i := int32(0); i < 5; i++ {
		five = append(five, [2]int32{cx - 100 + i*40, cy + 60})
	}
	gesture("five-finger tap", five, 0, 0)
	gesture("one-finger tap", [][2]int32{{cx, cy - 80}}, 0, 0)
	fmt.Println(time.Now().Format("15:04:05.000"), "INJECT done")
}
