package main

import (
	"image"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
)

// Popup is a struct with information about the popup.
type Popup struct {
	// The popup window and image.
	win *xwindow.Window
	img *xgraphics.Image

	// The width and height of the popup.
	w, h int

	// The foreground and background colors in hex.
	bg, fg string

	// A channel where the popup should be send to to once its ready
	// to be redrawn.
	redraw chan *Popup
}

func initPopup(x, y, w, h int, bg, fg string) (*Popup, error) {
	popup := new(Popup)
	var err error

	// Create a window for the bar. This window listens to button press events
	// in order to respond to them.
	popup.win, err = xwindow.Generate(X)
	if err != nil {
		return nil, err
	}
	popup.win.Create(X.RootWin(), x, y, w, h, xproto.CwBackPixel|xproto.
		CwEventMask, 0x000000, xproto.EventMaskButtonPress)

	// EWMH stuff.
	// TODO: `WmStateSet` and `WmDesktopSet` are basically here to keep OpenBox
	// happy, can I somehow remove them and just use `_NET_WM_WINDOW_TYPE_DOCK`
	// like I can with WindowChef?
	if err := ewmh.WmWindowTypeSet(X, popup.win.Id, []string{
		"_NET_WM_WINDOW_TYPE_DOCK"}); err != nil {
		return nil, err
	}
	if err := ewmh.WmStateSet(X, popup.win.Id, []string{
		"_NET_WM_STATE_STICKY"}); err != nil {
		return nil, err
	}
	if err := ewmh.WmDesktopSet(X, popup.win.Id, ^uint(0)); err != nil {
		return nil, err
	}
	if err := ewmh.WmNameSet(X, popup.win.Id, "melonbar"); err != nil {
		return nil, err
	}

	// Map window.
	popup.win.Map()

	// TODO: Moving the window is again a hack to keep OpenBox happy.
	popup.win.Move(x, y)

	// Create the bar image.
	popup.img = xgraphics.New(X, image.Rect(0, 0, w, h))
	if err := popup.img.XSurfaceSet(popup.win.Id); err != nil {
		panic(err)
	}
	popup.img.XDraw()

	popup.w = w
	popup.h = h

	popup.bg = bg
	popup.fg = fg

	popup.redraw = make(chan *Popup)

	return popup, nil
}

// TODO: I don't know if this actually frees memory and shit.
func (popup *Popup) destroy() *Popup {
	popup.win.Destroy()
	popup.img.Destroy()
	close(popup.redraw)
	popup = nil

	return popup
}
