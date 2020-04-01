package main

import (
	"fmt"
	"image"
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/elliotchance/orderedmap"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Bar is a struct with information about the bar.
type Bar struct {
	// Bar window, and bar image.
	win *xwindow.Window
	img *xgraphics.Image

	// The width and height of the bar.
	w, h int

	// This is a sum of all of the block widths, used to draw a block to the
	// right of the last block.
	xsum int

	// Text drawer.
	drawer *font.Drawer

	// A map that stores the various blocks.
	blocks *orderedmap.OrderedMap

	// A map that stores the various popups.
	popups *orderedmap.OrderedMap

	// Store is an interface to store variables and objects to be used by other
	// blocks or popups.
	store map[string]interface{}

	// A channel where the block should be send to to once its ready to be
	// redrawn.
	redraw chan *Block
}

func initBar(x, y, w, h int) (*Bar, error) {
	bar := new(Bar)
	var err error

	// Create a window for the bar. This window listens to button press events
	// in order to respond to them.
	bar.win, err = xwindow.Generate(X)
	if err != nil {
		return nil, err
	}
	bar.win.Create(X.RootWin(), x, y, w, h, xproto.CwBackPixel|xproto.
		CwEventMask, 0x000000, xproto.EventMaskButtonPress)

	// EWMH stuff to make the window behave like an actual bar.
	if err := initEWMH(bar.win.Id); err != nil {
		return nil, err
	}

	// Map window.
	bar.win.Map()

	// XXX: Moving the window is again a hack to keep OpenBox happy.
	bar.win.Move(x, y)

	// Create the bar image.
	bar.img = xgraphics.New(X, image.Rect(0, 0, w, h))
	if err := bar.img.XSurfaceSet(bar.win.Id); err != nil {
		return nil, err
	}
	bar.img.XDraw()

	// Set bar width and height.
	bar.w = w
	bar.h = h

	// Set bar font face.
	bar.drawer = &font.Drawer{
		Dst:  bar.img,
		Face: face,
	}

	// Creat blocks and popups map.
	bar.blocks = orderedmap.NewOrderedMap()
	bar.popups = orderedmap.NewOrderedMap()

	// Create store map.
	bar.store = make(map[string]interface{})

	// Create redraw channel.
	bar.redraw = make(chan *Block)

	return bar, nil
}

func (bar *Bar) draw(block *Block) error {
	// Calculate the required x coordinate for the different aligments.
	var x int
	tw := bar.drawer.MeasureString(block.txt).Round()
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
	x += 2

	// Color the background.
	block.img.For(func(cx, cy int) xgraphics.BGRA {
		// XXX: Hack for music block.
		if block.w == 660 {
			if cx < x+block.xoff {
				return xgraphics.BGRA{B: 103, G: 89, R: 68, A: 0xFF}
			}
		}

		return block.bg
	})

	// Set foreground color.
	bar.drawer.Src = image.NewUniform(block.fg)

	// Draw the text.
	bar.drawer.Dot = fixed.P(x, 16)
	bar.drawer.DrawString(block.txt)

	// Redraw the bar.
	block.img.XDraw()
	bar.img.XPaint(bar.win.Id)

	return nil
}

func (bar *Bar) listen() {
	for {
		if err := bar.draw(<-bar.redraw); err != nil {
			log.Fatalln(err)
		}
	}
}
