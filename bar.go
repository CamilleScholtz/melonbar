package main

import (
	"fmt"
	"image"
	"os"
	"strconv"
	"sync"

	"github.com/BurntSushi/freetype-go/freetype/truetype"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
)

// Bar is a struct with information about the bar.
type Bar struct {
	// Connection to the X server, the bar window, and the bar image.
	xu  *xgbutil.XUtil
	win *xwindow.Window
	img *xgraphics.Image

	// The font and fontsize that should be used.
	font  *truetype.Font
	fsize float64

	// The width and height of the bar.
	w, h int

	// This is a sum of all of the block widths, used to draw a block
	// to the right of the last block.
	xsum int

	// A map with information about the block, see the `Block` type.
	block *sync.Map

	// A channel where the block should be send to to once its ready
	// to be redrawn.
	redraw chan *Block
}

func initBar(x, y, w, h int, font string, fsize float64) (*Bar,
	error) {
	bar := new(Bar)
	var err error

	// Set up a connection to the X server.
	bar.xu, err = xgbutil.NewConn()
	if err != nil {
		return nil, err
	}

	// Run the main X event loop, this is uses to catch events.
	go xevent.Main(bar.xu)

	// Listen to the root window for property change events, used to
	// check if the user changed the focused window or active
	// workspace for example.
	if err := xwindow.New(bar.xu, bar.xu.RootWin()).Listen(
		xproto.EventMaskPropertyChange); err != nil {
		return nil, err
	}

	// Create a window for the bar. This window listens to button
	// press events in order to respond to them.
	bar.win, err = xwindow.Generate(bar.xu)
	if err != nil {
		return nil, err
	}
	bar.win.Create(bar.xu.RootWin(), x, y, w, h, xproto.CwBackPixel|
		xproto.CwEventMask, 0x000000, xproto.EventMaskButtonPress)

	// EWMH stuff to make the window behave like an actuale bar.
	// TODO: `WmStateSet` and `WmDesktopSet` are basically here to
	// keep OpenBox happy, can I somehow remove them and just use
	// `_NET_WM_WINDOW_TYPE_DOCK` like I can with WindowChef?
	if err := ewmh.WmWindowTypeSet(bar.xu, bar.win.Id, []string{
		"_NET_WM_WINDOW_TYPE_DOCK"}); err != nil {
		return nil, err
	}
	if err := ewmh.WmStateSet(bar.xu, bar.win.Id, []string{
		"_NET_WM_STATE_STICKY"}); err != nil {
		return nil, err
	}
	if err := ewmh.WmDesktopSet(bar.xu, bar.win.Id, ^uint(0)); err !=
		nil {
		return nil, err
	}
	if err := ewmh.WmNameSet(bar.xu, bar.win.Id, "melonbar"); err !=
		nil {
		return nil, err
	}

	// Map window.
	bar.win.Map()

	// TODO: Moving the window is again a hack to keep OpenBox happy.
	bar.win.Move(x, y)

	// Create the bar image.
	bar.img = xgraphics.New(bar.xu, image.Rect(0, 0, w, h))
	bar.img.XSurfaceSet(bar.win.Id)
	bar.img.XDraw()

	// Load font.
	// TODO: I don't *really* want to use `ttf` fonts but there
	// doesn't seem to be any `pcf` Go library at the moment.
	f, err := os.Open(font)
	if err != nil {
		return nil, err
	}
	bar.font, err = xgraphics.ParseFont(f)
	if err != nil {
		return nil, err
	}
	bar.fsize = fsize

	bar.w = w
	bar.h = h

	bar.block = new(sync.Map)
	bar.redraw = make(chan *Block)

	// Listen to mouse events and execute the requires function.
	xevent.ButtonPressFun(func(_ *xgbutil.XUtil,
		ev xevent.ButtonPressEvent) {
		// Determine what block the cursor is in.
		// TODO: This feels a bit slow at the moment, can I improve
		// it?
		var block *Block
		bar.block.Range(func(val, i interface{}) bool {
			block = i.(*Block)
			if ev.EventX > int16(block.x) && ev.EventX < int16(
				block.x+block.w) {
				return false
			}
			return true
		})

		// Execute the function as specified.
		block.actions["button"+strconv.Itoa(int(ev.Detail))]()
	}).Connect(bar.xu, bar.win.Id)

	return bar, nil
}

func (bar *Bar) draw(block *Block) error {
	// Calculate the required x coordinate for the different
	// aligments.
	tw, th := xgraphics.Extents(bar.font, bar.fsize, block.txt)
	var x int
	switch block.align {
	case 'l':
		x = block.x + block.xoff
	case 'c':
		x = block.x + ((block.w / 2) - (tw / 2)) + block.xoff
	case 'r':
		x = (block.x + block.w) - tw + block.xoff
	default:
		return fmt.Errorf("draw %#U: Not a valid aligment rune",
			block.align)
	}

	// Color the backround.
	block.img.For(func(cx, cy int) xgraphics.BGRA {
		// Hack for music block background.
		// TODO: I should handle this in `initBlock()`.
		if block.w == 660 {
			if cx < x+block.xoff {
				return hexToBGRA("#445967")
			}
			return hexToBGRA(block.bg)
		}

		return hexToBGRA(block.bg)
	})

	// Draw the text.
	// TODO: Center text vertically a bit more precisely.
	if _, _, err := block.img.Text(x, (bar.h/2)-(th/2)-2, hexToBGRA(
		block.fg), bar.fsize, bar.font, block.txt); err != nil {
		return err
	}

	block.img.XDraw()
	bar.img.XPaint(bar.win.Id)
	return nil
}
