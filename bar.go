package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"path"
	"strconv"
	"sync"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
	"golang.org/x/image/font"
	"golang.org/x/image/font/plan9font"
	"golang.org/x/image/math/fixed"
)

// Bar is a struct with information about the bar.
type Bar struct {
	// Connection to the X server, the bar window, and the bar image.
	xu  *xgbutil.XUtil
	win *xwindow.Window
	img *xgraphics.Image

	// The width and height of the bar.
	w, h int

	// This is a sum of all of the block widths, used to draw a block to the
	// right of the last block.
	xsum int

	// The font that should be used.
	face font.Face

	// A map with information about the block, see the `Block` type.
	block *sync.Map

	// A channel where the block should be send to to once its ready to be
	// redrawn.
	redraw chan *Block

	// A channel where a boolean should be send once a block has initizalized,
	// notifying that the next block can intialize.
	ready chan bool
}

func initBar(x, y, w, h int, fp string) (*Bar, error) {
	bar := new(Bar)
	var err error

	// Set up a connection to the X server.
	bar.xu, err = xgbutil.NewConn()
	if err != nil {
		return nil, err
	}

	// Run the main X event loop, this is used to catch events.
	go xevent.Main(bar.xu)

	// Listen to the root window for property change events, used to check if
	// the user changed the focused window or active workspace for example.
	if err := xwindow.New(bar.xu, bar.xu.RootWin()).Listen(xproto.
		EventMaskPropertyChange); err != nil {
		return nil, err
	}

	// Create a window for the bar. This window listens to button press events
	// in order to respond to them.
	bar.win, err = xwindow.Generate(bar.xu)
	if err != nil {
		return nil, err
	}
	bar.win.Create(bar.xu.RootWin(), x, y, w, h, xproto.CwBackPixel|xproto.
		CwEventMask, 0x000000, xproto.EventMaskButtonPress)

	// EWMH stuff to make the window behave like an actual bar.
	// XXX: `WmStateSet` and `WmDesktopSet` are basically here to keep OpenBox
	// happy, can I somehow remove them and just use `_NET_WM_WINDOW_TYPE_DOCK`
	// like I can with WindowChef?
	if err := ewmh.WmWindowTypeSet(bar.xu, bar.win.Id, []string{
		"_NET_WM_WINDOW_TYPE_DOCK"}); err != nil {
		return nil, err
	}
	if err := ewmh.WmStateSet(bar.xu, bar.win.Id, []string{
		"_NET_WM_STATE_STICKY"}); err != nil {
		return nil, err
	}
	if err := ewmh.WmDesktopSet(bar.xu, bar.win.Id, ^uint(0)); err != nil {
		return nil, err
	}
	if err := ewmh.WmNameSet(bar.xu, bar.win.Id, "melonbar"); err != nil {
		return nil, err
	}

	// Map window.
	bar.win.Map()

	// XXX: Moving the window is again a hack to keep OpenBox happy.
	bar.win.Move(x, y)

	// Create the bar image.
	bar.img = xgraphics.New(bar.xu, image.Rect(0, 0, w, h))
	if err := bar.img.XSurfaceSet(bar.win.Id); err != nil {
		return nil, err
	}
	bar.img.XDraw()

	bar.w = w
	bar.h = h

	// Load font.
	fr := func(name string) ([]byte, error) {
		return ioutil.ReadFile(path.Join(path.Dir(fp), name))
	}
	fd, err := fr(path.Base(fp))
	if err != nil {
		return nil, err
	}
	bar.face, err = plan9font.ParseFont(fd, fr)
	if err != nil {
		return nil, err
	}

	bar.block = new(sync.Map)
	bar.redraw = make(chan *Block)

	// Listen to mouse events and execute the required function.
	xevent.ButtonPressFun(func(_ *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
		// Determine what block the cursor is in.
		// TODO: This feels a bit slow at the moment, can I improve it?
		var block *Block
		bar.block.Range(func(name, i interface{}) bool {
			block = i.(*Block)
			// XXX: Hack for music block.
			if name == "music" {
				tw := font.MeasureString(bar.face, block.txt).Ceil()
				if ev.EventX > int16(block.x+(block.w-tw+(block.xoff*2))) && ev.
					EventX < int16(block.x+block.w) {
					return false
				}
				return true
			}

			if ev.EventX > int16(block.x) && ev.EventX < int16(block.x+block.
				w) {
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
	d := &font.Drawer{
		Dst:  block.img,
		Src:  image.NewUniform(hexToBGRA(block.fg)),
		Face: bar.face,
	}

	// Calculate the required x coordinate for the different aligments.
	var x int
	tw := d.MeasureString(block.txt).Ceil()
	switch block.align {
	case 'l':
		x = block.x
	case 'c':
		x = block.x + ((block.w / 2) - (tw / 2))
	case 'r':
		x = (block.x + block.w) - tw
	case 'a':
		x = (bar.w / 2) - (tw / 2)
	default:
		return fmt.Errorf("draw %#U: Not a valid aligment rune", block.align)
	}
	x += block.xoff

	// Color the background.
	block.img.For(func(cx, cy int) xgraphics.BGRA {
		// XXX: Hack for music block.
		if block.w == 660 {
			if cx < x+block.xoff {
				return hexToBGRA("#445967")
			}
			return hexToBGRA(block.bg)
		}

		return hexToBGRA(block.bg)
	})

	// Draw the text.
	d.Dot = fixed.P(x, 18)
	d.DrawString(block.txt)

	// Redraw the bar.
	block.img.XDraw()
	bar.img.XPaint(bar.win.Id)

	return nil
}

func (bar *Bar) initBlocks(blocks []func()) {
	bar.ready = make(chan bool)

	for _, f := range blocks {
		go f()
		<-bar.ready
	}

	close(bar.ready)
}

func (bar *Bar) listen() {
	for {
		if err := bar.draw(<-bar.redraw); err != nil {
			panic(err)
		}
	}
}
