//go:build windows

package main

import "unsafe"

var (
	procGetWindowPlacement = user32.NewProc("GetWindowPlacement")
	procSetWindowPlacement = user32.NewProc("SetWindowPlacement")
	procIsWindowVisible    = user32.NewProc("IsWindowVisible")
	procMonitorFromPoint   = user32.NewProc("MonitorFromPoint")
)

const (
	swHide               = 0
	swShowNormal         = 1
	swShowMinimized      = 2
	swShowMaximized      = 3
	wpfRestoreToMaximize = 0x0002
	monitorDefaultToNull = 0

	// Parte della barra del titolo che deve restare su uno schermo, così la
	// finestra si può sempre afferrare e spostare.
	titleBarProbe = 15
)

type windowPlacementInfo struct {
	length, flags, showCmd uint32
	ptMinPosition          [2]int32
	ptMaxPosition          [2]int32
	rcNormalPosition       rect
}

// windowPlacement legge le dimensioni "normali" della finestra e se è
// ingrandita (anche se in questo momento è ridotta a icona).
func windowPlacement() (WindowState, bool) {
	hwnd := findMainWindow()
	if hwnd == 0 {
		return WindowState{}, false
	}
	wp := windowPlacementInfo{length: uint32(unsafe.Sizeof(windowPlacementInfo{}))}
	if ok, _, _ := procGetWindowPlacement.Call(hwnd, uintptr(unsafe.Pointer(&wp))); ok == 0 {
		return WindowState{}, false
	}
	r := wp.rcNormalPosition
	return WindowState{
		X:      r.left,
		Y:      r.top,
		Width:  r.right - r.left,
		Height: r.bottom - r.top,
		Maximised: wp.showCmd == swShowMaximized ||
			(wp.showCmd == swShowMinimized && wp.flags&wpfRestoreToMaximize != 0),
	}, true
}

// restoreWindowPlacement applica le dimensioni salvate. Se la barra del titolo
// non cade su nessuno schermo (es. monitor scollegato) non fa nulla e la
// finestra resta centrata come da default.
func restoreWindowPlacement(w WindowState) {
	if !titleBarOnScreen(w) {
		return
	}
	hwnd := findMainWindow()
	if hwnd == 0 {
		return
	}
	wp := windowPlacementInfo{
		length: uint32(unsafe.Sizeof(windowPlacementInfo{})),
		rcNormalPosition: rect{
			left:   w.X,
			top:    w.Y,
			right:  w.X + w.Width,
			bottom: w.Y + w.Height,
		},
	}
	// La finestra di solito è ancora nascosta (Wails la mostra a pagina
	// caricata): in quel caso va solo posizionata, senza mostrarla.
	wp.showCmd = swHide
	if visible, _, _ := procIsWindowVisible.Call(hwnd); visible != 0 {
		wp.showCmd = swShowNormal
		if w.Maximised {
			wp.showCmd = swShowMaximized
		}
	}
	procSetWindowPlacement.Call(hwnd, uintptr(unsafe.Pointer(&wp)))
}

// titleBarOnScreen controlla che il centro della barra del titolo sia su uno
// schermo. Le coordinate salvate sono relative all'area di lavoro (senza
// barra delle applicazioni): lo scarto è trascurabile per questo controllo.
func titleBarOnScreen(w WindowState) bool {
	x := w.X + w.Width/2
	y := w.Y + titleBarProbe
	point := uintptr(uint32(x)) | uintptr(uint32(y))<<32
	monitor, _, _ := procMonitorFromPoint.Call(point, monitorDefaultToNull)
	return monitor != 0
}
