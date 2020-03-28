package main

import (
	"image"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
	"golang.org/x/image/font"
)

// Popup is a struct with information about the popup.
type Popup struct {
	// The popup window and image.
	win *xwindow.Window
	img *xgraphics.Image

	// The position, width and height of the popup.
	x, y, w, h int

	// Text drawer.
	drawer *font.Drawer

	// If the popup is currently open or not.
	open bool

	// The fuction that updates the block, this will be executes as a goroutine.
	update func()
}

func (bar *Bar) drawPopup(key string) error {
	popup := bar.popup(key)

	// If the popup is already open, we destroy it.
	if popup.open {
		popup.destroy()
		return nil
	}

	// Create a window for the popup. This window listens to button press
	// events in order to respond to them.
	var err error
	popup.win, err = xwindow.Generate(X)
	if err != nil {
		return err
	}
	popup.win.Create(X.RootWin(), popup.x, popup.y, popup.w, popup.h, xproto.
		CwBackPixel|xproto.CwEventMask, 0x000000, xproto.EventMaskButtonPress)

	// EWMH stuff.
	if err := initEWMH(popup.win.Id); err != nil {
		return err
	}

	// Map window.
	popup.win.Map()

	// XXX: Moving the window is again a hack to keep OpenBox happy.
	popup.win.Move(popup.x, popup.y)

	// Create the popup image.
	popup.img = xgraphics.New(X, image.Rect(0, 0, popup.w, popup.h))
	if err := popup.img.XSurfaceSet(popup.win.Id); err != nil {
		panic(err)
	}
	popup.img.XDraw()

	// Set popup font face.
	popup.drawer = &font.Drawer{
		Dst:  popup.img,
		Face: face,
	}

	// Run update function.
	popup.update()

	return nil
}

func (bar *Bar) popup(key string) *Popup {
	i, _ := bar.popups.Get(key)
	return i.(*Popup)
}

/*func initPopup(x, y, w, h int, bg, fg string) (*Popup, error) {
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
	if err := initEWMH(popup.win.Id); err != nil {
		return nil, err
	}

	// Map window.
	popup.win.Map()

	// XXX: Moving the window is again a hack to keep OpenBox happy.
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

	popup.drawer = &font.Drawer{
		Dst:  popup.img,
		Face: face,
	}

	// Color the background.
	popup.img.For(func(cx, cy int) xgraphics.BGRA {
		return popup.bg
	})

	// Draw the popup.
	popup.draw()

	return popup, nil
}*/

func (popup *Popup) draw() {
	popup.img.XDraw()
	popup.img.XPaint(popup.win.Id)

	// Set popup status to open.
	popup.open = true
}

// TODO: I don't know if this actually frees memory and shit.
func (popup *Popup) destroy() {
	popup.win.Destroy()
	popup.img.Destroy()

	// Set popup status to closed.
	popup.open = false
}
